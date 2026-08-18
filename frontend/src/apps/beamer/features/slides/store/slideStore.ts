import { z } from "zod";
import { slideSchema, type Slide } from "../types/slide.schema";
import { createLogger } from "@core/logger/logger";
import { sortByPriorityDesc } from "../utils/slideshow.logic";

const logger = createLogger("Stream");

type SlideDictionary = Record<number, Slide>;
type Listener = () => void;

export type ConnectionStatus = "connecting" | "connected" | "disconnected";

/**
 * SlideStore (Framework-Agnostic State Manager)
 *
 * WHAT IT DOES:
 * - Manages the Server-Sent Events (SSE) connection to receive real-time slide updates.
 * - Validates incoming data safely using Zod schemas.
 * - Maintains a single source of truth for all slides in a normalized dictionary (SlideDictionary).
 * - Implements a simple Observer pattern to notify subscribers (e.g., React hooks) of state changes.
 *
 * WHAT IT DOES NOT DO:
 * - It has ZERO dependencies on React (no hooks, no component re-renders).
 * - It does not handle UI logic, animations, or slideshow playback timing.
 */
export class SlideStore {
  private slideMap: SlideDictionary = {};

  private slideArray: Slide[] = [];

  // Two independent subscriber sets so that status changes don't trigger
  // expensive slide-array recomputation for data subscribers, and vice
  // versa. Keeps the Observer pattern clean: status changes are conceptually
  // orthogonal to slide data.
  private readonly dataListeners: Set<Listener> = new Set();
  private readonly statusListeners: Set<Listener> = new Set();
  private evtSource: EventSource | null = null;

  private status: ConnectionStatus = "disconnected";

  private reconnectTimeout: number | null = null;
  private watchdogTimeout: number | null = null;
  // NOTE: This interval must be at least 2x the server's ping interval (currently 10s).
  // The margin accounts for network jitter, latency, and browser background throttling.
  private readonly WATCHDOG_INTERVAL = 20000;

  // Exponential-backoff parameters (with ±20% jitter) for reconnect attempts
  // when the hub is unreachable. Reset to 0 on every successful message receipt.
  private reconnectAttempts = 0;
  private readonly RECONNECT_BASE_MS = 1000;
  private readonly RECONNECT_CAP_MS = 30000;

  // --- 1. Reactivity (Observer Pattern) ---

  /**
   * Subscribes a listener function to slide-data changes (INIT/CREATE/UPDATE/DELETE).
   * Status changes do NOT fire this listener.
   * @param listener - Function to be called whenever the slide state updates.
   * @returns A cleanup function to unsubscribe.
   */
  public subscribe = (listener: Listener): (() => void) => {
    this.dataListeners.add(listener);
    return () => this.dataListeners.delete(listener);
  };

  /**
   * Subscribes a listener function to connection-status changes.
   * Slide-data changes do NOT fire this listener.
   */
  public subscribeStatus = (listener: Listener): (() => void) => {
    this.statusListeners.add(listener);
    return () => this.statusListeners.delete(listener);
  };

  /**
   * Notifies data subscribers about a slide mutation and rebuilds the cached array.
   */
  private notifyData() {
    this.slideArray = Object.values(this.slideMap);
    this.dataListeners.forEach((listener) => listener());
  }

  /**
   * Notifies status subscribers about a connection-status change.
   */
  private notifyStatus() {
    this.statusListeners.forEach((listener) => listener());
  }

  // --- 2. State Access (Queries) ---

  public getSlides(): Slide[] {
    return this.slideArray;
  }

  public getById(id: number): Slide | null {
    return this.slideMap[id] || null;
  }

  public getByType(type: string): Slide[] {
    return this.getSlides()
      .filter((s) => s.content.type === type && s.status === "active")
      .sort(sortByPriorityDesc);
  }

  public getUrgent(): Slide[] {
    return this.getSlides()
      .filter(
        (s) =>
          s.content.type === "news" &&
          s.display_options.is_urgent &&
          s.status === "active",
      )
      .sort(sortByPriorityDesc);
  }

  // --- 3. State Mutations ---

  private initSlides(slides: Slide[]) {
    const newMap: SlideDictionary = {};
    slides.forEach((s) => (newMap[s.id] = s));
    this.slideMap = newMap;
    this.notifyData();
  }

  private upsertSlide(slide: Slide) {
    this.slideMap = { ...this.slideMap, [slide.id]: slide };
    this.notifyData();
  }

  private deleteSlide(id: number) {
    const newMap = { ...this.slideMap };
    delete newMap[id];
    this.slideMap = newMap;
    this.notifyData();
  }

