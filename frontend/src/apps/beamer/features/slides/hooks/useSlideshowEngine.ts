import { useEffect, useMemo, useReducer, useCallback } from "react";
import { useSlideManager } from "./useSlideManager";
import { createLogger } from "@core/logger/logger";
import { useCurrentPlaylist } from "./useCurrentPlaylist";
import { Playlist, PlaylistStep } from "@core/config/config.schemas";
import { SlideshowEngine } from "../types/slideshow.types";
import { engineReducer, initialEngineState } from "./engineReducer";

const NEXT_TICK_TIMEOUT = 0;
// After how many consecutive NEXT actions that did not advance the engine
// do we declare it stuck and fall back to STANDBY via RESET_PLAYLIST?
const STUCK_THRESHOLD = 3;
const logger = createLogger("Slideshow");

export const useSlideshowEngine = (): SlideshowEngine => {
  const { getByType, slides: allSlides } = useSlideManager();
  const activePlaylist = useCurrentPlaylist() as Playlist | null;

  const [state, dispatch] = useReducer(engineReducer, initialEngineState);

  const toastSlides = getByType("social.text");

  // --- Actions ---
  const next = useCallback(() => {
    dispatch({
      type: "NEXT",
      payload: { activePlaylist, activeSlides: allSlides },
    });
  }, [activePlaylist, allSlides]);

  const previous = useCallback(() => dispatch({ type: "PREVIOUS" }), []);
  const togglePause = useCallback(() => dispatch({ type: "TOGGLE_PAUSE" }), []);

  // --- Derived State ---
  const currentSlide = useMemo(() => {
    const id = state.history[state.historyPointer];
    return allSlides.find((s) => s.id === id) || null;
  }, [state.history, state.historyPointer, allSlides]);

  const currentStep = useMemo<PlaylistStep | undefined>(() => {
    if (!activePlaylist || !state.displayedStepInfo) return undefined;
    return activePlaylist.steps[state.displayedStepInfo.stepIndex];
  }, [activePlaylist, state.displayedStepInfo]);

  // --- Effects ---
  useEffect(() => {
    if (state.history.length === 0 && allSlides.length > 0) {
      logger.debug("Kickstarting initial slide");
      const timeoutId = setTimeout(next, NEXT_TICK_TIMEOUT);
      return () => clearTimeout(timeoutId);
    }
  }, [state.history.length, allSlides.length, next]);

  useEffect(() => {
    if (currentSlide && currentSlide.status !== "active") {
      const timeoutId = setTimeout(next, NEXT_TICK_TIMEOUT);
      return () => clearTimeout(timeoutId);
    }
  }, [currentSlide, next]);

  // Watchdog: if the engine has failed to advance N consecutive times AND
  // has nothing playable to show, fall back to STANDBY by resetting the
  // playlist. Without this, `currentSlide === null` (e.g. backend deleted
  // the displayed slide) or a stuck history pointer left the user staring
  // at a black screen indefinitely.
  useEffect(() => {
    const isStuck = !currentSlide || currentSlide.status !== "active";
    if (isStuck && state.recoveryAttempts >= STUCK_THRESHOLD) {
      logger.warn(
        "Engine stuck: recoveryAttempts=%d, resetting playlist",
        state.recoveryAttempts,
      );
      dispatch({ type: "RESET_PLAYLIST" });
    }
  }, [currentSlide, state.recoveryAttempts]);

  useEffect(() => {
    logger.info("Playlist changed, resetting engine pointers");
    const timeoutId = setTimeout(() => {
      dispatch({ type: "RESET_PLAYLIST" });
      next();
    }, 0);
    return () => clearTimeout(timeoutId);
    // eslint-disable-next-line @eslint-react/exhaustive-deps
  }, [activePlaylist?.id]);

  // --- fallback in case there is no playlist ---
  if (!activePlaylist) {
    return {
      currentSlide: null,
      next,
      previous,
      togglePause,
      isPaused: state.isPaused,
      isUrgent: false,
      allowOverlay: true,
      toastSlides: [],
      duration: 10,
      stepInfo: null,
    };
  }

  return {
    currentSlide,
    next,
    previous,
    togglePause,
    isPaused: state.isPaused,
    isUrgent: currentSlide?.display_options?.is_urgent === true,
    allowOverlay: currentSlide?.display_options?.allow_social_overlay ?? false,
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
