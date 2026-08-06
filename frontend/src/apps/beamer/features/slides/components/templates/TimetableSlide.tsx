import { Slide } from "../../types/slide.schema";
import Markdown, { Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  TimetableTable,
  TimetableRow,
  TimetableTh,
  TimetableTd,
} from "./TimetableMarkdown";

const markdownComponents: Components = {
  table: TimetableTable,

  tr: TimetableRow,

  td: TimetableTd,
  th: TimetableTh,
};

export const TimetableSlide = ({ slide }: { slide: Slide }) => {
  return (
    <div className="w-full h-full slide-timetable">
      <div className="flex items-center justify-center w-full h-full slide-timetable__container">
        {slide.content.body ? (
          <div className="text-gray-200 p-8 max-w-4xl mx-auto slide-timetable__content">
            <h1 className="slide-timetable__title">{slide.content.title}</h1>
            <div className="slide-timetable__body">
              <Markdown
                remarkPlugins={[remarkGfm]}
                components={markdownComponents}
              >
                {slide.content.body}
              </Markdown>
            </div>
          </div>
        ) : (
          <h2 className="text-5xl font-light text-gray-500 tracking-widest uppercase slide-timetable__fallback">
            Timetable: {slide.content.title}
          </h2>
        )}
      </div>
    </div>
  );
};
