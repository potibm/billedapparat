import { useRef } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { animations } from "./features/animations/types/animations.schemas";
import { useAppConfig } from "@core/config/useConfig";
import * as Sentry from "@sentry/react";

// hooks
import { useAutoplay } from "./features/slides/hooks/useAutoplay";
import { useKeyboardControls } from "./features/slides/hooks/useKeyboardControls";
import { useSSEConnection } from "./features/slides/hooks/useSSEConnection";
import { getSlideAnimation } from "./features/slides/utils/getSlideAnimation";
import { useSlideshowEngine } from "./features/slides/hooks/useSlideshowEngine";

// components
import { DebugOverlay } from "./components/DebugOverlay";
import { SlideRenderer } from "./features/slides/components/SlideRenderer";
import { ToastManager } from "./features/slides/components/ToastManager";

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
    allowOverlay,
  } = useSlideshowEngine();

  const { version, environment } = useAppConfig();
  const containerRef = useRef<HTMLDivElement>(null);

  // 1. side effects and logic hooks
  useSSEConnection();
  useKeyboardControls(next, previous, togglePause);
  useAutoplay(next, duration, isPaused, isUrgent, !!currentSlide);

  const { activeAnimation, transition } = getSlideAnimation(
    currentSlide,
    isUrgent,
  );

  // 2. render logic
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
