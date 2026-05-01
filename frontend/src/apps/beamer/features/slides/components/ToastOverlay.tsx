import { AnimatePresence, motion } from "framer-motion";
import { Slide } from "../types/slide.schema";
import { FormattedText } from "./FormattedText";
import { AuthorHeader } from "./ui/AuthorHeader";

interface ToastOverlayProps {
  toasts: Slide[];
}

export const ToastOverlay = ({ toasts }: ToastOverlayProps) => {
  return (
    <div className="toast-overlay absolute bottom-12 right-12 z-50 w-[450px] flex flex-col gap-4 pointer-events-none overflow-hidden">
      <AnimatePresence mode="popLayout">
        {toasts.map((toast) => (
          <motion.div
            layout
            key={toast.id}
            initial={{ opacity: 0, x: 100, scale: 0.95 }}
            animate={{ opacity: 1, x: 0, scale: 1 }}
            exit={{ opacity: 0, x: 100, scale: 0.95 }}
            transition={{ duration: 0.5, type: "spring", bounce: 0.2 }}
            className="toast-overlay__item bg-black/60 backdrop-blur-md rounded-2xl shadow-2xl p-6 pointer-events-auto border border-gray-100"
          >
            <AuthorHeader
              displayName={toast.author?.display_name}
              username={toast.author?.username}
              avatarUrl={toast.author?.avatar?.local_url}
              className="toast-overlay__header mb-3"
            />

            <div className="toast-overlay__body text-white text-lg leading-snug whitespace-pre-wrap">
              <FormattedText text={toast.content.body} />
            </div>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  );
};
