import { use } from "react";
import { ConfigContext } from "./ConfigContext";

export const useAppConfig = () => {
  const context = use(ConfigContext);
  if (context === undefined) {
    throw new Error(
      "useAppConfig must be used within a ConfigContext.Provider",
    );
  }
  return context;
};
