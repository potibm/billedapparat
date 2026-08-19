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
  if (order === "asc" || order === "desc") {
    const dir = order === "asc" ? 1 : -1;
    return list.sort((a, b) => {
      // Primary: external_id (lexicographic — sorts YYYY-MM-DD dates
      // chronologically). Slides missing external_id sort last in asc,
      // first in desc.
      const ae = a.external_id;
      const be = b.external_id;
      const aHas = ae != null && ae !== "";
      const bHas = be != null && be !== "";
      if (aHas !== bHas) return dir * (aHas ? -1 : 1);
      if (!aHas) return 0; // both missing; fall through to tiebreakers

      // TS doesn't carry the narrowing through aHas/bHas; both are non-null strings here.
      const extCmp = (ae as string).localeCompare(be as string);
      if (extCmp !== 0) return dir * extCmp;

      // Secondary: external_sub_id (page within the external_id).
      const as = a.external_sub_id ?? 0;
      const bs = b.external_sub_id ?? 0;
      if (as !== bs) return dir * (as - bs);

      // Tiebreaker: id (DB-assigned, monotonic in batch inserts).
      return dir * (a.id - b.id);
    });
  }
  return list;
};

export const sortByPriorityDesc = (a: Slide, b: Slide) =>
  Number(b.display_options?.priority || 0) -
  Number(a.display_options?.priority || 0);
