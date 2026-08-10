import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  engineReducer,
  initialEngineState,
  EngineState,
} from "./engineReducer";
import type { Slide } from "../types/slide.schema";
import type { Playlist } from "@core/config/config.schemas";
import * as logic from "../utils/slideshow.logic";

// 1. Wir mocken die komplexe Business-Logik, da wir hier nur die State-Machine testen wollen
vi.mock("../utils/slideshow.logic", () => ({
  pickWeightedSlide: vi.fn(),
  sortSlides: vi.fn(),
  selectNextSlide: vi.fn(),
  findNextValidStep: vi.fn(),
}));

describe("engineReducer", () => {
  let baseState: EngineState;

  beforeEach(() => {
    // Vor jedem Test einen sauberen Zustand herstellen
    baseState = { ...initialEngineState };
    vi.clearAllMocks();
  });

  describe("Simple Actions", () => {
    it("should toggle the pause state", () => {
      const state1 = engineReducer(baseState, { type: "TOGGLE_PAUSE" });
      expect(state1.isPaused).toBe(true);

      const state2 = engineReducer(state1, { type: "TOGGLE_PAUSE" });
      expect(state2.isPaused).toBe(false);
    });

    it("should reset the playlist pointers", () => {
      const dirtyState: EngineState = {
        ...baseState,
        stepIndex: 5,
        stepCountPointer: 2,
        displayedStepInfo: { stepIndex: 5, stepCountPointer: 2 },
      };

      const newState = engineReducer(dirtyState, { type: "RESET_PLAYLIST" });

      expect(newState.stepIndex).toBe(0);
      expect(newState.stepCountPointer).toBe(0);
      expect(newState.displayedStepInfo).toBeNull();
    });

    it("should navigate to the PREVIOUS slide in history", () => {
      const stateWithHistory: EngineState = {
        ...baseState,
        history: [100, 101, 102],
        historyPointer: 2, // Wir sind aktuell beim letzten Slide (102)
      };

      const newState = engineReducer(stateWithHistory, { type: "PREVIOUS" });

      // Der Pointer sollte eins zurückgehen
      expect(newState.historyPointer).toBe(1);
      // Die History selbst darf sich nicht verändern
      expect(newState.history).toEqual([100, 101, 102]);
    });

    it("should NOT navigate PREVIOUS if at the beginning of history", () => {
      const stateWithHistory: EngineState = {
        ...baseState,
        history: [100, 101, 102],
        historyPointer: 0, // Wir sind schon ganz vorne
      };

      const newState = engineReducer(stateWithHistory, { type: "PREVIOUS" });
      expect(newState.historyPointer).toBe(0); // Bleibt bei 0
    });
  });

  describe("NEXT Action", () => {
    const defaultPayload = {
      hasUrgent: false,
      urgentSlides: [],
      activePlaylist: { id: 1, name: "Test", steps: [] } as Playlist,
      getByType: vi.fn(),
    };

    it("should do nothing if paused", () => {
      const pausedState = { ...baseState, isPaused: true };
      const newState = engineReducer(pausedState, {
        type: "NEXT",
        payload: defaultPayload,
      });

      expect(newState).toEqual(pausedState);
    });

    it("should navigate forward in history if we are currently looking at a past slide", () => {
      const stateWithHistory: EngineState = {
        ...baseState,
        history: [100, 101, 102],
        historyPointer: 0, // Wir sind beim ersten Slide, haben aber schon 3 gesehen
      };

      const newState = engineReducer(stateWithHistory, {
        type: "NEXT",
        payload: defaultPayload,
      });

      // Pointer geht einfach eins vor, keine neue Slide-Berechnung nötig
      expect(newState.historyPointer).toBe(1);
      expect(logic.findNextValidStep).not.toHaveBeenCalled();
    });

    it("should pick an urgent slide if hasUrgent is true", () => {
      const urgentSlide = { id: 999 } as Slide;
      // Wir sagen unserer Mock-Funktion, was sie zurückgeben soll
      vi.mocked(logic.pickWeightedSlide).mockReturnValue(urgentSlide);

      const newState = engineReducer(baseState, {
        type: "NEXT",
        payload: {
          ...defaultPayload,
          hasUrgent: true,
          urgentSlides: [urgentSlide],
        },
      });

      expect(logic.pickWeightedSlide).toHaveBeenCalled();
      expect(newState.history).toContain(999);
      expect(newState.historyPointer).toBe(0); // Erstes Element in der neuen History
    });
  });
});
