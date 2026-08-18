import {
  animations,
  type AnimationType,
} from "../../animations/types/animations.schemas";
import type { Slide } from "../types/slide.schema";

const ALL_ANIMATIONS: readonly AnimationType[] = Object.keys(
  animations,
) as AnimationType[];

const FADE_TRANSITION = { duration: 0.8, ease: "easeInOut" as const };
const URGENT_TRANSITION = { duration: 0.2, ease: "easeOut" as const };

const resolveAllowedKeys = (
  allowedAnimations: readonly AnimationType[],
): AnimationType[] => {
  const allowed = new Set(allowedAnimations);
  const filtered = ALL_ANIMATIONS.filter((key) => allowed.has(key));

  if (filtered.length === 0) {
    return ["fade"];
  }

  return filtered;
};

export const getSlideAnimation = (
  currentSlide: Slide | null,
  isUrgent: boolean,
  allowedAnimations: readonly AnimationType[] = ALL_ANIMATIONS,
) => {
  const allowedKeys = resolveAllowedKeys(allowedAnimations);

  if (!currentSlide) {
    return {
      activeAnimation: "fade" as AnimationType,
      transition: FADE_TRANSITION,
    };
  }

  if (isUrgent && allowedKeys.includes("urgent")) {
    return {
      activeAnimation: "urgent" as AnimationType,
      transition: URGENT_TRANSITION,
    };
  }

  const deterministicIndex = (currentSlide.id * 13) % allowedKeys.length;

  return {
    activeAnimation: allowedKeys[deterministicIndex],
    transition: FADE_TRANSITION,
  };
};
