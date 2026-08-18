import { useEffect, useSyncExternalStore } from "react";
import { slideStore } from "../store/slideStore";

export const useSSEStatus = () => {
  return useSyncExternalStore(slideStore.subscribe, slideStore.getStatus);
};

export const useSSELifecycle = () => {
  useEffect(() => {
    slideStore.connect();
    return () => {
      slideStore.disconnect();
    };
  }, []);
};
