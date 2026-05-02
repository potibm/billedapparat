import { createContext } from "react";
import { AppConfig } from "./config.schemas";

export const ConfigContext = createContext<AppConfig | undefined>(undefined);