  // --- Helper: Safe Parse & Validate ---
  private parseAndValidate<T>(
    eventName: string,
    rawData: string,
    schema: z.ZodType<T>,
  ): T | null {
    try {
      const parsedJson = JSON.parse(rawData);
      const validation = schema.safeParse(parsedJson);

      if (validation.success) {
        return validation.data;
      } else {
        logger.warn(
          `Failed to validate ${eventName} event`,
          "error",
          validation.error,
        );
        return null;
      }
    } catch (error) {
      logger.error(
        `Failed to parse JSON in ${eventName} event`,
        "error",
        error,
      );
      return null;
    }
  }

  // --- 4. Network / SSE Connection ---

  /**
   * Opens the Server-Sent Events (SSE) stream.
   * Should ideally be called only once at the root level of the application.
   */
  public connect() {
    if (this.evtSource) return; // Prevent multiple active connections

    this.setStatus("connecting");

    this.evtSource = new EventSource("/api/stream");
    this.resetWatchdog();

    this.evtSource.addEventListener("open", () => {
      logger.info("SSE connection opened");
      this.setStatus("connected");
    });

    const handleEvent = (handler: (e: MessageEvent) => void) => {
      return (e: MessageEvent) => {
        this.resetWatchdog(); // on every message: connection is alive
        this.reconnectAttempts = 0; // stable: reset exponential backoff
        handler(e);
      };
    };

    this.evtSource.addEventListener(
      "PING",
      handleEvent(() => {
        // noop
        logger.debug("PING event");
      }),
    );

    this.evtSource.addEventListener(
      "INIT",
      handleEvent((e: MessageEvent) => {
        const data = this.parseAndValidate(
          "INIT",
          e.data,
          z.array(slideSchema),
        );
        if (data) {
          logger.debug("Received INIT event", "count", data.length);
          this.initSlides(data);
        }
      }),
    );

    this.evtSource.addEventListener(
      "CREATE",
      handleEvent((e: MessageEvent) => {
        const data = this.parseAndValidate("CREATE", e.data, slideSchema);
        if (data) {
          logger.debug("Received CREATE event", "id", data.id);
          this.upsertSlide(data);
        }
      }),
    );

    this.evtSource.addEventListener(
      "UPDATE",
      handleEvent((e: MessageEvent) => {
        const data = this.parseAndValidate("UPDATE", e.data, slideSchema);
        if (data) {
          logger.debug("Received UPDATE event", "id", data.id);
          this.upsertSlide(data);
        }
      }),
    );

    this.evtSource.addEventListener(
      "DELETE",
      handleEvent((e: MessageEvent) => {
        const data = this.parseAndValidate("DELETE", e.data, z.number());
        if (data) {
          logger.debug("Received DELETE event", "id", data);
          this.deleteSlide(data);
        }
      }),
    );

    // We own reconnection on every `error` event (replacing the browser's
    // native ~3s auto-reconnect and any Last-Event-ID resumption). The
    // watchdog still drives forceReconnect after WATCHDOG_INTERVAL of
    // silence, but we flip the visible status immediately on error so the
    // overlay reflects reality instead of lying green for up to 20s.
    this.evtSource.addEventListener("error", () => {
      logger.warn("SSE error triggered (network or CORS issue).");

      const wasOpen = this.evtSource?.readyState === EventSource.OPEN;
      if (wasOpen) {
        this.setStatus("connecting");
      }

      this.forceReconnect();
    });
  }

  /**
   * Closes the SSE stream and cleans up the connection.
   */
  public disconnect() {
    if (this.evtSource) {
      this.evtSource.close();
      this.evtSource = null;
    }
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    if (this.watchdogTimeout) {
      window.clearTimeout(this.watchdogTimeout);
      this.watchdogTimeout = null;
    }
    this.setStatus("disconnected");
  }

  private resetWatchdog() {
    if (this.watchdogTimeout) {
      window.clearTimeout(this.watchdogTimeout);
    }

    this.watchdogTimeout = window.setTimeout(() => {
      logger.warn(
        "SSE Watchdog timeout: No messages received. Reconnecting...",
      );
      this.forceReconnect();
    }, this.WATCHDOG_INTERVAL);
  }

  private forceReconnect() {
    this.disconnect();

    this.setStatus("connecting");

    const exp = Math.min(
      this.RECONNECT_BASE_MS * 2 ** this.reconnectAttempts,
      this.RECONNECT_CAP_MS,
    );
    const jitter = exp * (0.8 + Math.random() * 0.4);
    this.reconnectAttempts += 1;

    this.reconnectTimeout = window.setTimeout(() => {
      logger.info(
        "Attempting to reconnect SSE...",
        "attempt",
        this.reconnectAttempts,
        "delay_ms",
        Math.round(jitter),
      );
      this.connect();
    }, jitter);
  }

  public getStatus = (): ConnectionStatus => {
    return this.status;
  };

  private setStatus(newStatus: ConnectionStatus) {
    if (this.status !== newStatus) {
      this.status = newStatus;
      this.notifyStatus();
    }
  }
}

/**
 * Singleton instance of the SlideStore.
 * Ensures the entire application shares the same SSE connection and state.
 */
export const slideStore = new SlideStore();
