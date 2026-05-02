import { useEffect, useState, ReactNode } from "react";
import { AppConfigSchema, AppConfig } from "./config.schemas";
import { createLogger } from "@core/logger/logger";
import { ConfigContext } from "./ConfigContext";

const log = createLogger("Config");
const API_HOST = import.meta.env.VITE_API_HOST ?? "http://localhost:3101";

export const ConfigProvider = ({ children }: { children: ReactNode }) => {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch(API_HOST + "/api/config")
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
        log.error("Failed to load system config:", err);
        setError(err.message);
      });
  }, []);

  if (error) {
    return (
      <div className="text-red-50">
        <h2>System Configuration Error</h2>
        <pre>{error}</pre>
      </div>
    );
  }

  if (!config) {
    return null;
  }

  return <ConfigContext value={config}>{children}</ConfigContext>;
};
