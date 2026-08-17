import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { ConfigContext } from "@core/config/ConfigContext";
import { useKeyboardControls } from "./useKeyboardControls";
import type { ReactNode } from "react";

const playlists = [
  { id: 1, name: "First", steps: [] },
  { id: 2, name: "Second", steps: [] },
  { id: 9, name: "Ninth", steps: [] },
];

// We only provide the `playlists` field; the rest is irrelevant for the
// keypress handling under test. `as never` skips structural typing on the
// rest of the AppConfig schema.
const baseConfig = { playlists } as never;

/**
 * Renders the hook inside a MemoryRouter + ConfigContext and exposes a
 * `getPathname()` accessor via the location observer. The hook and the
 * observer share the same `MemoryRouter`, so navigation triggered by the
 * keyboard listener is observable through `getPathname()`.
 *
 * Implementation note: `MemoryRouter` is keyed on its current location and
 * the `routes` declared beneath it. We cannot easily mount two separate
 * `renderHook` instances under one router (each would create its own).
 * Instead we rely on the hook's `useEffect` running on mount; the listener
 * is registered at module / document level so subsequent `pressKey` calls
 * take effect, even though their effect happens to be in a different
 * memoized render from the same hook instance.
 */
function setup(
  opts: {
    initialPath?: string;
    next?: () => void;
    prev?: () => void;
    toggle?: () => void;
  } = {},
) {
  const initialPath = opts.initialPath ?? "/beamer/1";
  let currentPath = initialPath;

  function LocationProbe() {
    const loc = useLocation();
    currentPath = loc.pathname;
    return null;
  }

  function TestHarness() {
    useKeyboardControls(
      opts.next ?? (() => undefined),
      opts.prev ?? (() => undefined),
      opts.toggle ?? (() => undefined),
    );
    return null;
  }

  renderHook(TestHarness, {
    wrapper: ({ children }: { children: ReactNode }) => (
      <ConfigContext value={baseConfig}>
        <MemoryRouter initialEntries={[initialPath]}>
          <Routes>
            <Route path="/beamer/:id?" element={null} />
          </Routes>
          <LocationProbe />
          {children}
        </MemoryRouter>
      </ConfigContext>
    ),
  });

  return {
    getPathname: () => currentPath,
  };
}

const pressKey = (key: string, target: EventTarget = document.body) => {
  let event: KeyboardEvent | undefined;
  act(() => {
    event = new KeyboardEvent("keydown", {
      key,
      bubbles: true,
      cancelable: true,
    });
    Object.defineProperty(event, "target", { value: target });
    document.dispatchEvent(event);
  });
  return event!;
};

