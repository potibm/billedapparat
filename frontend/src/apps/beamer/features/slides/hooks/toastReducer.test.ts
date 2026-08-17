import { describe, it, expect } from "vitest";
import {
  toastReducer,
  initToastState,
  TOAST_DURATION_SEC,
  MAX_VISIBLE_TOASTS,
  type ToastState,
} from "./toastReducer";
import type { Slide } from "../types/slide.schema";

// Minimal slide factory — the reducer only touches `.id`, so anything
// else is filler.
const makeSlide = (id: number): Slide =>
  ({
    id,
    status: "active",
    created_at: "",
    content: {},
    display_options: {
      priority: 0,
      is_urgent: false,
      allow_social_overlay: false,
    },
  }) as unknown as Slide;

const EMPTY_STATE: ToastState = {
  activeToasts: [],
  seenIds: new Set(),
  pendingSlides: [],
};

describe("toastReducer", () => {
  describe("initToastState", () => {
    it("should initialize active toasts from the first MAX_VISIBLE_TOASTS slides in reverse order", () => {
      const slides = [makeSlide(1), makeSlide(2), makeSlide(3), makeSlide(4)];

      const state = initToastState(slides);

      // 4 slides provided -> only first 3 taken, reversed -> [3, 2, 1]
      expect(state.activeToasts.map((t) => t.slide.id)).toEqual([3, 2, 1]);
      expect(
        state.activeToasts.every((t) => t.timeLeft === TOAST_DURATION_SEC),
      ).toBe(true);
    });

    it("should seed seenIds with every provided slide id", () => {
      const state = initToastState([makeSlide(1), makeSlide(2)]);

      expect(state.seenIds.has(1)).toBe(true);
      expect(state.seenIds.has(2)).toBe(true);
      expect(state.seenIds.size).toBe(2);
    });

    it("should start with an empty pending queue", () => {
      const state = initToastState([makeSlide(1)]);

      expect(state.pendingSlides).toEqual([]);
    });

    it("should handle an empty initial list", () => {
      const state = initToastState([]);

      expect(state.activeToasts).toEqual([]);
      expect(state.seenIds.size).toBe(0);
      expect(state.pendingSlides).toEqual([]);
    });
  });

  describe("SYNC_SLIDES", () => {
    it("should queue new slides into pendingSlides", () => {
      const state: ToastState = {
        ...EMPTY_STATE,
        seenIds: new Set([1]),
      };

      const next = toastReducer(state, {
        type: "SYNC_SLIDES",
        payload: [makeSlide(1), makeSlide(2), makeSlide(3)],
      });

      // 1 is already seen; 2 and 3 are new and should be queued.
      expect(next.pendingSlides.map((s) => s.id)).toEqual([2, 3]);
      expect(next.seenIds.has(1)).toBe(true);
      expect(next.seenIds.has(2)).toBe(true);
      expect(next.seenIds.has(3)).toBe(true);
    });

    it("should not re-queue slides that are already seen", () => {
      const seen = new Set([1, 2]);
      const state: ToastState = {
        ...EMPTY_STATE,
        seenIds: seen,
      };

      const next = toastReducer(state, {
        type: "SYNC_SLIDES",
        payload: [makeSlide(1), makeSlide(2)],
      });

      expect(next.pendingSlides).toEqual([]);
    });

    it("should drop active toasts whose slide id is no longer incoming", () => {
      const state: ToastState = {
        seenIds: new Set([1, 2]),
        activeToasts: [
          { slide: makeSlide(1), timeLeft: 10 },
          { slide: makeSlide(2), timeLeft: 5 },
        ],
        pendingSlides: [],
      };

      // 1 is still incoming, 2 is gone
      const next = toastReducer(state, {
        type: "SYNC_SLIDES",
        payload: [makeSlide(1)],
      });

      expect(next.activeToasts).toHaveLength(1);
      expect(next.activeToasts[0].slide.id).toBe(1);
    });

    it("should drop pendingSlides entries that are no longer incoming", () => {
      const state: ToastState = {
        seenIds: new Set([1]),
        activeToasts: [],
        pendingSlides: [makeSlide(2), makeSlide(3)],
      };

      const next = toastReducer(state, {
        type: "SYNC_SLIDES",
        payload: [makeSlide(1)],
      });

      // 3 was pending but is no longer incoming -> dropped
      // 2 was pending, ALSO no longer incoming, so it should also be dropped
      expect(next.pendingSlides).toEqual([]);
    });

    it("should preserve pending entries that survive the incoming filter", () => {
      // Slide 2 is already pending and is still in the incoming payload.
      // It must remain in pendingSlides — and must NOT be duplicated,
      // even though it is not yet in seenIds.
      const state: ToastState = {
        seenIds: new Set([1]),
        activeToasts: [],
        pendingSlides: [makeSlide(2)],
      };

      const next = toastReducer(state, {
        type: "SYNC_SLIDES",
        payload: [makeSlide(2)],
      });

      expect(next.pendingSlides.map((s) => s.id)).toEqual([2]);
    });

    it("should preserve pending entries that survive and add unseen incoming as new", () => {
      // seenIds contains 1. Pending = [3]. Incoming = [1, 3, 4].
      // - nextPending (filter incoming) = [3]
      // - newSlides (incoming but not in seenIds) = [3, 4]
      //   (1 was already in seenIds, so excluded)
      // - The reducer dedupes newSlides against nextPending, so 3 is
      //   not appended again. Final pendingSlides = [3, 4].
      const state: ToastState = {
        seenIds: new Set([1]),
        activeToasts: [],
        pendingSlides: [makeSlide(3)],
      };

      const next = toastReducer(state, {
        type: "SYNC_SLIDES",
        payload: [makeSlide(1), makeSlide(3), makeSlide(4)],
      });

      expect(next.pendingSlides.map((s) => s.id)).toEqual([3, 4]);
    });

    it("should not duplicate a slide that exists in both pendingSlides and incomingSlides when not yet in seenIds", () => {
      // Regression test for the duplication bug from
      // `.opencode/reviews/refactor-beamer-review.md` (High severity):
      // before the fix, a slide in both pendingSlides and incomingSlides
      // (but not yet in seenIds) would appear twice in the resulting
      // pending list. On the next TICK both copies would get promoted to
      // activeToasts for the full TOAST_DURATION_SEC, causing the same
      // toast to display twice ~30s apart.
      const state: ToastState = {
        seenIds: new Set([1]),
        activeToasts: [],
        pendingSlides: [makeSlide(2)],
      };

      const next = toastReducer(state, {
        type: "SYNC_SLIDES",
        payload: [makeSlide(2)],
      });

      expect(next.pendingSlides).toHaveLength(1);
      expect(next.pendingSlides.every((s) => s.id === 2)).toBe(true);
    });

    it("should prune seenIds for slides no longer in the incoming payload", () => {
      // Regression test for Bug #3 from `.opencode/reviews/beamer-refactor-2026-08-16.md`:
      // seenIds used to grow unbounded because the reducer only added, never pruned.
      const state: ToastState = {
        ...EMPTY_STATE,
        seenIds: new Set([1, 2, 3]),
      };

      const next = toastReducer(state, {
        type: "SYNC_SLIDES",
        payload: [makeSlide(1)], // 2 and 3 disappeared
      });

      // 1 is still incoming, so it stays; 2 and 3 are pruned.
      expect(next.seenIds.has(1)).toBe(true);
      expect(next.seenIds.has(2)).toBe(false);
      expect(next.seenIds.has(3)).toBe(false);
      expect(next.seenIds.size).toBe(1);
    });
  });

  describe("TICK", () => {
    it("should decrement timeLeft on every active toast", () => {
      const state: ToastState = {
        seenIds: new Set(),
        activeToasts: [
          { slide: makeSlide(1), timeLeft: 10 },
          { slide: makeSlide(2), timeLeft: 5 },
        ],
        pendingSlides: [],
      };

      const next = toastReducer(state, { type: "TICK" });

      expect(next.activeToasts[0].timeLeft).toBe(9);
      expect(next.activeToasts[1].timeLeft).toBe(4);
    });

    it("should remove toasts whose timer hits zero", () => {
      const state: ToastState = {
        seenIds: new Set(),
        activeToasts: [
          { slide: makeSlide(1), timeLeft: 1 },
          { slide: makeSlide(2), timeLeft: 10 },
        ],
        pendingSlides: [],
      };

      const next = toastReducer(state, { type: "TICK" });

      expect(next.activeToasts.map((t) => t.slide.id)).toEqual([2]);
    });

    it("should promote pending slides into activeToasts on tick", () => {
      const newSlide = makeSlide(99);
      const state: ToastState = {
        seenIds: new Set([1]),
        activeToasts: [{ slide: makeSlide(1), timeLeft: 5 }],
        pendingSlides: [newSlide],
      };

      const next = toastReducer(state, { type: "TICK" });

      expect(next.pendingSlides).toEqual([]);
      const promoted = next.activeToasts.find((t) => t.slide.id === 99);
      expect(promoted).toBeDefined();
      expect(promoted?.timeLeft).toBe(TOAST_DURATION_SEC);
    });

    it("should promote pending slides in reverse order (oldest last in queue becomes first active)", () => {
      const state: ToastState = {
        seenIds: new Set(),
        activeToasts: [],
        pendingSlides: [makeSlide(10), makeSlide(20), makeSlide(30)],
      };

      const next = toastReducer(state, { type: "TICK" });

      // pendingSlides is iterated slice().reverse(), so the LAST pending slide
      // becomes the FIRST newly-promoted active toast.
      const newIds = next.activeToasts.map((t) => t.slide.id);
      expect(newIds[0]).toBe(30);
      expect(newIds[1]).toBe(20);
      expect(newIds[2]).toBe(10);
    });

    it("should cap activeToasts at MAX_VISIBLE_TOASTS (oldest are dropped)", () => {
      // 2 existing with timeLeft=0 (about to expire) and 3 pending.
      // After tick: existing expire, pending get promoted reverse-order,
      // then cap cuts to MAX_VISIBLE_TOASTS.
      const state: ToastState = {
        seenIds: new Set([1, 2]),
        activeToasts: [
          { slide: makeSlide(1), timeLeft: 1 },
          { slide: makeSlide(2), timeLeft: 1 },
        ],
        pendingSlides: [makeSlide(3), makeSlide(4), makeSlide(5)],
      };

      const next = toastReducer(state, { type: "TICK" });

      expect(next.activeToasts).toHaveLength(MAX_VISIBLE_TOASTS);
      // Existing expire, pending reverse-pick [5, 4, 3], cap = [5, 4, 3]
      expect(next.activeToasts.map((t) => t.slide.id)).toEqual([5, 4, 3]);
    });

    it("should not promote when pendingSlides is empty", () => {
      const state: ToastState = {
        seenIds: new Set([1]),
        activeToasts: [{ slide: makeSlide(1), timeLeft: 5 }],
        pendingSlides: [],
      };

      const next = toastReducer(state, { type: "TICK" });

      expect(next.activeToasts).toHaveLength(1);
      expect(next.activeToasts[0].timeLeft).toBe(4);
      expect(next.pendingSlides).toEqual([]);
    });

    it("should clear pendingSlides even when MAX_VISIBLE_TOASTS would still cap them", () => {
      const state: ToastState = {
        seenIds: new Set(),
        activeToasts: [
          { slide: makeSlide(1), timeLeft: 5 },
          { slide: makeSlide(2), timeLeft: 5 },
          { slide: makeSlide(3), timeLeft: 5 },
        ],
        pendingSlides: [makeSlide(4), makeSlide(5)],
      };

      const next = toastReducer(state, { type: "TICK" });

      // pendingSlides is consumed in full each tick; overflow is dropped later.
      expect(next.pendingSlides).toEqual([]);
    });
  });

  describe("default", () => {
    it("should return the same state reference for unknown actions", () => {
      const state = toastReducer(EMPTY_STATE, {
        type: "UNKNOWN" as never,
        payload: undefined as never,
      });

      expect(state).toBe(EMPTY_STATE);
    });
  });
});
