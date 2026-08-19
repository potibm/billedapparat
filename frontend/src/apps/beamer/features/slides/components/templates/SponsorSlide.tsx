import { Slide } from "../../types/slide.schema";

export const SponsorSlide = ({ slide }: { slide: Slide }) => {
  const isVideo = slide.content.media?.mime_type?.startsWith("video/");

  return (
    <div className="flex items-center justify-center w-full h-full bg-black slide-sponsor">
      {slide.content.media?.local_url ? (
        isVideo ? (
          <video
            src={slide.content.media.local_url}
            autoPlay
            loop
            muted
            playsInline
            className="w-full h-full object-contain slide-sponsor__video"
          />
        ) : (
          <img
            src={slide.content.media.local_url}
            alt={slide.content.title || "Sponsor"}
            title={slide.content.title || "Sponsor"}
            className="w-full h-full object-contain slide-sponsor__image"
          />
        )
      ) : (
        <h2 className="text-5xl font-light text-gray-500 tracking-widest uppercase slide-sponsor__fallback">
          Sponsor: {slide.content.title}
        </h2>
      )}
    </div>
  );
};