beforeEach(() => {
  document.body.focus();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("useKeyboardControls", () => {
  it("should attach a keydown listener to the document on mount and remove it on unmount", () => {
    const addSpy = vi.spyOn(globalThis, "addEventListener");
    const removeSpy = vi.spyOn(globalThis, "removeEventListener");

    const { unmount } = renderHook(
      () => useKeyboardControls(vi.fn(), vi.fn(), vi.fn()),
      {
        wrapper: ({ children }: { children: ReactNode }) => (
          <ConfigContext value={baseConfig}>
            <MemoryRouter initialEntries={["/beamer/1"]}>
              <Routes>
                <Route path="/beamer/:id?" element={null} />
              </Routes>
              {children}
            </MemoryRouter>
          </ConfigContext>
        ),
      },
    );

    expect(addSpy).toHaveBeenCalledWith("keydown", expect.any(Function));
    expect(removeSpy).not.toHaveBeenCalled();

    unmount();

    expect(removeSpy).toHaveBeenCalledWith("keydown", expect.any(Function));
  });

  it("should call next() when ArrowRight is pressed", () => {
    const next = vi.fn();
    const previous = vi.fn();
    const togglePause = vi.fn();

    setup({ next, prev: previous, toggle: togglePause });
    pressKey("ArrowRight");

    expect(next).toHaveBeenCalledTimes(1);
    expect(previous).not.toHaveBeenCalled();
    expect(togglePause).not.toHaveBeenCalled();
  });

  it("should call previous() when ArrowLeft is pressed", () => {
    const next = vi.fn();
    const previous = vi.fn();
    const togglePause = vi.fn();

    setup({ next, prev: previous, toggle: togglePause });
    pressKey("ArrowLeft");

    expect(previous).toHaveBeenCalledTimes(1);
    expect(next).not.toHaveBeenCalled();
    expect(togglePause).not.toHaveBeenCalled();
  });

  it("should call togglePause() and prevent default on Space", () => {
    const next = vi.fn();
    const previous = vi.fn();
    const togglePause = vi.fn();

    setup({ next, prev: previous, toggle: togglePause });
    const event = pressKey(" ");

    expect(togglePause).toHaveBeenCalledTimes(1);
    expect(event.defaultPrevented).toBe(true);
    expect(next).not.toHaveBeenCalled();
    expect(previous).not.toHaveBeenCalled();
  });

  it("should also accept the legacy 'Spacebar' key value", () => {
    const togglePause = vi.fn();
    setup({ toggle: togglePause });

    pressKey("Spacebar");

    expect(togglePause).toHaveBeenCalledTimes(1);
  });

  it("should ignore keypresses whose target is not document.body", () => {
    const next = vi.fn();
    const previous = vi.fn();
    const togglePause = vi.fn();

    setup({ next, prev: previous, toggle: togglePause });

    const otherElement = document.createElement("input");
    pressKey("ArrowRight", otherElement);
    pressKey("ArrowLeft", otherElement);
    pressKey(" ", otherElement);

    expect(next).not.toHaveBeenCalled();
    expect(previous).not.toHaveBeenCalled();
    expect(togglePause).not.toHaveBeenCalled();
  });

  it("should navigate to /beamer/:id when a matching playlist id is pressed", () => {
    const harness = setup();
    expect(harness.getPathname()).toBe("/beamer/1");

    pressKey("2");

    expect(harness.getPathname()).toBe("/beamer/2");
  });

  it("should not navigate when the pressed number has no matching playlist", () => {
    const harness = setup();
    expect(harness.getPathname()).toBe("/beamer/1");

    // No playlist with id 5 in the config — the listener should silently
    // no-op rather than crash or change route.
    pressKey("5");

    expect(harness.getPathname()).toBe("/beamer/1");
  });

  it("should ignore number 0 and non-numeric keys (no navigation)", () => {
    const harness = setup();

    pressKey("0");
    pressKey("a");

    expect(harness.getPathname()).toBe("/beamer/1");
  });

  it("should invoke the LATEST callbacks via stateRef without rebinding the listener", () => {
    // Regression test for the stateRef pattern introduced in this branch.
    // The keydown listener is attached exactly once (deps: []), but every
    // re-render updates stateRef.current with the fresh callbacks so the
    // listener always reads the latest values.
    const addSpy = vi.spyOn(globalThis, "addEventListener");
    const firstNext = vi.fn();
    const previous = vi.fn();
    const togglePause = vi.fn();

    const { rerender } = renderHook(
      ({ n }: { n: () => void }) =>
        useKeyboardControls(n, previous, togglePause),
      {
        wrapper: ({ children }: { children: ReactNode }) => (
          <ConfigContext value={baseConfig}>
            <MemoryRouter initialEntries={["/beamer/1"]}>
              <Routes>
                <Route path="/beamer/:id?" element={null} />
              </Routes>
              {children}
            </MemoryRouter>
          </ConfigContext>
        ),
        initialProps: { n: firstNext },
      },
    );

    // Listener attached exactly once
    const keydownCalls = addSpy.mock.calls.filter((c) => c[0] === "keydown");
    expect(keydownCalls).toHaveLength(1);

    // Re-render with a new callback
    const secondNext = vi.fn();
    rerender({ n: secondNext });

    // Still only ONE listener (the []-deps effect did not re-run)
    const keydownCallsAfter = addSpy.mock.calls.filter(
      (c) => c[0] === "keydown",
    );
    expect(keydownCallsAfter).toHaveLength(1);

    // The press invokes the SECOND callback (latest), not the first.
    pressKey("ArrowRight");
    expect(firstNext).not.toHaveBeenCalled();
    expect(secondNext).toHaveBeenCalledTimes(1);
  });
});
