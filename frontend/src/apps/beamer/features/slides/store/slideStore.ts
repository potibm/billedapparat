import { z } from "zod";
import { slideSchema, type Slide } from "../types/slide.schema";
import { createLogger } from "@core/logger/logger";

const logger = createLogger("Slides");

type SlideDictionary = Record<number, Slide>;
type Listener = () => void;

const sortByPriorityDesc = (a: Slide, b: Slide) =>
  Number(b.display_options?.priority || 0) -
  Number(a.display_options?.priority || 0);

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

  private readonly listeners: Set<Listener> = new Set();
  private evtSource: EventSource | null = null;

  private reconnectTimeout: number | null = null;

  // --- 1. Reactivity (Observer Pattern) ---

  /**
   * Subscribes a listener function to state changes.
   * @param listener - Function to be called whenever the slide state updates.
   * @returns A cleanup function to unsubscribe.
   */
  public subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  /**
   * Notifies all active subscribers about a state mutation.
   */
  private notify() {
    this.slideArray = Object.values(this.slideMap);
    this.listeners.forEach((listener) => listener());
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
    this.notify();
  }

  private upsertSlide(slide: Slide) {
    this.slideMap = { ...this.slideMap, [slide.id]: slide };
    this.notify();
  }

  private deleteSlide(id: number) {
    const newMap = { ...this.slideMap };
    delete newMap[id];
    this.slideMap = newMap;
    this.notify();
  }

  // --- Helper: Safe Parse & Validate ---
  private parseAndValidate<T>(
    eventName: string,
    rawData: string,
    schema: z.ZodType<T>,
  ): T | null {
    try {
      const parsedJson = JSON.parse(rawData); // <-- Try/Catch fängt das JSON-Problem
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

    this.evtSource = new EventSource("/api/stream");

    this.evtSource.addEventListener("INIT", (e: MessageEvent) => {
      const data = this.parseAndValidate("INIT", e.data, z.array(slideSchema));
      if (data) {
        logger.debug("Received INIT event", "count", data.length);
        this.initSlides(data);
      }
    });

    this.evtSource.addEventListener("CREATE", (e: MessageEvent) => {
      const data = this.parseAndValidate("CREATE", e.data, slideSchema);
      if (data) {
        logger.debug("Received CREATE event", "id", data.id);
        this.upsertSlide(data);
      }
    });

    this.evtSource.addEventListener("UPDATE", (e: MessageEvent) => {
      const data = this.parseAndValidate("UPDATE", e.data, slideSchema);
      if (data) {
        logger.debug("Received UPDATE event", "id", data.id);
        this.upsertSlide(data);
      }
    });

    this.evtSource.addEventListener("DELETE", (e: MessageEvent) => {
      const data = this.parseAndValidate("DELETE", e.data, z.number());
      if (data) {
        logger.debug("Received DELETE event", "id", data);
        this.deleteSlide(data);
      }
    });

    this.evtSource.onerror = () => {
      logger.warn("SSE connection lost. Attempting to reconnect...");

      if (this.evtSource?.readyState === EventSource.CLOSED) {
        this.disconnect();
        if (this.reconnectTimeout) clearTimeout(this.reconnectTimeout);
        this.reconnectTimeout = window.setTimeout(() => this.connect(), 5000);
      }
    };
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
  }
}

/**
 * Singleton instance of the SlideStore.
 * Ensures the entire application shares the same SSE connection and state.
 */
export const slideStore = new SlideStore();
