import { useAppConfig } from "@core/config/useConfig";

interface DebugOverlayProps {
  isUrgent: boolean;
  stepInfo: {
    playlistName: string;
    type: string;
    current: number;
    total: number;
  } | null;
  activeAnimation: string;
  duration: number;
  isPaused: boolean;
}

export const DebugOverlay = ({
  isUrgent,
  stepInfo,
  activeAnimation,
  duration,
  isPaused,
}: DebugOverlayProps) => {
  const { version, environment } = useAppConfig();

  if (environment === "production") return null;

  return (
    <div className="absolute bottom-2 right-2 text-[10px] text-white/20 pointer-events-none font-mono z-50">
      {environment} | v{version} |{" "}
      {isUrgent
        ? "URGENT"
        : `${stepInfo?.playlistName} | ${stepInfo?.type} (${stepInfo?.current}/${stepInfo?.total})`}{" "}
      | {activeAnimation} | {duration}s{isPaused ? " | PAUSED" : ""}
    </div>
  );
};
