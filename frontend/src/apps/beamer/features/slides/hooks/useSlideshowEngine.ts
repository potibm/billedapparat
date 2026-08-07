import { useState, useEffect, useCallback, useMemo } from "react";
import { useSlideManager } from "./useSlideManager";
import { createLogger } from "@core/logger/logger";
import { useCurrentPlaylist } from "./useCurrentPlaylist";
import { PlaylistStep } from "@core/config/config.schemas";
import { SlideshowEngine } from "../types/slideshow.types";
import {
  pickWeightedSlide,
  sortSlides,
  selectNextSlide,
  findNextValidStep,
} from "../utils/slideshow.logic";

const HISTORY_LIMIT = 50;
const NEXT_TICK_TIMEOUT = 0;

const logger = createLogger("Slideshow");

export const useSlideshowEngine = (): SlideshowEngine => {
  const { getByType, getUrgent, slides: allSlides } = useSlideManager();
  const activePlaylist = useCurrentPlaylist();

  const [stepIndex, setStepIndex] = useState(0);
  const [stepCountPointer, setStepCountPointer] = useState(0);

  const [displayedStepInfo, setDisplayedStepInfo] = useState<{
    stepIndex: number;
    stepCountPointer: number;
  } | null>(null);

  const [history, setHistory] = useState<number[]>([]);
  const [historyPointer, setHistoryPointer] = useState(-1);
  const [isPaused, setIsPaused] = useState(false);

  const urgentSlides = getUrgent();
  const hasUrgent = urgentSlides.length > 0;
  const toastSlides = getByType("social.text");

  const updateHistory = useCallback((id: number) => {
    setHistory((prev) => [...prev, id].slice(-HISTORY_LIMIT));
    setHistoryPointer((prev) => Math.min(prev + 1, HISTORY_LIMIT - 1));
  }, []);

  const advanceStepPointers = useCallback(
    (currentStepCount: number, stepsLength: number) => {
      setStepCountPointer((prevCount) => {
        const nextCount = prevCount + 1;
        if (nextCount >= currentStepCount) {
          setStepIndex((prevIndex) => (prevIndex + 1) % stepsLength);
          return 0;
        }
        return nextCount;
      });
    },
    [],
  );

  const next = useCallback(() => {
    // 1. Guard & history navigation
    if (isPaused || !activePlaylist?.steps) return;

    if (historyPointer < history.length - 1) {
      setHistoryPointer((prev) => prev + 1);
      return;
    }

    const currentlyShownId = history[historyPointer];

    // 2. Urgent override
    if (hasUrgent) {
      const selected = pickWeightedSlide(urgentSlides);
      if (selected && selected.id !== currentlyShownId) {
        setHistory((prev) => [...prev, selected.id].slice(-HISTORY_LIMIT));
        setHistoryPointer((prev) => Math.min(prev + 1, HISTORY_LIMIT - 1));
      }
      return; // Immer zurückkehren, wenn hasUrgent true ist
    }

    // 3. Find playlist step
    const result = findNextValidStep(
      activePlaylist.steps,
      stepIndex,
      getByType,
    );

    if (!result) {
      logger.warn("No slides found for any step in playlist");
      return;
    }

    const { step, index: foundIndex, candidates } = result;

    // In case findNextValidStep found a different index than the current one:
    if (foundIndex !== stepIndex) {
      setStepIndex(foundIndex);
      setStepCountPointer(0);
    }

    // 4. Select slide
    const sorted = sortSlides(candidates, step.order);
    const selected = selectNextSlide(
      sorted,
      step,
      stepCountPointer,
      currentlyShownId,
    );

    // 5. Syncronize state
    if (selected) {
      setDisplayedStepInfo({
        stepIndex: foundIndex,
        stepCountPointer: stepCountPointer,
      });
      updateHistory(selected.id);
      advanceStepPointers(step.count, activePlaylist.steps.length);
    }
  }, [
    isPaused,
    historyPointer,
    history,
    hasUrgent,
    activePlaylist,
    stepIndex,
    stepCountPointer,
    urgentSlides,
    getByType,
    advanceStepPointers,
    updateHistory,
  ]);

  const previous = useCallback(() => {
    if (historyPointer > 0) {
      setHistoryPointer((prev) => prev - 1);
    }
  }, [historyPointer]);

  const togglePause = useCallback(() => {
    setIsPaused((prev) => !prev);
  }, []);

  const currentSlide = useMemo(() => {
    const id = history[historyPointer];
    return allSlides.find((s) => s.id === id) || null;
  }, [history, historyPointer, allSlides]);

  const currentStep = useMemo<PlaylistStep | undefined>(() => {
    if (!activePlaylist) return undefined;
    return activePlaylist.steps[stepIndex];
  }, [activePlaylist, stepIndex]);

  useEffect(() => {
    if (history.length === 0 && allSlides.length > 0) {
      logger.debug("Kickstarting initial slide");
      const timeoutId = setTimeout(next, NEXT_TICK_TIMEOUT);
      return () => clearTimeout(timeoutId);
    }
  }, [history.length, allSlides.length, next]);

  useEffect(() => {
    if (hasUrgent && currentSlide?.content.type !== "urgent") {
      const timeoutId = setTimeout(next, NEXT_TICK_TIMEOUT);
      return () => clearTimeout(timeoutId);
    }
  }, [hasUrgent, currentSlide, next]);

  useEffect(() => {
    logger.info("Playlist changed, resetting engine pointers");
    const timeoutId = setTimeout(() => {
      setStepIndex(0);
      setStepCountPointer(0);
      next();
    }, 0);

    return () => clearTimeout(timeoutId);
    // eslint-disable-next-line @eslint-react/exhaustive-deps
  }, [activePlaylist?.id]);

  if (!activePlaylist) {
    return {
      currentSlide: null,
      next,
      previous,
      togglePause: () => setIsPaused((p) => !p),
      isPaused,
      isUrgent: false,
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
    isPaused,
    isUrgent: currentSlide?.content.type === "urgent",
    toastSlides,
    duration: currentStep?.duration || 10,
    stepInfo:
      displayedStepInfo && activePlaylist
        ? {
            type: activePlaylist.steps[displayedStepInfo.stepIndex].type,
            current: displayedStepInfo.stepCountPointer + 1,
            total: activePlaylist.steps[displayedStepInfo.stepIndex].count,
            playlistName: activePlaylist.name,
          }
        : null,
  };
};
