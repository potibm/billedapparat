import { Slide } from "../../types/slide.schema";
import Markdown from "react-markdown";

interface TextSlideProps {
  slide: Slide;
  variant: "news" | "urgent";
  fallbackPrefix: string;
}

export const TextSlide = ({
  slide,
  variant,
  fallbackPrefix,
}: TextSlideProps) => {
  return (
    <div className={`w-full h-full slide-${variant}`}>
      <div
        className={`flex items-center justify-center w-full h-full slide-${variant}__container`}
      >
        {slide.content.body ? (
          <div
            className={`text-gray-200 p-8 max-w-4xl mx-auto slide-${variant}__content`}
          >
            <h1 className={`slide-${variant}__title`}>{slide.content.title}</h1>
            <div className={`slide-${variant}__body`}>
              <Markdown>{slide.content.body}</Markdown>
            </div>
          </div>
        ) : (
          <h2
            className={`text-5xl font-light text-gray-500 tracking-widest uppercase slide-${variant}__fallback`}
          >
            {fallbackPrefix} {slide.content.title}
          </h2>
        )}
      </div>
    </div>
  );
};
