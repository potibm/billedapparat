import { Slide } from "../../types/slide.schema";
import { FormattedText } from "../FormattedText";
import { AuthorHeader } from "../ui/AuthorHeader";

export const SocialSlide = ({ slide }: { slide: Slide }) => {
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

      <AuthorHeader
        displayName={slide.author?.display_name}
        username={slide.author?.username}
        avatarUrl={slide.author?.avatar?.local_url}
        createdAt={slide.origin_created_at}
        className="social-slide__header"
      />

      {slide.content.body && (
        <div className="text-lg leading-relaxed text-gray-800 social-slide__body">
          <FormattedText text={slide.content.body} />
        </div>
      )}

      {!slide.content.body && slide.content.title && (
        <h2 className="text-2xl font-bold text-gray-400 uppercase">
          Social: {slide.content.title}
        </h2>
      )}
    </div>
  );
};
