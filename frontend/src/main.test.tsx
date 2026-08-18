import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { StrictMode } from "react";
import { bootstrapApp } from "./main";
import { ConfigContext } from "@core/config/ConfigContext";
import * as Sentry from "@sentry/react";
import { configureOidc } from "@admin/providers/authProvider";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";

vi.mock("react-dom/client", () => ({
  createRoot: vi.fn(),
}));

vi.mock("@sentry/react", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@sentry/react")>();
  return {
    ...actual,
    isInitialized: vi.fn(() => false),
    init: vi.fn(),
    replayIntegration: vi.fn(() => "replayIntegration"),
    browserTracingIntegration: vi.fn(() => "browserTracingIntegration"),
  };
});

vi.mock("@admin/providers/authProvider", () => ({
  configureOidc: vi.fn(),
}));

vi.mock("./App.tsx", () => ({
  default: function App() {
    return null;
  },
}));

const validConfig = {
  version: "1.0.0",
  environment: "test",
  sentry: {
    dsn: "https://public@sentry.example.com/1",
    environment: "test",
    version: "1.0.0",
  },
  format: {
    date: {
      locale: "en-US",
      options: {},
    },
  },
  playlists: [
    {
      id: 1,
      name: "Test",
      steps: [{ type: "sponsor", order: "random" }],
    },
  ],
  admin_urls: {
    timetable: "",
    news: "",
  },
  beamer: {
    allowed_animations: ["fade", "slideRight", "zoomIn", "flip", "urgent"],
  },
};

describe("bootstrapApp", () => {
  let rootElement: HTMLDivElement;
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    rootElement = document.createElement("div");
    rootElement.id = "root";
    document.body.appendChild(rootElement);

    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.spyOn(console, "warn").mockImplementation(() => {});
    consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    rootElement.remove();
  });

  it("renders a fatal error when the root element is missing", async () => {
    rootElement.remove();

    await bootstrapApp();

    expect(document.body.innerHTML).toContain("System Configuration Error");
    expect(document.body.innerHTML).toContain(
      "Failed to find the root element",
    );
    expect(createRoot).not.toHaveBeenCalled();
    expect(consoleErrorSpy.mock.calls[0][0]).toContain(
      "Bootstrap failed: Root element missing",
    );
  });

  it("fetches config and renders the app on success", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        statusText: "OK",
        json: async () => validConfig,
      }),
    );

    const renderMock = vi.fn();
    vi.mocked(createRoot).mockReturnValue({
      render: renderMock,
    } as unknown as ReturnType<typeof createRoot>);

    await bootstrapApp();

    expect(fetch).toHaveBeenCalledWith(
      "/api/config",
      expect.objectContaining({ signal: expect.anything() }),
    );
    expect(createRoot).toHaveBeenCalledWith(rootElement);
    expect(renderMock).toHaveBeenCalledTimes(1);

    const tree = renderMock.mock.calls[0][0];
    expect(tree.type).toBe(StrictMode);
    expect(tree.props.children.type).toBe(ConfigContext);
    expect(tree.props.children.props.value).toMatchObject({
      version: validConfig.version,
      environment: validConfig.environment,
    });
    expect(tree.props.children.props.children.type).toBe(App);
  });

  it("initializes Sentry when a DSN is present", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        statusText: "OK",
        json: async () => validConfig,
      }),
    );

    const renderMock = vi.fn();
    vi.mocked(createRoot).mockReturnValue({
      render: renderMock,
    } as unknown as ReturnType<typeof createRoot>);

    await bootstrapApp();

    expect(Sentry.init).toHaveBeenCalledWith(
      expect.objectContaining({
        dsn: validConfig.sentry.dsn,
        environment: validConfig.sentry.environment,
        release: validConfig.sentry.version,
        replaysSessionSampleRate: 0,
        replaysOnErrorSampleRate: 1,
        integrations: expect.arrayContaining([
          "replayIntegration",
          "browserTracingIntegration",
        ]),
      }),
    );
  });

  it("does not re-initialize Sentry if already initialized", async () => {
    vi.mocked(Sentry.isInitialized).mockReturnValue(true);

    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        statusText: "OK",
        json: async () => validConfig,
      }),
    );

    const renderMock = vi.fn();
    vi.mocked(createRoot).mockReturnValue({
      render: renderMock,
    } as unknown as ReturnType<typeof createRoot>);

    await bootstrapApp();

    expect(Sentry.init).not.toHaveBeenCalled();
    expect(renderMock).toHaveBeenCalledTimes(1);
  });

  it("configures OIDC when auth type is oidc", async () => {
    const configWithOidc = {
      ...validConfig,
      auth: {
        type: "oidc",
        name: "Dex",
        authority: "https://dex.example.com",
        client_id: "billedapparat",
      },
    };

    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        statusText: "OK",
        json: async () => configWithOidc,
      }),
    );

    const renderMock = vi.fn();
    vi.mocked(createRoot).mockReturnValue({
      render: renderMock,
    } as unknown as ReturnType<typeof createRoot>);

    await bootstrapApp();

    expect(configureOidc).toHaveBeenCalledWith(
      "https://dex.example.com",
      "billedapparat",
    );
    expect(renderMock).toHaveBeenCalledTimes(1);
  });

  it("renders an error when the fetch throws", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("Network failure")),
    );

    const renderMock = vi.fn();
    vi.mocked(createRoot).mockReturnValue({
      render: renderMock,
    } as unknown as ReturnType<typeof createRoot>);

    await bootstrapApp();

    expect(renderMock).toHaveBeenCalledTimes(1);
    expect(consoleErrorSpy.mock.calls[0][0]).toContain("Bootstrap failed:");
    expect(consoleErrorSpy.mock.calls[0]).toEqual(
      expect.arrayContaining([expect.any(Error)]),
    );
  });

  it("renders an error when the config response is not ok", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        statusText: "Service Unavailable",
      }),
    );

    const renderMock = vi.fn();
    vi.mocked(createRoot).mockReturnValue({
      render: renderMock,
    } as unknown as ReturnType<typeof createRoot>);

    await bootstrapApp();

    expect(renderMock).toHaveBeenCalledTimes(1);
    expect(consoleErrorSpy.mock.calls[0][0]).toContain("Bootstrap failed:");
    expect(consoleErrorSpy.mock.calls[0]).toEqual(
      expect.arrayContaining([expect.any(Error)]),
    );
  });

  it("renders an error when the config fails schema validation", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        statusText: "OK",
        json: async () => ({ invalid: true }),
      }),
    );

    const renderMock = vi.fn();
    vi.mocked(createRoot).mockReturnValue({
      render: renderMock,
    } as unknown as ReturnType<typeof createRoot>);

    await bootstrapApp();

    expect(renderMock).toHaveBeenCalledTimes(1);
    expect(consoleErrorSpy.mock.calls[0][0]).toContain("Bootstrap failed:");
    expect(consoleErrorSpy.mock.calls[0]).toEqual(
      expect.arrayContaining([expect.any(Error)]),
    );
  });
});
