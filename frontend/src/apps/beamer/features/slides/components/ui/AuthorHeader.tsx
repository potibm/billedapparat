import { Avatar } from "flowbite-react";
import { SourceIcon } from "@core/sources/SourceIcon";

interface AuthorHeaderProps {
  displayName?: string | null;
  username?: string | null;
  avatarUrl?: string | null;
  createdAt?: string | null;
  source?: string | null;
  className?: string; // Allows passing BEM classes or layout spacing from the outside
}

export const AuthorHeader = ({
  displayName,
  username,
  avatarUrl,
  createdAt,
  source,
  className = "",
}: AuthorHeaderProps) => {
  const authorName = displayName || "Unknown";

  return (
    <div className={`author-header flex items-center gap-3 ${className}`}>
      <div className="author-header__avatar">
        <Avatar img={avatarUrl || undefined} rounded size="sm" />
      </div>

      <div className="author-header__info flex flex-col sm:flex-row sm:items-center sm:gap-2">
        <span className="author-header__name font-bold text-gray-300">
          {authorName}
        </span>

        {username && (
          <span className="author-header__username text-sm text-gray-500 hidden sm:inline">
            @{username}
          </span>
        )}

        {createdAt && (
          <span className="author-header__date text-sm text-gray-500">
            <span className="hidden sm:inline">• </span>
            {new Date(createdAt).toLocaleDateString()}
          </span>
        )}

        {source && (
          <span className="author-header__source inline-flex items-center gap-1 text-sm text-gray-500">
            <span className="hidden sm:inline">• </span>
            <SourceIcon source={source} width={16} height={16} />
          </span>
        )}
      </div>
    </div>
  );
};
