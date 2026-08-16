import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAppConfig } from "@core/config/useConfig";

export function useKeyboardControls(
  next: () => void,
  previous: () => void,
  togglePause: () => void,
) {
  const { playlists } = useAppConfig();
  const navigate = useNavigate();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target !== document.body) return;

      // 1. Playback controls
      if (e.key === "ArrowRight") next();
      if (e.key === "ArrowLeft") previous();
      if (e.key === " " || e.key === "Spacebar") {
        e.preventDefault();
        togglePause();
      }

      // 2. Playlist shortcuts
      const keyNumber = Number.parseInt(e.key, 10);
      if (keyNumber >= 1 && keyNumber <= 9) {
        const targetPlaylist = playlists.find((p) => p.id === keyNumber);
        if (targetPlaylist) {
          navigate(`/beamer/${targetPlaylist.id}`);
        }
      }
    };

    globalThis.addEventListener("keydown", handleKeyDown);
    return () => globalThis.removeEventListener("keydown", handleKeyDown);
  }, [next, previous, togglePause, playlists, navigate]);
}
