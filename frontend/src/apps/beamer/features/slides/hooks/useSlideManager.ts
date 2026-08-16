import { useState, useEffect, useCallback, useMemo } from "react";
import { slideStore } from "../store/slideStore";

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
  // We use a simple counter to force a React re-render when the external store notifies us.
  const [tick, setTick] = useState(0);

  useEffect(() => {
    // Subscribe to store mutations
    const unsubscribe = slideStore.subscribe(() => {
      setTick((t) => t + 1); // Triggers a re-render
    });

    // Cleanup subscription on unmount
    return () => {
      unsubscribe();
    };
  }, []);

  // --- Memoized State Getters ---
  // The `[tick]` dependency is crucial: Every time the store mutates (tick increases),
  // these functions evaluate fresh data from the store.

  // eslint-disable-next-line @eslint-react/exhaustive-deps
  const slides = useMemo(() => slideStore.getSlides(), [tick]);

  const getByType = useCallback(
    (type: string) => slideStore.getByType(type),
    // eslint-disable-next-line @eslint-react/exhaustive-deps
    [tick],
  );

  // eslint-disable-next-line @eslint-react/exhaustive-deps
  const getUrgent = useCallback(() => slideStore.getUrgent(), [tick]);

  return {
    slides,
    getByType,
    getUrgent,
  };
};
