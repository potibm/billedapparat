import { describe, expect, it, test, vi } from "vitest";
import * as logic from "./slideshow.logic";
import { Slide } from "../types/slide.schema";
import { PlaylistStep } from "@core/config/config.schemas";

const mockSlides: Slide[] = [
  {
    id: 1,
    created_at: "2026-05-01T10:00:00Z",
    display_options: { priority: 1 },
  } as Slide,
  {
    id: 2,
    created_at: "2026-05-01T11:00:00Z",
    display_options: { priority: 10 },
  } as Slide,
  {
    id: 3,
    created_at: "2026-05-01T09:00:00Z",
    display_options: { priority: 1 },
  } as Slide,
];

const { selectNextSlide, sortSlides, findNextValidStep, pickWeighted } = logic;

describe("Slideshow Logic Utils", () => {
  describe("sortSlides", () => {
    it("should sort slides by date ascending (asc)", () => {
      const sorted = sortSlides(mockSlides, "asc");
      expect(sorted[0].id).toBe(3); // 09:00
      expect(sorted[2].id).toBe(2); // 11:00
    });

    it("should sort slides by date descending (desc)", () => {
      const sorted = sortSlides(mockSlides, "desc");
      expect(sorted[0].id).toBe(2); // 11:00
      expect(sorted[2].id).toBe(3); // 09:00
    });
  });

  describe("pickWeighted", () => {
    test("pickWeighted returns the only item", () => {
      const items = [{ id: 1, w: 10 }];
      const result = pickWeighted(items, (i) => i.w);
      expect(result?.id).toBe(1);
    });

    it("should return null for an empty array", () => {
      expect(pickWeighted([], (_i) => 1)).toBeNull();
    });

    test("pickWeighted respects high priority", () => {
      const items = [
        { id: "rare", w: 1 },
        { id: "often", w: 999 },
      ];
      vi.spyOn(Math, "random").mockReturnValue(0.99);
      const result = pickWeighted(items, (i) => i.w);
      expect(result?.id).toBe("often");
      vi.restoreAllMocks();
    });
  });

  describe("findNextValidStep", () => {
    const steps: PlaylistStep[] = [
      { type: "sponsor", count: 1, order: "random", duration: 10 },
      { type: "news", count: 1, order: "asc", duration: 10 },
      { type: "empty", count: 1, order: "asc", duration: 10 },
    ];

    it("should return the current step if slides exist", () => {
      const getSlides = (type: string) =>
        type === "sponsor" ? [mockSlides[0]] : [];
      const result = findNextValidStep(steps, 0, getSlides);

      expect(result?.index).toBe(0);
      expect(result?.step.type).toBe("sponsor");
    });

    it("should skip empty steps and find the next one with slides", () => {
      const getSlides = (type: string) =>
        type === "news" ? [mockSlides[0]] : [];
      // we are starting at index 0 (sponsor), but only 'news' has slides
      const result = findNextValidStep(steps, 0, getSlides);

      expect(result?.index).toBe(1); // 'news' is index 1
      expect(result?.step.type).toBe("news");
    });

    it("should return null if no step has any slides", () => {
      const getSlides = () => []; // no slides for any type
      const result = findNextValidStep(steps, 0, getSlides);
      expect(result).toBeNull();
    });
  });

  describe("selectNextSlide", () => {
    const mockCandidates: Slide[] = [
      { id: 10, display_options: { priority: 1 } } as Slide,
      { id: 20, display_options: { priority: 1 } } as Slide,
    ];

    it("should return null if no candidates are provided", () => {
      const step = { order: "asc" } as PlaylistStep;
      expect(selectNextSlide([], step, 0, undefined)).toBeNull();
    });

    it("should select by pointer index for non-random orders", () => {
      const step = { order: "asc" } as PlaylistStep;
      // pointer 1 should select the second element (ID 20)
      const result = selectNextSlide(mockCandidates, step, 1, undefined);
      expect(result?.id).toBe(20);
    });

    it("should return a valid candidate when order is random", () => {
      const step = { order: "random" } as PlaylistStep;
      const result = selectNextSlide(mockCandidates, step, 0, undefined);

      // it should be one of the candidates (ID 10 or 20)
      expect([10, 20]).toContain(result?.id);
    });

    it("should filter out the lastId in random mode to prevent direct repetition", () => {
      const step = { order: "random" } as PlaylistStep;

      // we will run the random selection multiple times to check that ID 10 is never returned when it is lastId
      for (let i = 0; i < 100; i++) {
        const result = selectNextSlide(mockCandidates, step, 0, 10);
        expect(result?.id).toBe(20); // Es MUSS 20 sein
      }
    });

    it("should fallback to the same slide if it is the only candidate even if it was lastId", () => {
      const step = { order: "random" } as PlaylistStep;
      const singleCandidate = [
        { id: 10, display_options: { priority: 1 } } as Slide,
      ];

      // when only ID 10 is available, it must be selected even if it was lastId
      const result = selectNextSlide(singleCandidate, step, 0, 10);
      expect(result?.id).toBe(10);
    });
  });
});
