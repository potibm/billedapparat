import { useEffect, useRef, useMemo } from "react";
import { useSlideshowEngine } from "./features/slides/hooks/useSlideshowEngine";
import { SlideRenderer } from "./features/slides/components/SlideRenderer";
import { AnimatePresence, motion } from "framer-motion";
import {
  animations,
  AnimationType,
} from "./features/animations/types/animations.schemas";
import { createLogger } from "@core/logger/logger";

const SLIDE_DURATION = 10000;

const effectLogger = createLogger("Effect");

export const BeamerApp = () => {
  const { currentSlide, next, previous, togglePause, isUrgent } =
    useSlideshowEngine();
  const containerRef = useRef<HTMLDivElement>(null);

  const activeAnimation = useMemo(() => {
    if (!currentSlide) return "fade"; // Fallback

    const keys = Object.keys(animations) as AnimationType[];
    const randomAnim = keys[Math.floor(Math.random() * keys.length)];

    effectLogger.debug(
      "Selected animation",
      "animation",
      randomAnim,
      "slideId",
      currentSlide.id,
    );
    return randomAnim;
  }, [currentSlide]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "ArrowRight") next();
      if (e.key === "ArrowLeft") previous();
      if (e.key === " " || e.key === "Spacebar") togglePause();
    };
    globalThis.addEventListener("keydown", handleKeyDown);
    return () => globalThis.removeEventListener("keydown", handleKeyDown);
  }, [next, previous, togglePause]);

  useEffect(() => {
    if (isUrgent) return;
    const timer = setInterval(next, SLIDE_DURATION);
    return () => clearInterval(timer);
  }, [next, isUrgent]);

  if (!currentSlide)
    return <div className="bg-black h-screen">We are empty...</div>;

  return (
    <div
      ref={containerRef}
      className={`h-screen w-screen transition-all duration-1000 ${isUrgent ? "border-8 border-red-600" : ""}`}
    >
      <AnimatePresence mode="wait">
        <motion.div
          key={currentSlide.id}
          {...animations[activeAnimation]}
          transition={{ duration: 1, ease: "easeInOut" }}
          className="absolute inset-0"
        >
          <SlideRenderer slide={currentSlide} />
        </motion.div>
      </AnimatePresence>
    </div>
  );
};
