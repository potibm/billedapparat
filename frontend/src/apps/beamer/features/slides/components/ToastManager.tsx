import { useToastEngine } from "../hooks/useToastEngine";
import { ToastOverlay } from "./ToastOverlay";
import { Slide } from "../types/slide.schema";

interface ToastManagerProps {
  toastSlides: Slide[];
  allowOverlay: boolean;
}

export const ToastManager = ({
  toastSlides,
  allowOverlay,
}: ToastManagerProps) => {
  const activeToasts = useToastEngine(toastSlides, allowOverlay);

  return <ToastOverlay toasts={activeToasts} />;
};
