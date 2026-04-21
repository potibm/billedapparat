import { Slide } from "../types/slide.schema";
import { NewsSlide } from "./templates/NewsSlide";
import { SponsorSlide } from "./templates/SponsorSlide";
// import { UrgentSlide } from "./templates/UrgentSlide";
// import { NewsSlide } from "./templates/NewsSlide";
// import { TimetableSlide } from "./templates/TimetableSlide";

interface SlideRendererProps {
  slide: Slide;
}

export const SlideRenderer = ({ slide }: SlideRendererProps) => {
  switch (slide.content.type) {
    case "sponsor":
      return <SponsorSlide slide={slide} />;
    case "news":
      return <NewsSlide slide={slide} />;

    /*case "urgent":
      return <UrgentSlide slide={slide} />;
   
    case "timetable":
      return <TimetableSlide slide={slide} />;
    */

    default:
      return (
        <div className="flex items-center justify-center w-full h-full bg-gray-900 text-white">
          <h2 className="text-4xl text-gray-400">
            {slide.content.title || "Loading content..."}
          </h2>
        </div>
      );
  }
};
