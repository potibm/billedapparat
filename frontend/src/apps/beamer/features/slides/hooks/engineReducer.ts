import { Playlist, PlaylistStep } from "@core/config/config.schemas";
import { Slide } from "../types/slide.schema";
import {
  pickWeightedSlide,
  sortSlides,
  selectNextSlide,
  findNextValidStep,
} from "../utils/slideshow.logic";

const HISTORY_LIMIT = 50;

// Unser gebündelter State
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

// Der Reducer: Nimmt den alten State + Aktion -> gibt neuen State zurück (Pure Function!)
export const engineReducer = (state: EngineState, action: EngineAction): EngineState => {
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
      const { hasUrgent, urgentSlides, activePlaylist, getByType } = action.payload;

      // 1. Guards
      if (state.isPaused || !activePlaylist?.steps) return state;

      // 2. History Navigation (Wir sind in der Vergangenheit und gehen einen Schritt vor)
      if (state.historyPointer < state.history.length - 1) {
        return { ...state, historyPointer: state.historyPointer + 1 };
      }

      const currentlyShownId = state.history[state.historyPointer];

      // 3. Urgent Override
      if (hasUrgent) {
        const selected = pickWeightedSlide(urgentSlides);
        if (selected && selected.id !== currentlyShownId) {
          const newHistory = [...state.history, selected.id].slice(-HISTORY_LIMIT);
          return {
            ...state,
            history: newHistory,
            historyPointer: Math.min(state.historyPointer + 1, HISTORY_LIMIT - 1),
          };
        }
        return state;
      }

      // 4. Reguläre Playlist-Logik
      const result = findNextValidStep(activePlaylist.steps, state.stepIndex, getByType);
      
      if (!result) return state; // Keine Slides gefunden (Logging machen wir im Hook)

      const { step, index: foundIndex, candidates } = result;

      // Haben wir einen neuen Step gefunden? Dann Pointers anpassen.
      let currentStepIndex = state.stepIndex;
      let currentStepCountPointer = state.stepCountPointer;

      if (foundIndex !== state.stepIndex) {
        currentStepIndex = foundIndex;
        currentStepCountPointer = 0;
      }

      // 5. Slide auswählen
      const sorted = sortSlides(candidates, step.order);
      const selected = selectNextSlide(sorted, step, currentStepCountPointer, currentlyShownId);

      if (!selected) return state;

      // 6. Pointer hochzählen (advanceStepPointers Logik)
      let nextStepIndex = currentStepIndex;
      let nextCountPointer = currentStepCountPointer + 1;

      if (nextCountPointer >= step.count) {
        nextStepIndex = (currentStepIndex + 1) % activePlaylist.steps.length;
        nextCountPointer = 0;
      }

      // 7. Neuen State zurückgeben
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
