import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { SlideStore } from "./slideStore";
import type { Slide } from "../types/slide.schema";

// --- 1. Mocking der EventSource API ---
interface MessageEvent {
  data: string;
}

type EventCallback = (event: MessageEvent) => void;

class MockEventSource {
  public url: string;
  public listeners: Record<string, EventCallback[]> = {};

  constructor(url: string) {
    this.url = url;
  }

  addEventListener(eventName: string, callback: EventCallback) {
    if (!this.listeners[eventName]) {
      this.listeners[eventName] = [];
    }
    this.listeners[eventName].push(callback);
  }

  close() {
    this.listeners = {};
  }

  // Hilfsfunktion für unsere Tests, um Server-Events zu simulieren
  emit(eventName: string, data: unknown) {
    const callbacks = this.listeners[eventName] || [];
    callbacks.forEach((cb) => cb({ data: JSON.stringify(data) }));
  }
}

// EventSource global im Test-Environment verfügbar machen
globalThis.EventSource = MockEventSource as unknown as typeof EventSource;

// --- 2. Testdaten (Dummy Slides) ---
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

// --- 3. Die eigentlichen Tests ---
describe("SlideStore", () => {
  let store: SlideStore;

  beforeEach(() => {
    // Vor jedem Test einen frischen Store erstellen
    store = new SlideStore();
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

    // Hole die gemockte EventSource Instanz
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    // Simuliere ein INIT Event vom Server
    mockSSE.emit("INIT", [mockSlide1, mockUrgentSlideLowPrio]);

    expect(store.getSlides()).toHaveLength(2);
    expect(store.getById(1)).toEqual(mockSlide1);
    expect(subscriber).toHaveBeenCalledTimes(1); // Observer Pattern funktioniert!
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

    // Zuerst initialisieren
    mockSSE.emit("INIT", [mockSlide1]);

    // Dann updaten
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

    // Event feuert nur die ID
    mockSSE.emit("DELETE", 1);

    expect(store.getSlides()).toHaveLength(0);
    expect(store.getById(1)).toBeNull();
  });

  it("should correctly sort and filter urgent slides (getUrgent)", () => {
    store.connect();
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    // Wir laden 3 Slides: Ein normales Bild, und zwei Urgent-News mit unterschiedlicher Prio
    mockSSE.emit("INIT", [
      mockSlide1,
      mockUrgentSlideLowPrio,
      mockUrgentSlideHighPrio,
    ]);

    const urgentSlides = store.getUrgent();

    // Sollte das normale Bild (mockSlide1) herausgefiltert haben
    expect(urgentSlides).toHaveLength(2);

    // Sollte nach Priorität absteigend sortiert sein (HighPrio zuerst)
    expect(urgentSlides[0].id).toBe(3); // id: 3 hat Prio 99
    expect(urgentSlides[1].id).toBe(2); // id: 2 hat Prio 5
  });

  it("should correctly unsubscribe listeners", () => {
    const subscriber = vi.fn();
    const unsubscribe = store.subscribe(subscriber);

    store.connect();
    const mockSSE = (store as unknown as { evtSource: MockEventSource })
      .evtSource;

    // Ein Event auslösen -> Subscriber sollte aufgerufen werden
    mockSSE.emit("CREATE", mockSlide1);
    expect(subscriber).toHaveBeenCalledTimes(1);

    // Abmelden
    unsubscribe();

    // Noch ein Event auslösen -> Subscriber darf NICHT nochmal aufgerufen werden
    mockSSE.emit("CREATE", mockUrgentSlideLowPrio);
    expect(subscriber).toHaveBeenCalledTimes(1);
  });
});
