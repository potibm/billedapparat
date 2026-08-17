import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { ConfigContext } from "@core/config/ConfigContext";
import { useCurrentPlaylist } from "./useCurrentPlaylist";
import type { Slide } from "../types/slide.schema";

// Mock useSlideManager so we can control the urgent-slides list.
vi.mock("./useSlideManager", () => ({
  useSlideManager: vi.fn(),
}));

import { useSlideManager } from "./useSlideManager";
const mockUseSlideManager = vi.mocked(useSlideManager);

const makeUrgentSlide = (id: number): Slide =>
  ({
    id,
    status: "active",
    created_at: "2024-01-01T00:00:00+00:00",
    content: { type: "news", title: "urgent" },
    display_options: {
      priority: 1,
      is_urgent: true,
      allow_social_overlay: false,
    },
  }) as Slide;

const makeNonUrgentSlide = (id: number): Slide =>
  ({
    id,
    status: "active",
    created_at: "2024-01-01T00:00:00+00:00",
    content: { type: "sponsor" },
    display_options: {
      priority: 1,
      is_urgent: false,
      allow_social_overlay: false,
    },
  }) as Slide;

const baseConfig = {
  playlists: [
    {
      id: 1,
      name: "Default",
      steps: [{ type: "news", count: 1, order: "asc", duration: 10 }],
    },
    {
      id: 2,
      name: "Alt",
      steps: [{ type: "sponsor", count: 1, order: "asc", duration: 10 }],
    },
  ],
} as Parameters<typeof ConfigContext.Provider>[0]["value"];

const renderUseCurrentPlaylist = (path: string) =>
  renderHook(() => useCurrentPlaylist(), {
    wrapper: ({ children }) => (
      <ConfigContext value={baseConfig}>
        <MemoryRouter initialEntries={[path]}>
          <Routes>
            <Route path="/beamer/:id?" element={<>{children}</>} />
          </Routes>
        </MemoryRouter>
      </ConfigContext>
    ),
  });

beforeEach(() => {
  mockUseSlideManager.mockReset();
});

describe("useCurrentPlaylist", () => {
  it("should return the first playlist when no id is provided", () => {
    mockUseSlideManager.mockReturnValue({
      slides: [],
      getByType: vi.fn(() => []),
      getUrgent: vi.fn(() => []),
    });

    const { result } = renderUseCurrentPlaylist("/beamer");

    expect(result.current?.id).toBe(1);
    expect(result.current?.name).toBe("Default");
  });

  it("should look up the playlist matching the URL id", () => {
    mockUseSlideManager.mockReturnValue({
      slides: [],
      getByType: vi.fn(() => []),
      getUrgent: vi.fn(() => []),
    });

    const { result } = renderUseCurrentPlaylist("/beamer/2");

    expect(result.current?.id).toBe(2);
  });

  it("should fall back to the first playlist when the id does not match", () => {
    mockUseSlideManager.mockReturnValue({
      slides: [],
      getByType: vi.fn(() => []),
      getUrgent: vi.fn(() => []),
    });

    const { result } = renderUseCurrentPlaylist("/beamer/999");

    expect(result.current?.id).toBe(1);
  });

  it("should return null when no playlists are configured", () => {
    mockUseSlideManager.mockReturnValue({
      slides: [],
      getByType: vi.fn(() => []),
      getUrgent: vi.fn(() => []),
    });

    const { result } = renderHook(() => useCurrentPlaylist(), {
      wrapper: ({ children }) => (
        <ConfigContext value={{ playlists: [] } as never}>
          <MemoryRouter initialEntries={["/beamer"]}>
            <Routes>
              <Route path="/beamer/:id?" element={<>{children}</>} />
            </Routes>
          </MemoryRouter>
        </ConfigContext>
      ),
    });

    expect(result.current).toBeNull();
  });

  it("should return a synthetic Urgent Override playlist when urgent slides are present", () => {
    mockUseSlideManager.mockReturnValue({
      slides: [makeNonUrgentSlide(1), makeUrgentSlide(42)],
      getByType: vi.fn(() => []),
      getUrgent: vi.fn(() => [makeUrgentSlide(42)]),
    });

    const { result } = renderUseCurrentPlaylist("/beamer/2"); // <-- normally would return playlist 2

    expect(result.current).toEqual({
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
    });
  });

  it("should override the default playlist when urgent slides are present and no id is given", () => {
    mockUseSlideManager.mockReturnValue({
      slides: [makeUrgentSlide(7)],
      getByType: vi.fn(() => []),
      getUrgent: vi.fn(() => [makeUrgentSlide(7)]),
    });

    const { result } = renderUseCurrentPlaylist("/beamer");

    expect(result.current?.id).toBe(-1);
    expect(result.current?.name).toBe("Urgent Override");
    expect(result.current?.steps[0].type).toBe("urgent");
  });

  it("should not override when no urgent slides are present", () => {
    const getUrgent = vi.fn(() => []);
    mockUseSlideManager.mockReturnValue({
      slides: [makeNonUrgentSlide(1)],
      getByType: vi.fn(() => []),
      getUrgent,
    });

    const { result } = renderUseCurrentPlaylist("/beamer/2");

    expect(getUrgent).toHaveBeenCalled();
    expect(result.current?.id).toBe(2); // normal playlist, not overridden
  });
});
