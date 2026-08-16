import { Slide } from "./slide.schema";
import { StepInfo } from "./playlist.types";

export interface SlideshowEngine {
  /** Currently active slide, that is going to be displayed */
  currentSlide: Slide | null;
  /** Function to move to the next slide (considers playlist logic) */
  next: () => void;
  /** Function to go back in the history */
  previous: () => void;
  /** Function to pause or resume the automatic slide change */
  togglePause: () => void;
  /** Current pause status */
  isPaused: boolean;
  /** Flag indicating whether the current slide is an urgent slide */
  isUrgent: boolean;
  /** Flag indicating the current slide allows overlays with toasts  */
  allowOverlay: boolean;
  /** List of slides to be displayed as overlays/toasts */
  toastSlides: Slide[];
  /** The display duration of the current slide in seconds */
  duration: number;
  /** Metadata about the current step in the playlist */
  stepInfo: StepInfo | null;
}
