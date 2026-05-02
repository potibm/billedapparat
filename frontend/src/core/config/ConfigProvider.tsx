import { useEffect, useState, ReactNode } from "react";
import { AppConfigSchema, AppConfig } from "./config.schemas";
import { createLogger } from "@core/logger/logger";
import { ConfigContext } from "./ConfigContext";
import * as Sentry from "@sentry/react";

const log = createLogger("Config");
const API_HOST = import.meta.env.VITE_API_HOST ?? "http://localhost:3101";

export const ConfigProvider = ({ children }: { children: ReactNode }) => {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const controller = new AbortController();
    const timeoutId = globalThis.setTimeout(() => controller.abort(), 10000);

    fetch(API_HOST + "/api/config", { signal: controller.signal })
      .then((res) => {
        if (!res.ok) throw new Error(`Config error: ${res.statusText}`);
        return res.json();
      })
      .then((data) => {
        const validated = AppConfigSchema.parse(data);
        setConfig(validated);
        log.info("System config loaded successfully", validated);
      })
      .catch((err) => {
        if (err.name === "AbortError") {
          log.debug("Fetch aborted (cleanly)");
          return;
        }
        log.error("Failed to load system config:", err);
        setError(err instanceof Error ? err.message : "Unknown config error");
      })
      .finally(() => {
        globalThis.clearTimeout(timeoutId);
        setLoading(false);
      });

    return () => {
      controller.abort();
      globalThis.clearTimeout(timeoutId);
    };
  }, []);

  useEffect(() => {
    if (config?.sentry.dsn) {
      if (Sentry.isInitialized()) {
        log.warn(
          "Sentry is already initialized, skipping re-initialization with new config",
        );
        return;
      }

      Sentry.init({
        dsn: config.sentry.dsn,
        environment: config.sentry.environment,
        release: config.sentry.version,
        replaysSessionSampleRate: config.sentry.replay_session_sample_rate,
        replaysOnErrorSampleRate: config.sentry.replay_error_sample_rate,
        integrations: [
          Sentry.replayIntegration(),
          Sentry.browserTracingIntegration(),
        ],
      });

      log.info(
        "Sentry initialized",
        "version",
        config.version,
        "environment",
        config.environment,
      );
    }
  }, [config]);

  if (error) {
    return (
      <div className="text-red-50">
        <h2>System Configuration Error</h2>
        <pre>{error}</pre>
      </div>
    );
  }

  if (loading) {
    return null;
  }

  if (!config) {
    return null;
  }

  return <ConfigContext value={config}>{children}</ConfigContext>;
};
