import { Slide } from "../../types/slide.schema";

export const SponsorSlide = ({ slide }: { slide: Slide }) => {
  return (
    <div className="flex items-center justify-center w-full h-full bg-black">
      {slide.content.media?.local_url ? (
        <img
          src={slide.content.media?.local_url}
          alt={slide.content.title || "Sponsor"}
          title={slide.content.title || "Sponsor"}
          className="w-full h-full object-contain"
        />
      ) : (
        <h2 className="text-5xl font-light text-gray-500 tracking-widest uppercase">
          Sponsor: {slide.content.title}
        </h2>
      )}
    </div>
  );
};
