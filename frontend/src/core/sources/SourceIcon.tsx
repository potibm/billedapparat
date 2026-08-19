import type { IconType } from "react-icons";
import { SiMastodon, SiBluesky, SiTwitch, SiDiscord } from "react-icons/si";
import { PouetIcon } from "./icons/PouetIcon";

interface SourceIconProps {
  source: string | null | undefined;
  width?: number | string;
  height?: number | string;
  className?: string;
  title?: string;
  customIcon?: React.ReactNode;
}

const iconMap: Record<string, IconType> = {
  mastodon: SiMastodon,
  bluesky: SiBluesky,
  twitch: SiTwitch,
  discord: SiDiscord,
  pouet: PouetIcon,
};

export const SourceIcon = ({
  source,
  width,
  height,
  className = "",
  title,
  customIcon,
}: SourceIconProps) => {
  if (!source) return null;

  const IconComponent = iconMap[source];

  if (IconComponent) {
    const sizeValue =
      width !== undefined ? width : height !== undefined ? height : 16;
    const size =
      typeof sizeValue === "number" ? sizeValue : parseInt(String(sizeValue));
    return (
      <IconComponent
        size={size}
        className={className}
        title={title || source}
      />
    );
  }

  if (customIcon) {
    return <>{customIcon}</>;
  }

  return <span className={`text-sm text-gray-500 ${className}`}>{source}</span>;
};
