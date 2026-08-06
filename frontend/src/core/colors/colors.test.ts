import { describe, it, expect } from "vitest";
import { getContrastTextColor } from "./colors"; // adjust path if necessary

describe("getContrastTextColor", () => {
  it("should return dark text (text-gray-900) for light colors", () => {
    // White
    expect(getContrastTextColor("#FFFFFF")).toBe("text-gray-900");
    // Light Yellow
    expect(getContrastTextColor("#FFFF00")).toBe("text-gray-900");
    // Light Gray
    expect(getContrastTextColor("#D3D3D3")).toBe("text-gray-900");
  });

  it("should return light text (text-white) for dark colors", () => {
    // Black
    expect(getContrastTextColor("#000000")).toBe("text-white");
    // Dark Blue
    expect(getContrastTextColor("#00008B")).toBe("text-white");
    // Dark Green
    expect(getContrastTextColor("#006400")).toBe("text-white");
  });

  it("should correctly process short hex codes (3 characters)", () => {
    // Short White (#FFF) -> text-gray-900
    expect(getContrastTextColor("#FFF")).toBe("text-gray-900");
    // Short Black (#000) -> text-white
    expect(getContrastTextColor("#000")).toBe("text-white");
  });

  it("should work even if the hash symbol (#) is missing", () => {
    expect(getContrastTextColor("FFFFFF")).toBe("text-gray-900");
    expect(getContrastTextColor("000000")).toBe("text-white");
    expect(getContrastTextColor("FFF")).toBe("text-gray-900");
  });

  it('should return "text-white" as a fallback for invalid inputs', () => {
    // Too short
    expect(getContrastTextColor("#12")).toBe("text-white");
    // Too long
    expect(getContrastTextColor("#1234567")).toBe("text-white");
    // Completely empty
    expect(getContrastTextColor("")).toBe("text-white");
    // Not a hex code
    expect(getContrastTextColor("red")).toBe("text-white");
  });
});
