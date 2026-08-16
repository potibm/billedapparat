import { useMemo } from "react";
import { useParams } from "react-router-dom";
import { useAppConfig } from "@core/config/useConfig";
import { createLogger } from "@core/logger/logger";

const logger = createLogger("Playlist");

export const useCurrentPlaylist = () => {
  const { playlists } = useAppConfig();
  const { id } = useParams<{ id: string }>();

  const activePlaylist = useMemo(() => {
    if (!playlists || playlists.length === 0) {
      logger.warn("No playlists available in config");
      return null;
    }

    if (!id) return playlists[0];

    const playlistId = Number.parseInt(id, 10);
    const found = playlists.find((p) => p.id === playlistId);

    if (!found) {
      logger.warn(
        `Playlist ID ${id} not found, falling back to first available`,
      );
      return playlists[0];
    }

    return found;
  }, [id, playlists]);

  return activePlaylist;
};
