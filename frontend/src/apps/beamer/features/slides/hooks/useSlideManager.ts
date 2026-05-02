import { useEffect, useState, useCallback, useMemo } from "react";
import { z } from "zod";
import { slideSchema, type Slide } from "../types/slide.schema";
import { createLogger } from "@core/logger/logger";

type SlideDictionary = Record<number, Slide>;

const logger = createLogger("Slides");

export const useSlideManager = () => {
  const [slideMap, setSlideMap] = useState<SlideDictionary>({});

  const initSlides = useCallback((slides: Slide[]) => {
    const newMap: SlideDictionary = {};
    slides.forEach((s) => (newMap[s.id] = s));
    setSlideMap(newMap);
  }, []);

  const upsertSlide = useCallback((slide: Slide) => {
    setSlideMap((prev) => ({ ...prev, [slide.id]: slide }));
  }, []);

  const deleteSlide = useCallback((id: number) => {
    setSlideMap((prev) => {
      const newMap = { ...prev };
      delete newMap[id];
      return newMap;
    });
  }, []);

  useEffect(() => {
    const evtSource = new EventSource("/api/stream");

    const handleInit = (e: MessageEvent) => {
      const parsed = z.array(slideSchema).safeParse(JSON.parse(e.data));
      logger.debug(
        "Received INIT event",
        "count",
        parsed.data?.length,
        "rawData",
        e.data,
      );
      if (parsed.error) {
        logger.warn(
          "Failed to parse INIT event",
          "error",
          parsed.error,
          "rawData",
          e.data,
        );
      }
      if (parsed.success) initSlides(parsed.data);
    };

    const handleCreate = (e: MessageEvent) => {
      const parsed = slideSchema.safeParse(JSON.parse(e.data));
      logger.debug("Received CREATE event", "id", parsed.data?.id);
      if (parsed.error) {
        logger.warn(
          "Failed to parse CREATE event",
          "error",
          parsed.error,
          "rawData",
          e.data,
        );
      }
      if (parsed.success) upsertSlide(parsed.data);
    };

    const handleUpdate = (e: MessageEvent) => {
      const parsed = slideSchema.safeParse(JSON.parse(e.data));
      logger.debug("Received UPDATE event", "id", parsed.data?.id);
      if (parsed.success) upsertSlide(parsed.data);
    };

    const handleDelete = (e: MessageEvent) => {
      const id = z.number().parse(JSON.parse(e.data));
      logger.debug("Received DELETE event", "id", id);
      deleteSlide(id);
    };

    evtSource.addEventListener("INIT", handleInit);
    evtSource.addEventListener("CREATE", handleCreate);
    evtSource.addEventListener("UPDATE", handleUpdate);
    evtSource.addEventListener("DELETE", handleDelete);

    return () => {
      evtSource.removeEventListener("INIT", handleInit);
      evtSource.removeEventListener("CREATE", handleCreate);
      evtSource.removeEventListener("UPDATE", handleUpdate);
      evtSource.removeEventListener("DELETE", handleDelete);
      evtSource.close();
    };
  }, [initSlides, upsertSlide, deleteSlide]);

  const allSlides = useMemo(() => Object.values(slideMap), [slideMap]);

  const getById = useCallback(
    (id: number) => {
      return slideMap[id] || null;
    },
    [slideMap],
  );

  const getByType = useCallback(
    (type: string) => {
      return allSlides
        .filter((s) => s.content.type === type && s.status === "active")
        .sort((a, b) => {
          const prioA = Number(a.display_options?.priority || 0);
          const prioB = Number(b.display_options?.priority || 0);
          return prioB - prioA;
        });
    },
    [allSlides],
  );

  const getUrgent = useCallback(() => {
    return allSlides
      .filter((s) => s.content.type === "urgent" && s.status === "active")
      .sort((a, b) => {
        const prioA = Number(a.display_options?.priority || 0);
        const prioB = Number(b.display_options?.priority || 0);
        return prioB - prioA;
      });
  }, [allSlides]);

  return {
    slides: allSlides,
    getById,
    getByType,
    getUrgent,
  };
};
