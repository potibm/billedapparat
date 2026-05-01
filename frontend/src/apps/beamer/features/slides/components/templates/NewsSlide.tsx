import { Slide } from "../../types/slide.schema";
import Markdown from "react-markdown";

export const NewsSlide = ({ slide }: { slide: Slide }) => {
  return (
    <div className="slide-news w-full h-full">
      <div className="slide-news__container flex items-center justify-center w-full h-full">
        {slide.content.body ? (
          <div className="slide-news__content text-gray-200 p-8 max-w-4xl mx-auto">
            <h1>{slide.content.title}</h1>
            <Markdown>{slide.content.body}</Markdown>
          </div>
        ) : (
          <h2 className="slide-news__fallback text-5xl font-light text-gray-500 tracking-widest uppercase">
            News: {slide.content.title}
          </h2>
        )}
      </div>
    </div>
  );
};
