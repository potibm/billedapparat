import { Slide } from "../../types/slide.schema";
import Markdown from "react-markdown";

export const UrgentSlide = ({ slide }: { slide: Slide }) => {
  return (
    <div className="w-full h-full slide-news">
      <div className="flex items-center justify-center w-full h-full slide-urgent__container">
        {slide.content.body ? (
          <div className="text-gray-200 p-8 max-w-4xl mx-auto slide-urgent__content">
            <h1 className="slide-urgent__title">{slide.content.title}</h1>
            <div className="slide-urgent__body">
              <Markdown>{slide.content.body}</Markdown>
            </div>
          </div>
        ) : (
          <h2 className="text-5xl font-light text-gray-500 tracking-widest uppercase slide-urgent__fallback">
            Urgent: {slide.content.title}
          </h2>
        )}
      </div>
    </div>
  );
};
