import { useEffect, useReducer, useRef } from "react";
import { Slide } from "../types/slide.schema";
import { toastReducer, initToastState } from "./toastReducer";

export const useToastEngine = (toastSlides: Slide[], allowOverlay: boolean) => {
  const [state, dispatch] = useReducer(
    toastReducer,
    toastSlides,
    initToastState,
  );

  const prevIdsRef = useRef<string>("");

  // Effect 1: Push new slides into the reducer
  useEffect(() => {
    // signature of current slides
    const currentIds = toastSlides.map((s) => s.id).join(",");

    // only dispatch, wenn the signature has changed
    if (prevIdsRef.current !== currentIds) {
      dispatch({ type: "SYNC_SLIDES", payload: toastSlides });
      prevIdsRef.current = currentIds;
    }
  }, [toastSlides]);

  // Effect 2: Timer tick
  useEffect(() => {
    if (!allowOverlay) return; // stop the timer when no overlays are allowed

    const timer = setInterval(() => {
      dispatch({ type: "TICK" });
    }, 1000);

    return () => clearInterval(timer);
  }, [allowOverlay]);

  // UI return
  if (!allowOverlay) return [];

  return state.activeToasts.map((t) => t.slide);
};
