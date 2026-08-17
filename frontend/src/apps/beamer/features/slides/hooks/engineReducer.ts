import { Playlist } from "@core/config/config.schemas";
import { Slide } from "../types/slide.schema";
import {
  sortSlides,
  sortByPriorityDesc,
  selectNextSlide,
  findNextValidStep,
} from "../utils/slideshow.logic";

const HISTORY_LIMIT = 50;

// The bundled engine state
export interface EngineState {
  stepIndex: number;
  stepCountPointer: number;
  displayedStepInfo: { stepIndex: number; stepCountPointer: number } | null;
  history: number[];
  historyPointer: number;
  isPaused: boolean;
  /**
   * Counts consecutive NEXT actions that did not advance the engine
   * (because no candidates were available in the playlist). Reset on any
   * successful advance or explicit RESET_PLAYLIST. A watcher in
   * `useSlideshowEngine` dispatches RESET_PLAYLIST when this exceeds a
   * small threshold, so the Beamer can fall back to STANDBY instead of
   * remaining stuck on an inactive slide.
   */
  recoveryAttempts: number;
}

export const initialEngineState: EngineState = {
  stepIndex: 0,
  stepCountPointer: 0,
  displayedStepInfo: null,
  history: [],
  historyPointer: -1,
  isPaused: false,
  recoveryAttempts: 0,
};

// All allowed actions
export type EngineAction =
  | { type: "TOGGLE_PAUSE" }
  | { type: "PREVIOUS" }
  | { type: "RESET_PLAYLIST" }
  | {
      type: "NEXT";
      payload: {
        activePlaylist: Playlist | null;
        activeSlides: Slide[];
      };
    };

// The reducer: takes old state + action -> returns new state (pure function)
export const engineReducer = (
  state: EngineState,
  action: EngineAction,
): EngineState => {
  switch (action.type) {
    case "TOGGLE_PAUSE":
      return { ...state, isPaused: !state.isPaused, recoveryAttempts: 0 };

    case "PREVIOUS":
      if (state.historyPointer > 0) {
        return {
          ...state,
          historyPointer: state.historyPointer - 1,
          recoveryAttempts: 0,
        };
      }
      return state;

    case "RESET_PLAYLIST":
      return {
        ...state,
        stepIndex: 0,
        stepCountPointer: 0,
        displayedStepInfo: null,
        history: [],
        historyPointer: -1,
        recoveryAttempts: 0,
      };

    case "NEXT": {
      const { activePlaylist, activeSlides } = action.payload;

      // 1. Guards
      if (state.isPaused || !activePlaylist?.steps) return state;

      // 2. History navigation (we are in the past and move one step forward)
      if (state.historyPointer < state.history.length - 1) {
        return {
          ...state,
          historyPointer: state.historyPointer + 1,
          recoveryAttempts: 0,
        };
      }

      const currentlyShownId = state.history[state.historyPointer];

      const pureGetByType = (type: string) => {
        // for the virtual urgent playlist
        if (type === "urgent") {
          return activeSlides
            .filter(
              (s) =>
                s.display_options?.is_urgent === true && s.status === "active",
            )
            .sort(sortByPriorityDesc);
        }

        return activeSlides
          .filter((s) => s.content.type === type && s.status === "active")
          .sort(sortByPriorityDesc);
      };

      // 3. Regular playlist logic
      const result = findNextValidStep(
        activePlaylist.steps,
        state.stepIndex,
        pureGetByType,
      );

      if (!result) {
        // No valid step found; count this as a recovery attempt.
        return { ...state, recoveryAttempts: state.recoveryAttempts + 1 };
      }

      const { step, index: foundIndex, candidates } = result;

      // Found a new step? Then adjust pointers.
      let currentStepIndex = state.stepIndex;
      let currentStepCountPointer = state.stepCountPointer;

      if (foundIndex !== state.stepIndex) {
        currentStepIndex = foundIndex;
        currentStepCountPointer = 0;
      }

      // 5. Select slide
      const sorted = sortSlides(candidates, step.order);
      const selected = selectNextSlide(
        sorted,
        step,
        currentStepCountPointer,
        currentlyShownId,
      );

      if (!selected) {
        // Step exists but selection returned no candidate; count as recovery.
        return { ...state, recoveryAttempts: state.recoveryAttempts + 1 };
      }

      // 6. Increment pointers (advanceStepPointers logic)
      let nextStepIndex = currentStepIndex;
      let nextCountPointer = currentStepCountPointer + 1;

      if (nextCountPointer >= step.count) {
        nextStepIndex = (currentStepIndex + 1) % activePlaylist.steps.length;
        nextCountPointer = 0;
      }

      // 7. Return new state
      const newHistory = [...state.history, selected.id].slice(-HISTORY_LIMIT);

      return {
        ...state,
        stepIndex: nextStepIndex,
        stepCountPointer: nextCountPointer,
        displayedStepInfo: {
          stepIndex: foundIndex,
          stepCountPointer: currentStepCountPointer,
        },
        history: newHistory,
        historyPointer: Math.min(state.historyPointer + 1, HISTORY_LIMIT - 1),
        recoveryAttempts: 0,
      };
    }

    default:
      return state;
  }
};
