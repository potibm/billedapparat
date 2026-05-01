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

        // A. Zeit abziehen und gelöschte Slides entfernen
        currentList = currentList
          .map((t) => ({ ...t, timeLeft: t.timeLeft - 1 }))
          .filter(
            (t) =>
              t.timeLeft > 0 && toastSlides.some((s) => s.id === t.slide.id),
          );

        // B. Gibt es neue Slides im Posteingang? Dann jetzt einfügen!
        if (pendingSlidesRef.current.length > 0) {
          const added = pendingSlidesRef.current
            .slice()
            .reverse() // Die neuesten nach unten
            .map((slide) => ({
              slide,
              timeLeft: TOAST_DURATION_SEC,
            }));

          currentList = [...currentList, ...added];

          // Posteingang leeren, da wir sie jetzt verarbeitet haben
          pendingSlidesRef.current = [];
        }

        // C. Liste auf Maximum kürzen (älteste oben fallen raus)
        return currentList.slice(-MAX_VISIBLE_TOASTS);
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [allowOverlay, toastSlides]);

  if (!allowOverlay) return [];

  return activeToasts.map((t) => t.slide);
};
