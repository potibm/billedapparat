import { Avatar } from "flowbite-react";
import { Slide } from "../../types/slide.schema";
import { FormattedText } from "../FormattedText";

export const SocialSlide = ({ slide }: { slide: Slide }) => {
  const authorAvatarImage = slide.author?.avatar?.local_url;
  const authorName = slide.author?.display_name || "Unknown";

  return (
    <div className="w-full h-full p-4 flex flex-col items-start gap-4 social-slide__wrapper">
      {slide.content.media?.local_url && (
        <div className="w-full social-slide__media">
          <img
            src={slide.content.media?.local_url}
            alt={slide.content.title || "Social"}
            className="max-h-[80dvh] w-auto rounded-lg object-contain"
          />
        </div>
      )}

      <div className="flex items-center gap-3 social-slide__header">
        <Avatar img={authorAvatarImage} rounded size="sm" />
        <div className="flex flex-col sm:flex-row sm:items-center sm:gap-2">
          <span className="font-bold text-gray-900 dark:text-white">
            {authorName}
          </span>
          {slide.origin_created_at && (
            <span className="text-sm text-gray-500">
              • {new Date(slide.origin_created_at).toLocaleDateString()}
            </span>
          )}
        </div>
      </div>

      {slide.content.body && (
        <div className="text-lg leading-relaxed text-gray-800 social-slide__body">
          <FormattedText text={slide.content.body} />
        </div>
      )}

      {!slide.content.body && slide.content.title && (
        <h2 className="text-2xl font-bold text-gray-400 uppercase">
          News: {slide.content.title}
        </h2>
      )}
    </div>
  );
};
