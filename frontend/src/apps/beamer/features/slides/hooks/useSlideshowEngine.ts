import { useEffect, useMemo, useReducer, useCallback } from "react";
import { useSlideManager } from "./useSlideManager";
import { createLogger } from "@core/logger/logger";
import { useCurrentPlaylist } from "./useCurrentPlaylist";
import { PlaylistStep } from "@core/config/config.schemas";
import { SlideshowEngine } from "../types/slideshow.types";
import { engineReducer, initialEngineState } from "./engineReducer";

const NEXT_TICK_TIMEOUT = 0;
const logger = createLogger("Slideshow");

export const useSlideshowEngine = (): SlideshowEngine => {
  const { getByType, getUrgent, slides: allSlides } = useSlideManager();
  const activePlaylist = useCurrentPlaylist();

  // Unser neuer, zentraler State!
  const [state, dispatch] = useReducer(engineReducer, initialEngineState);

  const urgentSlides = getUrgent();
  const hasUrgent = urgentSlides.length > 0;
  const toastSlides = getByType("social.text");

  // --- Actions ---
  // Durch den Reducer brauchen diese Funktionen keine Dependency-Arrays
  // mehr, die sich ständig ändern.
  const next = useCallback(() => {
    dispatch({
      type: "NEXT",
      payload: { hasUrgent, urgentSlides, activePlaylist, getByType },
    });
  }, [hasUrgent, urgentSlides, activePlaylist, getByType]);

  const previous = useCallback(() => dispatch({ type: "PREVIOUS" }), []);
  const togglePause = useCallback(() => dispatch({ type: "TOGGLE_PAUSE" }), []);

  // --- Abgeleitete Daten (Derived State) ---
  const currentSlide = useMemo(() => {
    const id = state.history[state.historyPointer];
    return allSlides.find((s) => s.id === id) || null;
  }, [state.history, state.historyPointer, allSlides]);

  const currentStep = useMemo<PlaylistStep | undefined>(() => {
    if (!activePlaylist) return undefined;
    return activePlaylist.steps[state.stepIndex];
  }, [activePlaylist, state.stepIndex]);

  // --- Effekte (Timeouts und Kickstarts) ---
  useEffect(() => {
    if (state.history.length === 0 && allSlides.length > 0) {
      logger.debug("Kickstarting initial slide");
      const timeoutId = setTimeout(next, NEXT_TICK_TIMEOUT);
      return () => clearTimeout(timeoutId);
    }
  }, [state.history.length, allSlides.length, next]);

  useEffect(() => {
    if (hasUrgent && currentSlide?.display_options?.is_urgent !== true) {
      const timeoutId = setTimeout(next, NEXT_TICK_TIMEOUT);
      return () => clearTimeout(timeoutId);
    }
  }, [hasUrgent, currentSlide, next]);

  useEffect(() => {
    if (currentSlide && currentSlide.status !== "active") {
      const timeoutId = setTimeout(next, NEXT_TICK_TIMEOUT);
      return () => clearTimeout(timeoutId);
    }
  }, [currentSlide, next]);

  useEffect(() => {
    logger.info("Playlist changed, resetting engine pointers");
    const timeoutId = setTimeout(() => {
      dispatch({ type: "RESET_PLAYLIST" });
      next();
    }, 0);
    return () => clearTimeout(timeoutId);
    // eslint-disable-next-line @eslint-react/exhaustive-deps
  }, [activePlaylist?.id]);

  // --- Fallback, wenn keine Playlist da ist ---
  if (!activePlaylist) {
    return {
      currentSlide: null,
      next,
      previous,
      togglePause,
      isPaused: state.isPaused,
      isUrgent: false,
      toastSlides: [],
      duration: 10,
      stepInfo: null,
    };
  }

  // --- Return an die UI ---
  return {
    currentSlide,
    next,
    previous,
    togglePause,
    isPaused: state.isPaused,
    isUrgent: currentSlide?.display_options?.is_urgent === true,
    toastSlides,
    duration: currentStep?.duration || 10,
    stepInfo:
      state.displayedStepInfo &&
      state.displayedStepInfo.stepIndex < activePlaylist.steps.length
        ? {
            type: activePlaylist.steps[state.displayedStepInfo.stepIndex].type,
            current: state.displayedStepInfo.stepCountPointer + 1,
            total:
              activePlaylist.steps[state.displayedStepInfo.stepIndex].count,
            playlistName: activePlaylist.name,
          }
        : null,
  };
};
