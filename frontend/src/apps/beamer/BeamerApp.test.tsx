/// <reference types="@testing-library/jest-dom" />
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, act, fireEvent } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { ConfigContext } from "@core/config/ConfigContext";
import { AppConfig } from "@core/config/config.schemas";
import { BeamerApp } from "./BeamerApp";
import { slideStore } from "./features/slides/store/slideStore";
import type { Slide } from "./features/slides/types/slide.schema";

// ---------------------------------------------------------------------------
// Helpers for advancing React/vi timers consistently.
// ---------------------------------------------------------------------------

/**
 * Flushes pending macrotasks (setTimeout(0)) used by the slideshow engine to
 * kick off its initial slide, plus the React microtask queue. Needed because
 * the engine uses `setTimeout(next, 0)` to start the very first NEXT action,
 * and `act(async () => await Promise.resolve())` alone does not flush
 * macrotasks.
 */
const advanceInitialTick = async () => {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(0);
  });
};

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

// framer-motion's `AnimatePresence`/`motion.div` causes spurious renders
// across the `key={currentSlide.id}` boundary in jsdom. Replace motion with a
// pass-through component that renders children into a real DOM element (we
// only care about presence, not CSS animations).
vi.mock("framer-motion", () => ({
  AnimatePresence: ({ children }: { children: React.ReactNode }) => children,
  motion: new Proxy(
    {},
    {
      get:
        () =>
        ({ children }: { children: React.ReactNode }) =>
          children,
    },
  ),
}));

// Sentry's real ErrorBoundary calls into a global client during dev. Stub it
// so integration tests don't depend on Sentry internals.
vi.mock("@sentry/react", () => ({
  ErrorBoundary: ({ children }: { children: React.ReactNode }) => children,
}));

// ---------------------------------------------------------------------------
// Mock EventSource (real one doesn't exist in jsdom)
// ---------------------------------------------------------------------------

interface MessageEventLike {
  data: string;
}
type EventCallback = (event: MessageEventLike) => void;

class MockEventSource {
  public url: string;
  public readyState = 1;
  public listeners: Record<string, EventCallback[]> = {};
  public onerror: (() => void) | null = null;

  constructor(url: string) {
    this.url = url;
  }

  addEventListener(eventName: string, callback: EventCallback) {
    (this.listeners[eventName] ??= []).push(callback);
  }

  close() {
    this.listeners = {};
  }

  emit(eventName: string, data: unknown) {
    (this.listeners[eventName] ?? []).forEach((cb) =>
      cb({ data: JSON.stringify(data) }),
    );
  }
}

let lastMockSSE: MockEventSource | null = null;
const MockEventSourceProxy = new Proxy(MockEventSource, {
  construct(_target, args) {
    const instance = new MockEventSource(...(args as [string]));
    lastMockSSE = instance;
    return instance;
  },
});

globalThis.EventSource = MockEventSourceProxy as unknown as typeof EventSource;

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const makeSlide = (
  id: number,
  type:
    "news" | "sponsor" | "social.media" | "social.text" | "timetable" = "news",
  body: string | null = null,
  is_urgent = false,
): Slide =>
  ({
    id,
    status: "active",
    created_at: "2024-01-01T00:00:00+00:00",
    content: {
      type,
      title: `Title ${id}`,
      ...(body !== null ? { body } : {}),
    },
    display_options: {
      priority: 1,
      is_urgent,
      allow_social_overlay: true,
    },
  }) as Slide;

