import { useMemo } from "react";
import { useParams } from "react-router-dom";
import { useAppConfig } from "@core/config/useConfig";
import { createLogger } from "@core/logger/logger";
import { useSlideManager } from "./useSlideManager";

const logger = createLogger("Playlist");

export const useCurrentPlaylist = () => {
  const { playlists } = useAppConfig();
  const { id } = useParams<{ id: string }>();

  const { getUrgent } = useSlideManager();
  const urgentSlides = getUrgent();
  logger.debug("Urgent slides", "count", urgentSlides.length);

  const regularPlaylist = useMemo(() => {
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

  if (urgentSlides.length > 0) {
    logger.info("Urgent slides active! Intercepting normal playlist.");

    return {
      id: -1,
      name: "Urgent Override",
      steps: [
        {
          type: "urgent",
          count: 1,
          order: "desc",
          duration: 10,
        },
      ],
    };
  }

  return regularPlaylist;
};
