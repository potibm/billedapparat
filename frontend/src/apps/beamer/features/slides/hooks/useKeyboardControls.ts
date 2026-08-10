import { useEffect } from "react";

export function useKeyboardControls(
  next: () => void,
  previous: () => void,
  togglePause: () => void,
) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "ArrowRight") next();
      if (e.key === "ArrowLeft") previous();
      if (e.key === " " || e.key === "Spacebar") {
        e.preventDefault();
        togglePause();
      }
    };
    globalThis.addEventListener("keydown", handleKeyDown);
    return () => globalThis.removeEventListener("keydown", handleKeyDown);
  }, [next, previous, togglePause]);
}
