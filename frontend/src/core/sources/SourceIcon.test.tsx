import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { SourceIcon } from "./SourceIcon";

describe("SourceIcon", () => {
  describe("react-icons rendering", () => {
    it("should render mastodon icon", () => {
      const { container } = render(<SourceIcon source="mastodon" />);
      const svg = container.querySelector("svg");
      expect(svg).toBeInTheDocument();
    });

    it("should render bluesky icon", () => {
      const { container } = render(<SourceIcon source="bluesky" />);
      const svg = container.querySelector("svg");
      expect(svg).toBeInTheDocument();
    });

    it("should render twitch icon", () => {
      const { container } = render(<SourceIcon source="twitch" />);
      const svg = container.querySelector("svg");
      expect(svg).toBeInTheDocument();
    });

    it("should render discord icon", () => {
      const { container } = render(<SourceIcon source="discord" />);
      const svg = container.querySelector("svg");
      expect(svg).toBeInTheDocument();
    });

    it("should render pouet icon", () => {
      const { container } = render(<SourceIcon source="pouet" />);
      const svg = container.querySelector("svg");
      expect(svg).toBeInTheDocument();
    });
  });

  describe("custom icon fallback", () => {
    it("should render custom icon when provided for unknown source", () => {
      const customSvg = <svg data-testid="custom-svg" />;
      render(<SourceIcon source="unknown" customIcon={customSvg} />);
      expect(screen.getByTestId("custom-svg")).toBeInTheDocument();
    });

    it("should prefer react-icon over custom icon", () => {
      const customSvg = <svg data-testid="custom-svg" />;
      const { container } = render(
        <SourceIcon source="mastodon" customIcon={customSvg} />,
      );
      // Should render the react-icon (SVG from react-icons), not the custom one
      expect(screen.queryByTestId("custom-svg")).not.toBeInTheDocument();
      expect(container.querySelector("svg")).toBeInTheDocument();
    });
  });

  describe("text fallback", () => {
    it("should render text for unknown source", () => {
      render(<SourceIcon source="facebook" />);
      expect(screen.getByText("facebook")).toBeInTheDocument();
    });

    it("should render nothing for null source", () => {
      const { container } = render(<SourceIcon source={null} />);
      expect(container.firstChild).toBeNull();
    });

    it("should render nothing for undefined source", () => {
      const { container } = render(<SourceIcon source={undefined} />);
      expect(container.firstChild).toBeNull();
    });
  });

  describe("styling", () => {
    it("should apply custom size to react-icon", () => {
      const { container } = render(<SourceIcon source="mastodon" width={24} />);
      const svg = container.querySelector("svg");
      expect(svg).toHaveAttribute("width", "24");
      expect(svg).toHaveAttribute("height", "24");
    });

    it("should apply className to react-icon", () => {
      const { container } = render(
        <SourceIcon source="mastodon" className="custom-class" />,
      );
      const svg = container.querySelector("svg");
      expect(svg).toHaveClass("custom-class");
    });

    it("should apply className to text fallback", () => {
      render(<SourceIcon source="facebook" className="custom-class" />);
      const span = screen.getByText("facebook");
      expect(span).toHaveClass("custom-class");
    });
  });
});
