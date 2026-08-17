import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { SlideStore } from "./slideStore";
import type { Slide } from "../types/slide.schema";

// 1. Mocking the EventSource API
interface MessageEvent {
  data: string;
}

type EventCallback = (event: MessageEvent) => void;

class MockEventSource {
  public static instances: MockEventSource[] = [];
  public static readonly CONNECTING = 0;
  public static readonly OPEN = 1;
  public static readonly CLOSED = 2;
  public url: string;
  public readyState: number = MockEventSource.OPEN;
  public listeners: Record<string, EventCallback[]> = {};
  public onerror: (() => void) | null = null;

  constructor(url: string) {
    this.url = url;
    MockEventSource.instances.push(this);
  }

  addEventListener(eventName: string, callback: EventCallback) {
    if (!this.listeners[eventName]) {
      this.listeners[eventName] = [];
    }
    this.listeners[eventName].push(callback);
  }

  close() {
    this.readyState = MockEventSource.CLOSED;
    this.listeners = {};
  }

  // Helper to simulate server-sent events in tests
  emit(eventName: string, data: unknown) {
    const callbacks = this.listeners[eventName] || [];
    callbacks.forEach((cb) => cb({ data: JSON.stringify(data) }));
  }

  // Helper to push pre-serialized strings (e.g. malformed JSON).
  emitRaw(eventName: string, rawData: string) {
    const callbacks = this.listeners[eventName] || [];
    callbacks.forEach((cb) => cb({ data: rawData }));
  }

  // Helper to simulate an error event from the EventSource.
  triggerError() {
    this.readyState = MockEventSource.CLOSED;
    this.onerror?.();
  }
}

// Make EventSource available globally in the test environment
globalThis.EventSource = MockEventSource as unknown as typeof EventSource;

// 2. Dummy slides for tests
const mockSlide1: Slide = {
  id: 1,
  status: "active",
  created_at: "2024-01-01T00:00:00+00:00",
  content: { type: "image" },
  display_options: {
    priority: 10,
    is_urgent: false,
    allow_social_overlay: false,
  },
} as Slide;

const mockUrgentSlideLowPrio: Slide = {
  id: 2,
  status: "active",
  created_at: "2024-01-01T00:00:00+00:00",
  content: { type: "news", title: "Wichtig" },
  display_options: {
    priority: 5,
    is_urgent: true,
    allow_social_overlay: false,
  },
} as Slide;

const mockUrgentSlideHighPrio: Slide = {
  id: 3,
  status: "active",
  created_at: "2024-01-01T00:00:00+00:00",
  content: { type: "news", title: "Sehr Wichtig!" },
  display_options: {
    priority: 99,
    is_urgent: true,
    allow_social_overlay: false,
  },
} as Slide;

