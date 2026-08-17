import { Slide } from "../types/slide.schema";

export const TOAST_DURATION_SEC = 30;
export const MAX_VISIBLE_TOASTS = 3;

export type ActiveToast = {
  slide: Slide;
  timeLeft: number;
};

export interface ToastState {
  activeToasts: ActiveToast[];
  seenIds: Set<number>;
  pendingSlides: Slide[];
}

export type ToastAction =
  { type: "SYNC_SLIDES"; payload: Slide[] } | { type: "TICK" };

export const initToastState = (initialSlides: Slide[]): ToastState => ({
  activeToasts: initialSlides
    .slice(0, MAX_VISIBLE_TOASTS)
    .reverse()
    .map((slide) => ({ slide, timeLeft: TOAST_DURATION_SEC })),
  seenIds: new Set(initialSlides.map((s) => s.id)),
  pendingSlides: [],
});

export const toastReducer = (
  state: ToastState,
  action: ToastAction,
): ToastState => {
  switch (action.type) {
    case "SYNC_SLIDES": {
      const incomingSlides = action.payload;
      const incomingIds = new Set(incomingSlides.map((s) => s.id));

      // 1. Identify new slides
      const newSlides = incomingSlides.filter((s) => !state.seenIds.has(s.id));

      // 2. Update seenIds
      const nextSeenIds = new Set(state.seenIds);
      nextSeenIds.forEach((id) => {
        if (!incomingIds.has(id)) nextSeenIds.delete(id);
      });
      newSlides.forEach((s) => nextSeenIds.add(s.id));

      // 3. Remove old/deleted slides from active & pending
      const nextActive = state.activeToasts.filter((t) =>
        incomingIds.has(t.slide.id),
      );
      let nextPending = state.pendingSlides.filter((s) =>
        incomingIds.has(s.id),
      );

      // 4. Add slides into the queue
      if (newSlides.length > 0) {
        const nextPendingIds = new Set(nextPending.map((s) => s.id));
        const dedupedNew = newSlides.filter((s) => !nextPendingIds.has(s.id));

        if (dedupedNew.length > 0) {
          nextPending = [...nextPending, ...dedupedNew];
        }
      }

      return {
        ...state,
        seenIds: nextSeenIds,
        activeToasts: nextActive,
        pendingSlides: nextPending,
      };
    }

    case "TICK": {
      // 1. Reduce time and remove expired
      let nextActive = state.activeToasts
        .map((t) => ({ ...t, timeLeft: t.timeLeft - 1 }))
        .filter((t) => t.timeLeft > 0);

      let nextPending = state.pendingSlides;

      // 2. Adding pending slides
      if (nextPending.length > 0) {
        const added = nextPending
          .slice()
          .reverse() // from the bottom
          .map((slide) => ({ slide, timeLeft: TOAST_DURATION_SEC }));

        nextActive = [...nextActive, ...added];
        nextPending = []; // clean inbox
      }

      // 3. Reduce to maximum size (throw out oldest))
      nextActive = nextActive.slice(-MAX_VISIBLE_TOASTS);

      return {
        ...state,
        activeToasts: nextActive,
        pendingSlides: nextPending,
      };
    }

    default:
      return state;
  }
};
