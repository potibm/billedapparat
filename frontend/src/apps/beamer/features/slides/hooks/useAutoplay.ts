import { useEffect, useRef } from "react";

export function useAutoplay(
  next: () => void,
  duration: number,
  isPaused: boolean,
  isUrgent: boolean,
  hasCurrentSlide: boolean,
) {
  const nextRef = useRef(next);

  useEffect(() => {
    nextRef.current = next;
  }, [next]);

  useEffect(() => {
    // When paused, urgent, or no slide present: stop the timer
    if (isPaused || isUrgent || !hasCurrentSlide) return;

    const timer = setInterval(() => {
      nextRef.current();
    }, duration * 1000);

    return () => clearInterval(timer);
  }, [duration, isPaused, isUrgent, hasCurrentSlide]);
}
