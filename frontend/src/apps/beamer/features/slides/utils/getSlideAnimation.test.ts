import { describe, it, expect } from "vitest";
import { getSlideAnimation } from "./getSlideAnimation";
import * as animationsModule from "../../animations/types/animations.schemas";
import type { AnimationType } from "../../animations/types/animations.schemas";
import type { Slide } from "../types/slide.schema";

const ALL_ANIMATIONS: AnimationType[] = [
  "fade",
  "slideRight",
  "zoomIn",
  "flip",
  "urgent",
];

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

  it("should return fade when all animations are allowed none is urgent", () => {
    const result = getSlideAnimation(makeSlide(1), false, ALL_ANIMATIONS);

    expect(ALL_ANIMATIONS).toContain(result.activeAnimation);
  });

  it("should restrict non-urgent slides to the allowed animations", () => {
    const result = getSlideAnimation(makeSlide(1), false, ["fade"]);

    expect(result.activeAnimation).toBe("fade");
  });

  it("should fall back to fade when the allowed list is empty", () => {
    const result = getSlideAnimation(makeSlide(1), false, []);

    expect(result.activeAnimation).toBe("fade");
    expect(result.transition).toEqual({ duration: 0.8, ease: "easeInOut" });
  });

  it("should fall back to fade for urgent slides when urgent is not in the allowed list", () => {
    const result = getSlideAnimation(makeSlide(1), true, ["fade", "zoomIn"]);

    expect(["fade", "zoomIn"]).toContain(result.activeAnimation);
    expect(result.activeAnimation).not.toBe("urgent");
  });

  it("should still return urgent when isUrgent and urgent is allowed", () => {
    const result = getSlideAnimation(makeSlide(1), true, [
      "fade",
      "urgent",
      "zoomIn",
    ]);

    expect(result.activeAnimation).toBe("urgent");
    expect(result.transition).toEqual({ duration: 0.2, ease: "easeOut" });
  });

  it("should ignore unknown keys in the allowed list", () => {
    // The schema rejects unknown keys upstream, but the function must be
    // defensive in case it ever receives a stray value.
    const result = getSlideAnimation(makeSlide(1), false, [
      "fade",
      "unknown-anim",
    ] as unknown as AnimationType[]);

    expect(result.activeAnimation).toBe("fade");
  });
});
