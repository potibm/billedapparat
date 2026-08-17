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
vi.mock(import("../utils/slideshow.logic"), async (importOriginal) => {
  const actual = await importOriginal();
  return {
    pickWeightedSlide: vi.fn(),
    sortSlides: vi.fn(),
    selectNextSlide: vi.fn(),
    findNextValidStep: vi.fn(),
    // The reducer also calls sortByPriorityDesc inside its pureGetByType —
    // use the real one so the filter+sort logic is exercised end-to-end.
    sortByPriorityDesc: actual.sortByPriorityDesc,
  };
});

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
        isPaused: false,
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
    // The new payload is pure: just slides + playlist, no closures or flags.
    const defaultPayload = {
      activePlaylist: { id: 1, name: "Test", steps: [] } as Playlist,
      activeSlides: [] as Slide[],
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

    // The urgent test previously lived here but has been removed: the engine
    // no longer knows about urgent slides — that override is now implemented
    // as a virtual "urgent" playlist step produced by useCurrentPlaylist.

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

    it("should use activeSlides (passed in the payload) instead of an injected lookup", () => {
      // The reducer no longer accepts a getByType function in the payload.
      // We verify this by giving it a payload with only `activeSlides` and
      // asserting the call to findNextValidStep still receives a working
      // type-filter function (derived from activeSlides).
      const stateWithHistory: EngineState = {
        ...baseState,
        history: [1, 2, 3],
        historyPointer: 2,
      };

      const mockSlide = { id: 999 } as Slide;
      vi.mocked(logic.findNextValidStep).mockReturnValue({
        step: {
          type: "news",
          order: "asc",
          count: 1,
          duration: 10,
        },
        index: 0,
        candidates: [mockSlide],
      });
      vi.mocked(logic.sortSlides).mockReturnValue([mockSlide]);
      vi.mocked(logic.selectNextSlide).mockReturnValue(mockSlide);

      const newState = engineReducer(stateWithHistory, {
        type: "NEXT",
        payload: {
          activePlaylist: { id: 1, name: "Test", steps: [] } as Playlist,
          activeSlides: [mockSlide],
        },
      });

      expect(newState.history).toEqual([1, 2, 3, 999]);
      expect(newState.historyPointer).toBe(3);
    });

    it("should return state unchanged when findNextValidStep returns null", () => {
      const stateWithHistory: EngineState = {
        ...baseState,
        history: [1, 2, 3],
        historyPointer: 2,
      };

      vi.mocked(logic.findNextValidStep).mockReturnValue(null);

      const newState = engineReducer(stateWithHistory, {
        type: "NEXT",
        payload: {
          activePlaylist: {
            id: 1,
            name: "Test",
            steps: [{ type: "news", order: "asc", count: 1, duration: 10 }],
          } as Playlist,
          activeSlides: [],
        },
      });

      expect(newState).toEqual(stateWithHistory);
    });

    it("should map a virtual 'urgent' playlist step to activeSlides filtered by is_urgent", () => {
      // Document/regression-test for the architectural change:
      // urgent override is no longer a branch inside the reducer. Instead,
      // useCurrentPlaylist emits a synthetic playlist with a step of
      // type: "urgent", and the reducer's pureGetByType handles that
      // special-case by filtering activeSlides for is_urgent slides.
      //
      // We assert that the lookup function passed to findNextValidStep
      // is the pure reducer-internal function (not a mock) by inspecting
      // its behavior via the mocked findNextValidStep call.
      const urgentSlide = {
        id: 50,
        status: "active",
        display_options: {
          is_urgent: true,
          priority: 1,
          allow_social_overlay: false,
        },
        content: { type: "news" },
      } as Slide;
      const regularSlide = {
        id: 99,
        status: "active",
        display_options: {
          is_urgent: false,
          priority: 1,
          allow_social_overlay: false,
        },
        content: { type: "news" },
      } as Slide;

      // Mock findNextValidStep to capture and invoke its lookup callback
      let capturedLookup: ((type: string) => Slide[]) | null = null;
      vi.mocked(logic.findNextValidStep).mockImplementation(
        (_steps, _idx, lookup) => {
          capturedLookup = lookup;
          // Simulate the urgent step returning all urgent slides
          const urgents = lookup("urgent");
          return {
            step: { type: "urgent", order: "desc", count: 1, duration: 10 },
            index: 0,
            candidates: urgents,
          };
        },
      );
      vi.mocked(logic.sortSlides).mockImplementation((list) => list);
      vi.mocked(logic.selectNextSlide).mockReturnValue(urgentSlide);

      engineReducer(baseState, {
        type: "NEXT",
        payload: {
          activePlaylist: {
            id: -1,
            name: "Urgent Override",
            steps: [{ type: "urgent", order: "desc", count: 1, duration: 10 }],
          } as Playlist,
          activeSlides: [regularSlide, urgentSlide],
        },
      });

      expect(capturedLookup).not.toBeNull();
      // The lookup must filter by is_urgent=true and exclude non-urgent slides.
      const result = capturedLookup!("urgent");
      expect(result.map((s) => s.id)).toEqual([50]);
      // And the regular-looking branch filters by content.type.
      const regularResult = capturedLookup!("news");
      expect(regularResult.map((s) => s.id).sort()).toEqual([50, 99]);
    });
  });
});