// 3. Actual tests
describe("SlideStore", () => {
  let store: SlideStore;

  beforeEach(() => {
    // Create a fresh store before each test
    store = new SlideStore();
    MockEventSource.instances = [];
  });

  afterEach(() => {
    store.disconnect();
    vi.restoreAllMocks();
  });

  it("should initialize with an empty state", () => {
    expect(store.getSlides()).toEqual([]);
  });

  it("should handle INIT event and notify subscribers", () => {
    const subscriber = vi.fn();
    store.subscribe(subscriber);
    store.connect();

    // DE05 3705 0198 0011 1026 54
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    // Simulate an INIT event from the server
    mockSSE.emit("INIT", [mockSlide1, mockUrgentSlideLowPrio]);

    expect(store.getSlides()).toHaveLength(2);
    expect(store.getById(1)).toEqual(mockSlide1);
    expect(subscriber).toHaveBeenCalledTimes(1); // Observer pattern works
  });

  it("should handle CREATE event to add a new slide", () => {
    store.connect();
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    mockSSE.emit("CREATE", mockSlide1);

    expect(store.getSlides()).toHaveLength(1);
    expect(store.getById(1)).toEqual(mockSlide1);
  });

  it("should handle UPDATE event to modify an existing slide", () => {
    store.connect();
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    // Initialize first
    mockSSE.emit("INIT", [mockSlide1]);

    // Then update
    const updatedSlide = {
      ...mockSlide1,
      status: "inactive",
      created_at: "2024-01-01T00:00:00+00:00",
    };
    mockSSE.emit("UPDATE", updatedSlide);

    expect(store.getById(1)?.status).toBe("inactive");
  });

  it("should handle DELETE event to remove a slide", () => {
    store.connect();
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    mockSSE.emit("INIT", [mockSlide1]);
    expect(store.getSlides()).toHaveLength(1);

    // Event only fires the ID
    mockSSE.emit("DELETE", 1);

    expect(store.getSlides()).toHaveLength(0);
    expect(store.getById(1)).toBeNull();
  });

  it("should correctly sort and filter urgent slides (getUrgent)", () => {
    store.connect();
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    // Load 3 slides: a regular image and two urgent news items with different priorities
    mockSSE.emit("INIT", [
      mockSlide1,
      mockUrgentSlideLowPrio,
      mockUrgentSlideHighPrio,
    ]);

    const urgentSlides = store.getUrgent();

    // Should have filtered out the regular image (mockSlide1)
    expect(urgentSlides).toHaveLength(2);

    // Should be sorted by priority descending (high priority first)
    expect(urgentSlides[0].id).toBe(3); // id: 3 has priority 99
    expect(urgentSlides[1].id).toBe(2); // id: 2 has priority 5
  });

  it("should correctly unsubscribe listeners", () => {
    const subscriber = vi.fn();
    const unsubscribe = store.subscribe(subscriber);

    store.connect();
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    // Trigger an event -> subscriber should be called
    mockSSE.emit("CREATE", mockSlide1);
    expect(subscriber).toHaveBeenCalledTimes(1);

    // Unsubscribe
    unsubscribe();

    // Trigger another event -> subscriber must NOT be called again
    mockSSE.emit("CREATE", mockUrgentSlideLowPrio);
    expect(subscriber).toHaveBeenCalledTimes(1);
  });

  it("should rebuild the cached slides array on every notify", () => {
    store.connect();
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    mockSSE.emit("INIT", [mockSlide1]);
    const before = store.getSlides();

    mockSSE.emit("UPDATE", mockSlide1);
    const after = store.getSlides();

    // `notify()` rebuilds the array each time, so references differ.
    // The cache exists so callers can rely on the array identity between
    // notifications.
    expect(before).not.toBe(after);
    expect(after).toHaveLength(1);
  });

  describe("SSE safety", () => {
    it("should not throw when receiving malformed JSON", () => {
      store.connect();
      const mockSSE = (store as unknown as { evtSource: MockEventSource })
        .evtSource;

      expect(() => mockSSE.emitRaw("CREATE", "not json at all{")).not.toThrow();
      // Nothing was stored
      expect(store.getSlides()).toEqual([]);
      // And no subscriber was notified either
      expect(store.getById(1)).toBeNull();
    });

    it("should not throw when receiving JSON that fails Zod validation", () => {
      store.connect();
      const mockSSE = (store as unknown as { evtSource: MockEventSource })
        .evtSource;

      // Valid JSON but missing required schema fields
      expect(() =>
        mockSSE.emitRaw("CREATE", JSON.stringify({ id: 1, wrong: "shape" })),
      ).not.toThrow();
      expect(store.getSlides()).toEqual([]);
    });

    it("should still process valid events after a malformed one", () => {
      store.connect();
      const mockSSE = (store as unknown as { evtSource: MockEventSource })
        .evtSource;

      mockSSE.emitRaw("CREATE", "<<<invalid>>>");
      mockSSE.emit("CREATE", mockSlide1);

      expect(store.getSlides()).toHaveLength(1);
      expect(store.getById(1)).toEqual(mockSlide1);
    });
  });

  describe("SSE reconnect", () => {
    it("should schedule a reconnect after onerror with CLOSED readyState", () => {
      vi.useFakeTimers();
      try {
        store.connect();
        const firstSSE = (store as unknown as { evtSource: MockEventSource })
          .evtSource;
        expect(MockEventSource.instances).toHaveLength(1);

        // Simulate a server drop
        firstSSE.triggerError();
        expect(firstSSE.readyState).toBe(MockEventSource.CLOSED);

        // 5s timer not yet fired
        vi.advanceTimersByTime(4999);
        expect(MockEventSource.instances).toHaveLength(1);

        // 5s timer fires -> a fresh EventSource is constructed via connect()
        vi.advanceTimersByTime(1);
        expect(MockEventSource.instances).toHaveLength(2);
        const secondSSE = MockEventSource.instances[1];
        expect(secondSSE).not.toBe(firstSSE);
        expect(secondSSE.url).toBe(firstSSE.url);
      } finally {
        vi.useRealTimers();
      }
    });

    it("should clear the pending reconnect timer when disconnect() is called", () => {
      vi.useFakeTimers();
      try {
        store.connect();
        const mockSSE = (store as unknown as { evtSource: MockEventSource })
          .evtSource;

        mockSSE.triggerError();
        store.disconnect();

        // Even waiting far longer than 5s, the pending reconnect must NOT fire.
        vi.advanceTimersByTime(30_000);
        expect(MockEventSource.instances).toHaveLength(1);
      } finally {
        vi.useRealTimers();
      }
    });
  });
});
