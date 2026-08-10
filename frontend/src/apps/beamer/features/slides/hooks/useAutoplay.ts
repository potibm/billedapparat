import { useEffect } from "react";

export function useAutoplay(
  next: () => void,
  duration: number,
  isPaused: boolean,
  isUrgent: boolean,
  hasCurrentSlide: boolean,
) {
  useEffect(() => {
    // Wenn pausiert, dringend oder kein Slide da: Timer stoppen!
    if (isPaused || isUrgent || !hasCurrentSlide) return;

    const timer = setInterval(next, duration * 1000);
    return () => clearInterval(timer);
  }, [next, duration, isPaused, isUrgent, hasCurrentSlide]);
}
