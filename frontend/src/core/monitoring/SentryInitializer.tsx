import React, { useEffect } from "react";
import * as Sentry from "@sentry/react";
import { useAppConfig } from "../config/useConfig";
import { createLogger } from "@core/logger/logger";

const log = createLogger("Core");

const SentryInitializer: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const { sentry } = useAppConfig();

  useEffect(() => {
    if (sentry.dsn) {
      try {
        Sentry.init({
          dsn: sentry.dsn,
          environment: import.meta.env.MODE,
          release: `billedapparat@${sentry.version ?? "unknown"}`,
          integrations: [
            Sentry.browserTracingIntegration(),
            Sentry.replayIntegration(),
          ],
          tracesSampleRate: sentry.traces_sample_rate ?? 1,
          replaysSessionSampleRate: sentry.replay_session_sample_rate ?? 0.1,
          replaysOnErrorSampleRate: sentry.replay_error_sample_rate ?? 1,
        });
      } catch (error: unknown) {
        log.error("Error initializing Sentry", error);
      }
    }
  }, [sentry]);

  return (
    <Sentry.ErrorBoundary fallback={<p>Critical Error in Billedapparat.</p>}>
      {children}
    </Sentry.ErrorBoundary>
  );
};

export default SentryInitializer;
