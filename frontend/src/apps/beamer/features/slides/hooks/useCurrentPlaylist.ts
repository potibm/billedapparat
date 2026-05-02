import { useMemo, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useAppConfig } from "@core/config/useConfig";
import { createLogger } from "@core/logger/logger";

const logger = createLogger("Playlist");

export const useCurrentPlaylist = () => {
  const { playlists } = useAppConfig();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const activePlaylist = useMemo(() => {
    logger.info("Available playlists:", playlists);
    logger.info("URL playlist ID:", id);
    if (!id) return playlists[0];

    const playlistId = Number.parseInt(id, 10);
    const found = playlists.find((p) => p.id === playlistId);

    return found || playlists[0];
  }, [id, playlists]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target !== document.body) return;

      const key = Number.parseInt(e.key, 10);
      if (key >= 1 && key <= 9) {
        const targetPlaylist = playlists.find((p) => p.id === key);
        if (targetPlaylist) {
          navigate(`/beamer/${targetPlaylist.id}`);
        }
      }
    };

    globalThis.addEventListener("keydown", handleKeyDown);
    return () => globalThis.removeEventListener("keydown", handleKeyDown);
  }, [playlists, navigate]);

  return activePlaylist;
};
