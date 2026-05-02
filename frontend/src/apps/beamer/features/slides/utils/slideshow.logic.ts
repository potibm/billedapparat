import { PlaylistStep } from "@core/config/config.schemas";
import { Slide } from "../types/slide.schema";

/* find the next valid step if the current one has no slides */
export const findNextValidStep = (
  steps: PlaylistStep[],
  startIndex: number,
  getSlidesByType: (type: string) => Slide[],
) => {
  let attempts = 0;
  let currentIndex = startIndex;

  while (attempts < steps.length) {
    const step = steps[currentIndex];
    const candidates = getSlidesByType(step.type);

    if (candidates.length > 0) {
      return { step, index: currentIndex, candidates };
    }

    currentIndex = (currentIndex + 1) % steps.length;
    attempts++;
  }
  return null;
};

/* logic to select the next slide based on sorting */
export const selectNextSlide = (
  candidates: Slide[],
  step: PlaylistStep,
  pointer: number,
  lastId: number | undefined,
): Slide | null => {
  if (candidates.length === 0) return null;

  if (step.order === "random") {
    const otherCandidates = candidates.filter((s) => s.id !== lastId);
    const pool = otherCandidates.length > 0 ? otherCandidates : candidates;
    return pickWeightedSlide(pool);
  }

  return candidates[pointer % candidates.length];
};

/* Generic function to pick a weighted item */
export const pickWeighted = <T>(
  items: T[],
  getWeight: (item: T) => number,
): T | null => {
  if (items.length === 0) return null;

  const totalWeight = items.reduce(
    (sum, item) => sum + Math.max(0, getWeight(item)),
    0,
  );

  if (totalWeight === 0) {
    return items[Math.floor(Math.random() * items.length)];
  }

  let random = Math.random() * totalWeight;

  for (const item of items) {
    const weight = getWeight(item);
    if (random < weight) return item;
    random -= weight;
  }

  return items[0];
};

/* Specific wrapper function for slides to reduce complexity in the hook. */
export const pickWeightedSlide = (slides: Slide[]): Slide | null => {
  return pickWeighted(slides, (s) => s.display_options?.priority ?? 1);
};

export const sortSlides = (
  slides: Slide[],
  order: "asc" | "desc" | (string & {}),
): Slide[] => {
  const list = [...slides];
  if (order === "asc") {
    return list.sort(
      (a, b) =>
        new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
    );
  }
  if (order === "desc") {
    return list.sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );
  }
  return list;
};
