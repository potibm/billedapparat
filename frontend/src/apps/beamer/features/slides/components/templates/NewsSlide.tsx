import { Slide } from "../../types/slide.schema";
import Markdown from 'react-markdown'

export const NewsSlide = ({ slide }: { slide: Slide }) => {
  return (
    <div className="w-full h-full ba-slide-news-background">
        <div className="flex items-center justify-center w-full h-full ba-slide-news">
        {slide.content.body ? (
        
            <div className="ba-slide-content">
                <h1>{slide.content.title}</h1>
                <Markdown>{slide.content.body}</Markdown>
            </div>

        ) : (
            <h2 className="text-5xl font-light text-gray-500 tracking-widest uppercase">
            News: {slide.content.title}
            </h2>
        )}
        </div>
    </div>
  );
};
