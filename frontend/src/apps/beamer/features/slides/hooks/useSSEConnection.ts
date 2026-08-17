import { useEffect } from "react";
import { slideStore } from "../store/slideStore";

export const useSSEConnection = () => {
  useEffect(() => {
    slideStore.connect();
    return () => {
      slideStore.disconnect();
    };
  }, []);
};
