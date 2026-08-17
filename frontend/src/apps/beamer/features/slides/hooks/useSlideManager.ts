import { useSyncExternalStore, useCallback } from "react";
import { slideStore } from "../store/slideStore";

const subscribe = slideStore.subscribe.bind(slideStore);

/**
 * useSlideManager (React Bridge)
 *
 * WHAT IT DOES:
 * - Acts as a bridge between the framework-agnostic `slideStore` and the React component tree.
 * - Subscribes to store updates and forces a React re-render ONLY when the store data actually mutates.
 * - Exposes memoized getter functions to safely retrieve slides without triggering infinite render loops.
 *
 * WHAT IT DOES NOT DO:
 * - It does NOT manage the SSE connection (this is handled globally by `slideStore.connect()`).
 * - It does NOT hold the actual slide data in a `useState` array (to avoid memory duplication and complex sync issues).
 */
export const useSlideManager = () => {
  const slides = useSyncExternalStore(subscribe, () => slideStore.getSlides());

  const getByType = useCallback(
    (type: string) => slideStore.getByType(type),
    [],
  );

  const getUrgent = useCallback(() => slideStore.getUrgent(), []);

  return {
    slides,
    getByType,
    getUrgent,
  };
};
