import { useState, useEffect, useRef } from "react";
import { Slide } from "../types/slide.schema";

const TOAST_DURATION_SEC = 30;
const MAX_VISIBLE_TOASTS = 3;

type ActiveToast = {
  slide: Slide;
  timeLeft: number;
};

export const useToastEngine = (toastSlides: Slide[], allowOverlay: boolean) => {
  const [activeToasts, setActiveToasts] = useState<ActiveToast[]>(() => {
    return toastSlides
      .slice(0, MAX_VISIBLE_TOASTS)
      .reverse()
      .map((slide) => ({ slide, timeLeft: TOAST_DURATION_SEC }));
  });
  const seenIdsRef = useRef<Set<number>>(new Set(toastSlides.map((s) => s.id)));
  const pendingSlidesRef = useRef<Slide[]>([]);

  const unseenSlides = toastSlides.filter((s) => !seenIdsRef.current.has(s.id));
  if (unseenSlides.length > 0) {
    unseenSlides.forEach((s) => seenIdsRef.current.add(s.id));
    pendingSlidesRef.current = [...pendingSlidesRef.current, ...unseenSlides];
  }

  useEffect(() => {
    const timer = setInterval(() => {
      if (!allowOverlay) return;

      setActiveToasts((prev) => {
        let currentList = [...prev];

        // A. Decrement time and remove deleted slides
        currentList = currentList
          .map((t) => ({ ...t, timeLeft: t.timeLeft - 1 }))
          .filter(
            (t) =>
              t.timeLeft > 0 && toastSlides.some((s) => s.id === t.slide.id),
          );

        // B. Are there new slides in the inbox? Insert them now!
        if (pendingSlidesRef.current.length > 0) {
          const added = pendingSlidesRef.current
            .slice()
            .reverse() // Newest at the bottom
            .map((slide) => ({
              slide,
              timeLeft: TOAST_DURATION_SEC,
            }));

          currentList = [...currentList, ...added];

          // Clear inbox since we have now processed them
          pendingSlidesRef.current = [];
        }

        // C. Trim list to maximum (oldest at top are dropped)
        return currentList.slice(-MAX_VISIBLE_TOASTS);
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [allowOverlay, toastSlides]);

  if (!allowOverlay) return [];

  return activeToasts.map((t) => t.slide);
};
