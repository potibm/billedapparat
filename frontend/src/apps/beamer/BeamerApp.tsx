import { useEffect, useRef, useMemo } from "react";
import { useSlideshowEngine } from "./features/slides/hooks/useSlideshowEngine";
import { SlideRenderer } from "./features/slides/components/SlideRenderer";
import { AnimatePresence, motion } from "framer-motion";
import {
  animations,
  AnimationType,
} from "./features/animations/types/animations.schemas";
import { ToastManager } from "./features/slides/components/ToastManager";
import { useAppConfig } from "@core/config/useConfig";
import * as Sentry from "@sentry/react";

export const BeamerApp = () => {
  const {
    currentSlide,
    next,
    previous,
    togglePause,
    isUrgent,
    toastSlides,
    duration,
    stepInfo,
  } = useSlideshowEngine();

  const { version, environment } = useAppConfig();
  const containerRef = useRef<HTMLDivElement>(null);

  const allowOverlay =
    currentSlide?.display_options.allow_social_overlay ?? false;

  const activeAnimation = useMemo(() => {
    if (!currentSlide) return "fade";
    const keys = Object.keys(animations) as AnimationType[];
    return keys[Math.floor(Math.random() * keys.length)];
  }, [currentSlide]);

  // Keyboard controls
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

  useEffect(() => {
    if (isUrgent || !currentSlide) return;

    // duration is in seconds in the config, we need milliseconds
    const timer = setInterval(next, duration * 1000);

    return () => clearInterval(timer);
  }, [next, isUrgent, currentSlide, duration]);

  if (!currentSlide) {
    return (
      <div className="bg-black h-screen flex items-center justify-center text-slate-700">
        <div className="text-center">
          <p className="text-2xl font-mono">STANDBY</p>
          <p className="text-xs mt-2 uppercase tracking-widest">
            {environment} v{version}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className={`h-screen w-screen overflow-hidden bg-black transition-all duration-1000 ${
        isUrgent ? "ring-inset ring-12 ring-red-600" : ""
      }`}
    >
      <AnimatePresence mode="wait">
        <motion.div
          key={currentSlide.id}
          {...animations[activeAnimation]}
          transition={{ duration: 0.8, ease: "easeInOut" }}
          className="absolute inset-0"
        >
          <Sentry.ErrorBoundary fallback={<p>Slide Rendering Error</p>}>
            <SlideRenderer slide={currentSlide} />
          </Sentry.ErrorBoundary>
        </motion.div>
      </AnimatePresence>

      <ToastManager toastSlides={toastSlides} allowOverlay={allowOverlay} />

      {environment !== "production" && (
        <div className="absolute bottom-2 right-2 text-[10px] text-white/20 pointer-events-none font-mono">
          {environment} | v{version} | {stepInfo?.playlistName} |{" "}
          {stepInfo?.type} ({stepInfo?.current}/{stepInfo?.total}) |{" "}
          {activeAnimation} | {duration}s
        </div>
      )}
    </div>
  );
};
