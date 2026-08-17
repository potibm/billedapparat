import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  engineReducer,
  initialEngineState,
  EngineState,
} from "./engineReducer";
import type { Slide } from "../types/slide.schema";
import type { Playlist } from "@core/config/config.schemas";
import * as logic from "../utils/slideshow.logic";

// 1. Mock the complex business logic because we only want to test the state machine here
vi.mock("../utils/slideshow.logic", () => ({
  pickWeightedSlide: vi.fn(),
  sortSlides: vi.fn(),
  selectNextSlide: vi.fn(),
  findNextValidStep: vi.fn(),
}));

describe("engineReducer", () => {
  let baseState: EngineState;

  beforeEach(() => {
    // Reset to a clean state before each test
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

    it("should reset the playlist pointers and clear history", () => {
      const dirtyState: EngineState = {
        ...baseState,
        stepIndex: 5,
        stepCountPointer: 2,
        displayedStepInfo: { stepIndex: 5, stepCountPointer: 2 },
        history: [100, 101, 102],
        historyPointer: 2,
      };

      const newState = engineReducer(dirtyState, { type: "RESET_PLAYLIST" });

      expect(newState.stepIndex).toBe(0);
      expect(newState.stepCountPointer).toBe(0);
      expect(newState.displayedStepInfo).toBeNull();
      expect(newState.history).toEqual([]);
      expect(newState.historyPointer).toBe(-1);
    });

    it("should navigate to the PREVIOUS slide in history", () => {
      const stateWithHistory: EngineState = {
        ...baseState,
        history: [100, 101, 102],
        historyPointer: 2, // We are currently at the last slide (102)
      };

      const newState = engineReducer(stateWithHistory, { type: "PREVIOUS" });

      // Pointer should move back by one
      expect(newState.historyPointer).toBe(1);
      // The history itself must not change
      expect(newState.history).toEqual([100, 101, 102]);
    });

    it("should NOT navigate PREVIOUS if at the beginning of history", () => {
      const stateWithHistory: EngineState = {
        ...baseState,
        history: [100, 101, 102],
        historyPointer: 0, // We are already at the beginning
      };

      const newState = engineReducer(stateWithHistory, { type: "PREVIOUS" });
      expect(newState.historyPointer).toBe(0); // Stays at 0
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
        historyPointer: 0, // We are at the first slide but have already seen 3
      };

      const newState = engineReducer(stateWithHistory, {
        type: "NEXT",
        payload: defaultPayload,
      });

      // Pointer simply moves forward, no new slide calculation needed
      expect(newState.historyPointer).toBe(1);
      expect(logic.findNextValidStep).not.toHaveBeenCalled();
    });

    it("should pick an urgent slide if hasUrgent is true", () => {
      const urgentSlide = { id: 999 } as Slide;
      // Tell our mock function what to return
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
      expect(newState.historyPointer).toBe(0); // First element in the new history
    });

    it("should NOT replay old history when NEXT follows RESET_PLAYLIST (regression)", () => {
      // State mid-history: pointer is not at the end, so the NEXT handler
      // would normally short-circuit into the history-replay branch.
      const stateWithHistory: EngineState = {
        ...baseState,
        history: [100, 101, 102],
        historyPointer: 1,
      };

      const mockFreshSlide = { id: 999 } as Slide;
      vi.mocked(logic.findNextValidStep).mockReturnValue({
        step: {
          type: "news",
          order: "asc",
          count: 1,
          duration: 10,
        },
        index: 0,
        candidates: [mockFreshSlide],
      });
      vi.mocked(logic.sortSlides).mockReturnValue([mockFreshSlide]);
      vi.mocked(logic.selectNextSlide).mockReturnValue(mockFreshSlide);

      const resetState = engineReducer(stateWithHistory, {
        type: "RESET_PLAYLIST",
      });
      const newState = engineReducer(resetState, {
        type: "NEXT",
        payload: defaultPayload,
      });

      // The reducer must have asked for a fresh slide — proof it did not
      // replay a stale slide from the previous playlist's history.
      expect(logic.findNextValidStep).toHaveBeenCalledTimes(1);
      expect(newState.history).toEqual([999]);
      expect(newState.historyPointer).toBe(0);
    });
  });
});
