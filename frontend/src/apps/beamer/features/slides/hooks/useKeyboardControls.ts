import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useAppConfig } from "@core/config/useConfig";

export function useKeyboardControls(
  next: () => void,
  previous: () => void,
  togglePause: () => void,
) {
  const { playlists } = useAppConfig();
  const navigate = useNavigate();

  const stateRef = useRef({ next, previous, togglePause, playlists, navigate });

  useEffect(() => {
    stateRef.current = { next, previous, togglePause, playlists, navigate };
  }, [next, previous, togglePause, playlists, navigate]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target !== document.body) return;

      const {
        next: currentNext,
        previous: currentPrev,
        togglePause: currentToggle,
        playlists: currentPlaylists,
        navigate: currentNavigate,
      } = stateRef.current;

      // 1. Playback controls
      if (e.key === "ArrowRight") currentNext();
      if (e.key === "ArrowLeft") currentPrev();
      if (e.key === " " || e.key === "Spacebar") {
        e.preventDefault();
        currentToggle();
      }

      // 2. Playlist shortcuts
      const keyNumber = Number.parseInt(e.key, 10);
      if (keyNumber >= 1 && keyNumber <= 9) {
        const targetPlaylist = currentPlaylists.find((p) => p.id === keyNumber);
        if (targetPlaylist) {
          currentNavigate(`/beamer/${targetPlaylist.id}`);
        }
      }
    };

    globalThis.addEventListener("keydown", handleKeyDown);
    return () => globalThis.removeEventListener("keydown", handleKeyDown);
  }, []);
}
