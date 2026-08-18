import { useSSEStatus } from "../features/slides/hooks/useSSEConnection";

export const ConnectionStatusOverlay = () => {
  const connectionStatus = useSSEStatus();

  if (connectionStatus === "connected") return null;

  return (
    <div className="connectionstatus-overlay absolute bottom-4 left-4 z-50 pointer-events-none">
      {connectionStatus === "disconnected" && (
        <span
          className="connectionstatus-overlay__disconnected relative flex h-4 w-4"
          title="Disconnected"
        >
          <span className="relative inline-flex rounded-full h-4 w-4 bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.8)]"></span>
        </span>
      )}

      {connectionStatus === "connecting" && (
        <span
          className="connectionstatus-overlay__connecting relative flex h-4 w-4"
          title="Connecting..."
        >
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-orange-400 opacity-75"></span>
          <span className="relative inline-flex rounded-full h-4 w-4 bg-orange-500"></span>
        </span>
      )}
    </div>
  );
};
