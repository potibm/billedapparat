import {
  animations,
  type AnimationType,
} from "../../animations/types/animations.schemas";
import type { Slide } from "../types/slide.schema";

export const getSlideAnimation = (
  currentSlide: Slide | null,
  isUrgent: boolean,
) => {
  if (!currentSlide) {
    return {
      activeAnimation: "fade" as AnimationType,
      transition: { duration: 0.8, ease: "easeInOut" as const },
    };
  }

  if (isUrgent) {
    return {
      activeAnimation: "urgent" as AnimationType,
      transition: { duration: 0.2, ease: "easeOut" as const },
    };
  }

  const keys = Object.keys(animations) as AnimationType[];
  if (keys.length === 0) {
    // Defensive fallback: if no animations are registered, behave like the
    // no-slide case to avoid `x % 0` and `undefined` lookups downstream.
    return {
      activeAnimation: "fade" as AnimationType,
      transition: { duration: 0.8, ease: "easeInOut" as const },
    };
  }
  const deterministicIndex = (currentSlide.id * 13) % keys.length;

  return {
    activeAnimation: keys[deterministicIndex],
    transition: { duration: 0.8, ease: "easeInOut" as const },
  };
};
