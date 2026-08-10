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
import { slideStore } from "./features/slides/store/slideStore";
import { useKeyboardControls } from "./features/slides/hooks/useKeyboardControls";
import { useAutoplay } from "./features/slides/hooks/useAutoplay";
import { DebugOverlay } from "./components/DebugOverlay";

export const BeamerApp = () => {
  const {
    currentSlide,
    next,
    previous,
    togglePause,
    isUrgent,
    isPaused,
    toastSlides,
    duration,
    stepInfo,
  } = useSlideshowEngine();

  const { version, environment } = useAppConfig();
  const containerRef = useRef<HTMLDivElement>(null);

  // 1. Globale Verbindung
  useEffect(() => {
    slideStore.connect();
    return () => slideStore.disconnect();
  }, []);

  // 2. Logik-Hooks aufrufen (Der Code bleibt sauber!)
  useKeyboardControls(next, previous, togglePause);
  useAutoplay(next, duration, isPaused, isUrgent, !!currentSlide);

  // 3. UI Helper
  const allowOverlay =
    currentSlide?.display_options.allow_social_overlay ?? false;

  const activeAnimation = useMemo(() => {
    if (!currentSlide) return "fade";
    if (isUrgent) return "urgent";
    const keys = Object.keys(animations) as AnimationType[];
    return keys[Math.floor(Math.random() * keys.length)];
  }, [currentSlide, isUrgent]);

  const transition = useMemo(() => {
    return {
      duration: isUrgent ? 0.2 : 0.8,
      ease: (isUrgent ? "easeOut" : "easeInOut") as "easeOut" | "easeInOut",
    };
  }, [isUrgent]);

  // 4. Render Logik
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
      className={`h-screen w-screen overflow-hidden bg-black ${
        isUrgent ? "ring-inset ring-12 ring-red-600" : ""
      }`}
    >
      <AnimatePresence mode={isUrgent ? "sync" : "wait"}>
        <motion.div
          key={currentSlide.id}
          {...animations[activeAnimation]}
          transition={transition}
          className="absolute inset-0"
        >
          <Sentry.ErrorBoundary fallback={<p>Slide Rendering Error</p>}>
            <SlideRenderer slide={currentSlide} />
          </Sentry.ErrorBoundary>
        </motion.div>
      </AnimatePresence>

      <ToastManager toastSlides={toastSlides} allowOverlay={allowOverlay} />

      <DebugOverlay
        isUrgent={isUrgent}
        stepInfo={stepInfo}
        activeAnimation={activeAnimation}
        duration={duration}
        isPaused={isPaused}
      />
    </div>
  );
};
