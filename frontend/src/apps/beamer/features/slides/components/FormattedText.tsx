import React, { useMemo } from "react";

interface FormattedTextProps {
  text?: string | null;
  className?: string | null;
}

const MAX_URL_LENGTH = 35; // Define a maximum length for displayed URLs

export function FormattedText({
  text,
  className,
}: Readonly<FormattedTextProps>) {
  const parts = useMemo(() => {
    if (!text) return [];

    const regex = /(https?:\/\/[^\s]+|@[a-zA-Z0-9_.-]+|#[a-zA-Z0-9äöüÄÖÜß_]+)/g;
    const rawParts = text.split(regex);

    return rawParts.map((part) => ({
      id: crypto.randomUUID(),
      content: part,
    }));
  }, [text]);

  if (parts.length === 0) return null;

  return (
    <div
      className={`formatted-text text-gray-300 whitespace-pre-wrap ${className}`}
    >
      {parts.map((partObj) => {
        const { id, content } = partObj;

        const isUrl = new RegExp(/^https?:\/\//).exec(content);
        if (isUrl) {
          const displayContent =
            content.length > MAX_URL_LENGTH
              ? `${content.substring(0, MAX_URL_LENGTH)}...`
              : content;
          return (
            <span
              key={id}
              className="formatted-text__link text-blue-500 hover:text-blue-700 hover:underline"
              title={content}
            >
              {displayContent}
            </span>
          );
        }

        if (content.startsWith("@")) {
          return (
            <span
              key={id}
              className="formatted-text__mention text-emerald-600 font-medium"
            >
              {content}
            </span>
          );
        }

        if (content.startsWith("#")) {
          return (
            <span
              key={id}
              className="formatted-text__hashtag text-purple-600 font-medium"
            >
              {content}
            </span>
          );
        }

        return <React.Fragment key={id}>{content}</React.Fragment>;
      })}
    </div>
  );
}
