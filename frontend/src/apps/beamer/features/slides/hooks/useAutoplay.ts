import { useEffect, useRef } from "react";

export function useAutoplay(
  next: () => void,
  duration: number,
  isPaused: boolean,
  isUrgent: boolean,
  hasCurrentSlide: boolean,
  resetKey: string | number | undefined,
) {
  const nextRef = useRef(next);

  useEffect(() => {
    nextRef.current = next;
  }, [next]);

  useEffect(() => {
    if (isPaused || isUrgent || !hasCurrentSlide) return;

    const timer = setTimeout(() => {
      nextRef.current();
    }, duration * 1000);

    return () => clearTimeout(timer);
  }, [duration, isPaused, isUrgent, hasCurrentSlide, resetKey]);
}