const testConfig: AppConfig = {
  version: "1.2.3",
  environment: "test",
  sentry: {
    dsn: "",
    environment: "test",
    version: "1.2.3",
    traces_sample_rate: 0,
    replay_session_sample_rate: 0,
    replay_error_sample_rate: 0,
  },
  format: { date: { locale: "en-US", options: {} } },
  playlists: [
    {
      id: 1,
      name: "Default",
      steps: [
        {
          type: "news",
          order: "asc",
          count: 5,
          duration: 5,
        },
      ],
    },
  ],
  admin_urls: {},
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const renderBeamerAt = (path = "/beamer") =>
  render(
    <ConfigContext value={testConfig}>
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route path="/beamer/:id?" element={<BeamerApp />} />
        </Routes>
      </MemoryRouter>
    </ConfigContext>,
  );

const seedStoreWithSlides = (slides: Slide[]) => {
  // Push slides straight into the store's map to bypass the mock EventSource
  // bootstrap path — production code consumes the store the same way either
  // way, and we avoid racing with the SSE listener.
  (slideStore as unknown as { slideMap: Record<number, unknown> }).slideMap =
    {};
  (slideStore as unknown as { initSlides: (s: Slide[]) => void }).initSlides(
    slides,
  );
};

// ---------------------------------------------------------------------------
// Setup
// ---------------------------------------------------------------------------

beforeEach(() => {
  // The engine uses `setInterval(1000)` for autoplay and toast ticks, so we
  // run with fake timers throughout the suite. The teardown in tests/setup.ts
  // restores real timers between *files*, but we still need to do it ourselves
  // between *tests* in case the teardown order changes.
  vi.useFakeTimers();
  // Reset the singleton between tests so each starts from empty store.
  (slideStore as unknown as { slideMap: Record<number, unknown> }).slideMap =
    {};
  (slideStore as unknown as { slideArray: unknown[] }).slideArray = [];
  (slideStore as unknown as { listeners: Set<unknown> }).listeners = new Set();
  lastMockSSE = null;
});

afterEach(() => {
  // Disconnect any open EventSource and clean timers between tests (the
  // global setup file also calls useRealTimers + restoreAllMocks).
  slideStore.disconnect();
  vi.clearAllTimers();
  vi.useRealTimers();
});

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("BeamerApp", () => {
  it("should render the STANDBY screen when no slides are in the store", () => {
    renderBeamerAt();

    expect(screen.getByText("STANDBY")).toBeInTheDocument();
    expect(screen.getByText(/test v1\.2\.3/i)).toBeInTheDocument();
  });

  it("should render a slide once the store has slides (no SSE needed)", async () => {
    seedStoreWithSlides([makeSlide(1, "news", "Breaking news body")]);

    renderBeamerAt();

    // After the initial `setTimeout(next, NEXT_TICK_TIMEOUT=0)` the engine
    // should advance and pick slide 1.
    await advanceInitialTick();

    expect(screen.queryByText("STANDBY")).not.toBeInTheDocument();
    // The slide title text is rendered via <h1> by TextSlide.
    expect(screen.getByText("Title 1")).toBeInTheDocument();
  });

  it("should call slideStore.connect() on mount (SSE wiring)", () => {
    expect(lastMockSSE).toBeNull();

    renderBeamerAt();

    expect(lastMockSSE).not.toBeNull();
    expect(lastMockSSE!.url).toBe("/api/stream");
  });

  it("should pick up new slides delivered via the SSE INIT event", async () => {
    renderBeamerAt();

    // After mount, the MockEventSource should have listeners attached.
    expect(lastMockSSE).not.toBeNull();
    expect(lastMockSSE!.listeners.INIT).toBeDefined();

    await act(async () => {
      lastMockSSE!.emit("INIT", [makeSlide(42, "news", "via SSE")]);
      // Flush the initial-next setTimeout(0) the engine kicked off on mount.
      await vi.advanceTimersByTimeAsync(0);
    });

    expect(screen.queryByText("STANDBY")).not.toBeInTheDocument();
    expect(screen.getByText("Title 42")).toBeInTheDocument();
  });

  it("should advance to the next slide when ArrowRight is pressed", async () => {
    seedStoreWithSlides([
      makeSlide(1, "news", "First"),
      makeSlide(2, "news", "Second"),
    ]);

    renderBeamerAt();

    // Wait for initial tick
    await advanceInitialTick();

    expect(screen.getByText("Title 1")).toBeInTheDocument();

    // Dispatch keyboard event targeted at document.body (the real target the
    // hook listens on).
    await act(async () => {
      fireEvent.keyDown(document.body, { key: "ArrowRight" });
    });

    expect(screen.getByText("Title 2")).toBeInTheDocument();
    expect(screen.queryByText("Title 1")).not.toBeInTheDocument();
  });

  it("should toggle the pause state when Space is pressed (DebugOverlay reflects it)", async () => {
    seedStoreWithSlides([makeSlide(1, "news")]);

    renderBeamerAt();

    await advanceInitialTick();

    // DebugOverlay is gated on environment !== "production", so it should be
    // visible under our `environment: "test"` config.
    expect(screen.queryByText(/PAUSED/i)).toBeNull();

    await act(async () => {
      fireEvent.keyDown(document.body, { key: " " });
    });

    expect(screen.getByText(/PAUSED/i)).toBeDefined();
  });

  it("should fall back to STANDBY when the displayed slide goes inactive via SSE (regression for stuck-inactive bug)", async () => {
    // Regression test for review finding #7. The real bug scenario:
    // backend marks the currently-displayed slide `inactive`. SlideRenderer
    // returns null for inactive slides, so the user stares at a blank
    // background. Before the fix, BeamerApp's `if (!currentSlide)` branch
    // doesn't trigger (currentSlide is still the slide object — just
    // status=inactive), so STANDBY never renders.
    //
    // Fix: engineReducer increments `recoveryAttempts` on every NEXT that
    // finds no candidates. useSlideshowEngine has a watchdog that
    // dispatches RESET_PLAYLIST once the count reaches the threshold.
    // After RESET, history is empty -> currentSlide becomes null ->
    // BeamerApp renders STANDBY.
    const slide = makeSlide(1, "news", "live");
    seedStoreWithSlides([slide]);

    renderBeamerAt();
    await advanceInitialTick();

    // Confirm we are NOT on STANDBY yet: a slide is being displayed.
    expect(screen.queryByText("STANDBY")).toBeNull();

    // Mark the displayed slide inactive via SSE UPDATE. This is the
    // scenario the watchdog was designed for: the slide still exists,
    // it still sits in history, but rendering it produces nothing.
    const inactiveSlide = {
      ...slide,
      status: "inactive",
    } as Slide;
    await act(async () => {
      lastMockSSE!.emit("UPDATE", inactiveSlide);
      await vi.advanceTimersByTimeAsync(0);
    });

    // Without the watchdog, currentSlide is still set (slide exists),
    // so we render an empty frame from SlideRenderer returning null for
    // inactive slides. With the watchdog, repeated NEXTs bump the
    // recovery counter and trip RESET_PLAYLIST, clearing history.
    // Trigger enough NEXTS to exceed STUCK_THRESHOLD.
    for (let i = 0; i < 10; i += 1) {
      await act(async () => {
        fireEvent.keyDown(document.body, { key: "ArrowRight" });
      });
    }

    expect(screen.getByText("STANDBY")).toBeInTheDocument();
  });

  it("should report each slide's own step duration (not the next step's) in the engine", async () => {
    // Regression test for the social-media 1-second-display bug.
    //
    // Playlist: [social.media: 2s, news: 10s]
    // Bug:      useSlideshowEngine derived `duration` from
    //           `state.stepIndex`, which is the *next* step to search from,
    //           so social.media displayed for 10s (news's duration) and news
    //           displayed for 2s (social.media's duration).
    // Fix:      `currentStep` now reads from `state.displayedStepInfo.stepIndex`,
    //           which is the step the displayed slide was actually picked from.
    //
    // We assert the DebugOverlay's `Xs` text matches the displayed slide's
    // own step duration.
    const debugConfig: AppConfig = {
      ...testConfig,
      playlists: [
        {
          id: 1,
          name: "Social Media",
          steps: [
            { type: "social.media", order: "asc", count: 1, duration: 2 },
            { type: "news", order: "asc", count: 1, duration: 10 },
          ],
        },
      ],
    };

    const socialSlide = makeSlide(100, "social.media", "insta post");
    const newsSlide = makeSlide(200, "news", "breaking news");
    seedStoreWithSlides([socialSlide, newsSlide]);

    render(
      <ConfigContext value={debugConfig}>
        <MemoryRouter initialEntries={["/beamer"]}>
          <Routes>
            <Route path="/beamer/:id?" element={<BeamerApp />} />
          </Routes>
        </MemoryRouter>
      </ConfigContext>,
    );

    await advanceInitialTick();

    // First slide to display is social.media (step 0, duration 2s).
    // SocialSlide renders the body, not the title, so we assert via the body.
    expect(screen.getByText("insta post")).toBeInTheDocument();
    expect(screen.getByText(/2s/)).toBeInTheDocument();

    // Advance to the news slide (step 1, duration 10s).
    await act(async () => {
      fireEvent.keyDown(document.body, { key: "ArrowRight" });
    });

    // NewsSlide renders the title as <h1>.
    expect(screen.getByText("Title 200")).toBeInTheDocument();
    expect(screen.getByText(/10s/)).toBeInTheDocument();
  });

  it("should reset the autoplay timer when advancing manually via keyboard", async () => {
    // Regression test for the "manual next shows slide too briefly" bug.
    //
    // Bug: useAutoplay used `setInterval` keyed on `[duration, isPaused,
    // isUrgent, hasCurrentSlide]`. When the user pressed ArrowRight late
    // in the slide's display window, the reducer advanced `currentSlide`
    // but the deps didn't change (duration is the same, paused/urgent/
    // current unchanged), so the cleanup never ran. The new slide only
    // displayed for the *remaining* time on the original interval.
    //
    // Fix: useAutoplay now uses `setTimeout` and depends on a `resetKey`
    // (the current slide id), so the timer re-arms on every slide change.
    //
    // Test: step duration is 5s. We let 4s of auto-advance time elapse,
    // then press ArrowRight (simulating a user manually advancing late in
    // the cycle). With the bug, the original setInterval would fire just
    // 1s later, replacing the new slide. With the fix, slide 2 must still
    // be on screen 1.5s after the manual advance.
    seedStoreWithSlides([
      makeSlide(1, "news", "First"),
      makeSlide(2, "news", "Second"),
    ]);

    renderBeamerAt();
    await advanceInitialTick();

    expect(screen.getByText("Title 1")).toBeInTheDocument();

    // Burn 4s of the 5s duration so the original (un-reset) interval is
    // about to fire. This is where the bug manifests: the manual advance
    // happens with only 1s remaining on the old timer.
    await act(async () => {
      await vi.advanceTimersByTimeAsync(4000);
    });
    expect(screen.getByText("Title 1")).toBeInTheDocument();

    // Manual advance to slide 2 with only 1s remaining on the original
    // interval.
    await act(async () => {
      fireEvent.keyDown(document.body, { key: "ArrowRight" });
    });
    expect(screen.getByText("Title 2")).toBeInTheDocument();

    // Advance 1.5s. With the bug, the original interval fires at +1s
    // (total +5s from initial display) and replaces slide 2. With the
    // fix, the timer was reset on the manual advance, so slide 2 still
    // shows for another 3.5s.
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1500);
    });
    expect(screen.getByText("Title 2")).toBeInTheDocument();
    expect(screen.queryByText("Title 1")).not.toBeInTheDocument();
  });
});
