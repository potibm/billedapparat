import { useState, useEffect, useCallback, useMemo } from "react";
import { useSlideManager } from "./useSlideManager";
import { createLogger } from "@core/logger/logger";
import { Slide } from "../types/slide.schema";

const PLAYLIST_PATTERN = [
  "news",
  "sponsor",
  "sponsor",
  "news",
/*  "timetable",
 "sponsor",
  "social",
  "social",*/
];
const HISTORY_LIMIT = 50;
const NEXT_TICK_TIMEOUT = 0;

const logger = createLogger("Slideshow");

export const useSlideshowEngine = () => {
  const { getByType, getUrgent, slides: allSlides } = useSlideManager();

  const [currentIndex, setCurrentIndex] = useState(-1);
  const [history, setHistory] = useState<number[]>([]);
  const [historyPointer, setHistoryPointer] = useState(-1);

  const urgentSlides = getUrgent();
  const hasUrgent = urgentSlides.length > 0;

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
    if (historyPointer < history.length - 1) {
      setHistoryPointer((prev) => prev + 1);
      return;
    }

    const currentlyShownId = history[historyPointer];

    if (hasUrgent) {
      const urgentCandidates =
        urgentSlides.length > 1
          ? urgentSlides.filter((s) => s.id !== currentlyShownId)
          : urgentSlides;

      const selected = pickWeightedSlide(urgentCandidates);
      if (selected) {
        logger.debug("Selected URGENT slide", "id", selected.id);
        setHistory((prev) => [...prev, selected.id].slice(-HISTORY_LIMIT));
        setHistoryPointer((prev) => Math.min(prev + 1, HISTORY_LIMIT - 1));
        return;
      }
    }

    let nextIndex = currentIndex;
    let attempts = 0;
    let foundSlide: Slide | null = null;

    while (attempts < PLAYLIST_PATTERN.length) {
      nextIndex = (nextIndex + 1) % PLAYLIST_PATTERN.length;
      const nextType = PLAYLIST_PATTERN[nextIndex];

      const availableSlides = getByType(nextType);
      const candidates =
        availableSlides.length > 1
          ? availableSlides.filter((s) => s.id !== currentlyShownId)
          : availableSlides;

      foundSlide = pickWeightedSlide(candidates);

      if (foundSlide) {
        logger.debug(
          "Selected next slide",
          "id",
          foundSlide.id,
          "type",
          nextType,
        );
        setHistory((prev) => [...prev, foundSlide!.id].slice(-HISTORY_LIMIT));
        setHistoryPointer((prev) => Math.min(prev + 1, HISTORY_LIMIT - 1));
        setCurrentIndex(nextIndex);
        return;
      }

      logger.warn(
        `No active slides found for type "${nextType}". Skipping immediately to next type.`,
      );
      attempts++;
    }

    if (!foundSlide && allSlides.length > 0) {
      logger.error(
        "Playlist pattern mismatch: No slides matched any type in the pattern.",
      );
    }
  }, [
    currentIndex,
    getByType,
    hasUrgent,
    history,
    historyPointer,
    pickWeightedSlide,
    urgentSlides,
    allSlides.length,
  ]);

  const previous = useCallback(() => {
    if (historyPointer > 0) {
      setHistoryPointer((prev) => prev - 1);
    }
  }, [historyPointer]);

  const currentSlide = useMemo(() => {
    const id = history[historyPointer];
    return allSlides.find((s) => s.id === id) || null;
  }, [history, historyPointer, allSlides]);

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

  return {
    currentSlide,
    next,
    previous,
    isUrgent: currentSlide?.content.type === "urgent",
  };
};
