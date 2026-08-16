import { Playlist } from "@core/config/config.schemas";
import { Slide } from "../types/slide.schema";
import {
  pickWeightedSlide,
  sortSlides,
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
}

export const initialEngineState: EngineState = {
  stepIndex: 0,
  stepCountPointer: 0,
  displayedStepInfo: null,
  history: [],
  historyPointer: -1,
  isPaused: false,
};

// Alle erlaubten Aktionen
export type EngineAction =
  | { type: "TOGGLE_PAUSE" }
  | { type: "PREVIOUS" }
  | { type: "RESET_PLAYLIST" }
  | {
      type: "NEXT";
      payload: {
        hasUrgent: boolean;
        urgentSlides: Slide[];
        activePlaylist: Playlist | null;
        getByType: (type: string) => Slide[];
      };
    };

// The reducer: takes old state + action -> returns new state (pure function)
export const engineReducer = (
  state: EngineState,
  action: EngineAction,
): EngineState => {
  switch (action.type) {
    case "TOGGLE_PAUSE":
      return { ...state, isPaused: !state.isPaused };

    case "PREVIOUS":
      if (state.historyPointer > 0) {
        return { ...state, historyPointer: state.historyPointer - 1 };
      }
      return state;

    case "RESET_PLAYLIST":
      return {
        ...state,
        stepIndex: 0,
        stepCountPointer: 0,
        displayedStepInfo: null,
      };

    case "NEXT": {
      const { hasUrgent, urgentSlides, activePlaylist, getByType } =
        action.payload;

      // 1. Guards
      if (state.isPaused || !activePlaylist?.steps) return state;

      // 2. History navigation (we are in the past and move one step forward)
      if (state.historyPointer < state.history.length - 1) {
        return { ...state, historyPointer: state.historyPointer + 1 };
      }

      const currentlyShownId = state.history[state.historyPointer];

      // 3. Urgent Override
      if (hasUrgent) {
        const selected = pickWeightedSlide(urgentSlides);
        if (selected && selected.id !== currentlyShownId) {
          const newHistory = [...state.history, selected.id].slice(
            -HISTORY_LIMIT,
          );
          return {
            ...state,
            history: newHistory,
            historyPointer: Math.min(
              state.historyPointer + 1,
              HISTORY_LIMIT - 1,
            ),
          };
        }
        return state;
      }

      // 4. Regular playlist logic
      const result = findNextValidStep(
        activePlaylist.steps,
        state.stepIndex,
        getByType,
      );

      if (!result) return state; // No slides found (logging is handled in the hook)

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

      if (!selected) return state;

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
      };
    }

    default:
      return state;
  }
};
