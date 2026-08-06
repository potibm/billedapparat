// Calculate whether text should be dark or light based on background color
export const getContrastTextColor = (hexColor: string): string => {
  let hex = hexColor.replace("#", "");

  if (hex.length === 3) {
    hex = hex
      .split("")
      .map((char) => char + char)
      .join("");
  }

  if (hex.length !== 6) return "text-white";

  const r = Number.parseInt(hex.substring(0, 2), 16);
  const g = Number.parseInt(hex.substring(2, 4), 16);
  const b = Number.parseInt(hex.substring(4, 6), 16);

  // W3C brightness formula (YIQ)
  const brightness = (r * 299 + g * 587 + b * 114) / 1000;

  // If brightness value is above 128, the color is light -> dark text
  return brightness > 128 ? "text-gray-900" : "text-white";
};
