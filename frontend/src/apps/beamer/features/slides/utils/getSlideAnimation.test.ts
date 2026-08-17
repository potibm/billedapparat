import { describe, it, expect, vi, afterEach } from "vitest";
import { getSlideAnimation } from "./getSlideAnimation";
import * as animationsModule from "../../animations/types/animations.schemas";
import type { Slide } from "../types/slide.schema";

const makeSlide = (id: number): Slide =>
  ({
    id,
    status: "active",
    created_at: "",
    content: {},
    display_options: {
      priority: 0,
      is_urgent: false,
      allow_social_overlay: false,
    },
  }) as unknown as Slide;

afterEach(() => {
  vi.restoreAllMocks();
});

describe("getSlideAnimation", () => {
  it("should return the fade fallback when there is no current slide", () => {
    const result = getSlideAnimation(null, false);

    expect(result.activeAnimation).toBe("fade");
    expect(result.transition).toEqual({ duration: 0.8, ease: "easeInOut" });
  });

  it("should return the urgent animation when isUrgent is true (regardless of slide)", () => {
    const result = getSlideAnimation(makeSlide(1), true);

    expect(result.activeAnimation).toBe("urgent");
    expect(result.transition).toEqual({ duration: 0.2, ease: "easeOut" });
  });

  it("should pick an animation deterministically based on slide id", () => {
    // Same input -> same output
    const slide = makeSlide(42);
    const a = getSlideAnimation(slide, false);
    const b = getSlideAnimation(slide, false);

    expect(a.activeAnimation).toBe(b.activeAnimation);
  });

  it("should always return a key that exists in the registered animations map", () => {
    const slide = makeSlide(1);
    const result = getSlideAnimation(slide, false);

    expect(result.activeAnimation in animationsModule.animations).toBe(true);
  });

  it("should return the fade fallback if the animations map is empty", () => {
    vi.spyOn(animationsModule, "animations", "get").mockReturnValue(
      {} as unknown as typeof animationsModule.animations,
    );

    const result = getSlideAnimation(makeSlide(1), false);

    expect(result.activeAnimation).toBe("fade");
    expect(result.transition).toEqual({ duration: 0.8, ease: "easeInOut" });
  });

  it("should still return the urgent branch even if animations is empty", () => {
    vi.spyOn(animationsModule, "animations", "get").mockReturnValue(
      {} as unknown as typeof animationsModule.animations,
    );

    const result = getSlideAnimation(makeSlide(1), true);

    expect(result.activeAnimation).toBe("urgent");
  });
});
