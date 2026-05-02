import { useState, useEffect, useCallback, useMemo } from "react";
import { useSlideManager } from "./useSlideManager";
import { createLogger } from "@core/logger/logger";
import { Slide } from "../types/slide.schema";
import { useCurrentPlaylist } from "./useCurrentPlaylist";
import { PlaylistStep } from "@core/config/config.schemas";
import { SlideshowEngine } from "../types/slideshow.types";

const HISTORY_LIMIT = 50;
const NEXT_TICK_TIMEOUT = 0;

const logger = createLogger("Slideshow");

export const useSlideshowEngine = (): SlideshowEngine => {
  const { getByType, getUrgent, slides: allSlides } = useSlideManager();
  const activePlaylist = useCurrentPlaylist();

  const [stepIndex, setStepIndex] = useState(0);
  const [stepCountPointer, setStepCountPointer] = useState(0);

  const [history, setHistory] = useState<number[]>([]);
  const [historyPointer, setHistoryPointer] = useState(-1);
  const [isPaused, setIsPaused] = useState(false);

  const urgentSlides = getUrgent();
  const hasUrgent = urgentSlides.length > 0;
  const toastSlides = getByType("social.text");

  const sortSlides = useCallback((slides: Slide[], order: string): Slide[] => {
    const list = [...slides];
    switch (order) {
      case "asc":
        return list.sort(
          (a, b) =>
            new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
        );
      case "desc":
        return list.sort(
          (a, b) =>
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
        );
      case "random":
      default:
        return list;
    }
  }, []);

  const pickWeightedSlide = useCallback((slides: Slide[]): Slide | null => {
    if (slides.length === 0) return null;
    const totalWeight = slides.reduce(
      (sum, s) => sum + Number(s.display_options?.priority || 1),
      0,
    );
    let random = Math.random() * totalWeight;
    logger.debug(
      "Picking weighted slide",
      "totalWeight",
      totalWeight,
      "candidates",
      slides.map((s) => ({ id: s.id, priority: s.display_options?.priority })),
      "randomValue",
      random,
    );
    for (const slide of slides) {
      const weight = Number(slide.display_options?.priority || 1);
      if (random < weight) return slide;
      random -= weight;
    }
    return slides[0];
  }, []);

  const next = useCallback(() => {
    if (isPaused) return;

    if (historyPointer < history.length - 1) {
      setHistoryPointer((prev) => prev + 1);
      return;
    }

    const currentlyShownId = history[historyPointer];

    if (hasUrgent) {
      const selected = urgentSlides[0];
      if (selected && selected.id !== currentlyShownId) {
        setHistory((prev) => [...prev, selected.id].slice(-HISTORY_LIMIT));
        setHistoryPointer((prev) => Math.min(prev + 1, HISTORY_LIMIT - 1));
        return;
      }
    }

    // 3. Playlist Logic
    const steps = activePlaylist.steps;
    if (steps.length === 0) return;

    let currentStep = steps[stepIndex];
    let candidates = getByType(currentStep.type);

    let attempts = 0;
    while (candidates.length === 0 && attempts < steps.length) {
      const nextIdx = (stepIndex + 1) % steps.length;
      currentStep = steps[nextIdx];
      candidates = getByType(currentStep.type);
      setStepIndex(nextIdx);
      setStepCountPointer(0);
      attempts++;
    }

    // Sortierung und Auswahl
    const sortedCandidates = sortSlides(candidates, currentStep.order);

    // Bei "random" nutzen wir dein Gewichtungs-System
    let selected: Slide | null = null;
    if (currentStep.order === "random") {
      const otherCandidates = candidates.filter(
        (s) => s.id !== currentlyShownId,
      );
      const pool = otherCandidates.length > 0 ? otherCandidates : candidates;

      selected = pickWeightedSlide(pool);
      logger.debug("Weighted random selection", {
        type: currentStep.type,
        id: selected?.id,
        priority: selected?.display_options?.priority,
      });
    } else {
      // Bei asc/desc nehmen wir den nächsten aus der Liste basierend auf dem Pointer
      selected = sortedCandidates[stepCountPointer % sortedCandidates.length];
    }

    if (selected) {
      setHistory((prev) => [...prev, selected!.id].slice(-HISTORY_LIMIT));
      setHistoryPointer((prev) => Math.min(prev + 1, HISTORY_LIMIT - 1));

      const nextCountPointer = stepCountPointer + 1;
      if (nextCountPointer >= currentStep.count) {
        setStepIndex((prev) => (prev + 1) % steps.length);
        setStepCountPointer(0);
      } else {
        setStepCountPointer(nextCountPointer);
      }
    }
  }, [
    isPaused,
    historyPointer,
    history,
    hasUrgent,
    activePlaylist,
    stepIndex,
    stepCountPointer,
    getByType,
    sortSlides,
    urgentSlides,
    pickWeightedSlide,
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

  const currentStepConfig = useMemo(() => {
    return activePlaylist.steps[stepIndex];
  }, [activePlaylist, stepIndex]);

  const currentStep = useMemo<PlaylistStep | undefined>(() => {
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
  }, [activePlaylist.id]);

  return {
    currentSlide,
    next,
    previous,
    togglePause,
    isPaused,
    isUrgent: currentSlide?.content.type === "urgent",
    toastSlides,
    duration: currentStepConfig?.duration || 10,
    stepInfo: currentStep
      ? {
          type: currentStep.type,
          current: stepCountPointer + 1,
          total: currentStep.count,
          playlistName: activePlaylist.name,
        }
      : null,
  };
};
