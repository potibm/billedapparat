import { Avatar } from "flowbite-react";
import { Slide } from "../../types/slide.schema";
import { createLogger } from "@core/logger/logger";

const logger = createLogger("Slideshow");

export const SocialSlide = ({ slide }: { slide: Slide }) => {
  //  logger.info("Rendering SocialSlide with content:", slide);

  return (
    <div className="w-full h-full ba-slide-social-background">
      <div className="flex items-center justify-center w-full h-full ba-slide-social">
        {slide.content.body ? (
          <div className="ba-slide-content">
            <img
              src={slide.content.media?.local_url}
              alt={slide.content.title || "Social"}
              title={slide.content.title || "Social"}
              className="w-[70%] h-[70%] object-contain"
            />

            <div>
              <Avatar
                img={slide.author?.media?.local_url}
                alt={slide.author?.displayname || slide.author?.username}
                rounded
              />
            </div>

            <div>{slide.author?.displayname || slide.author?.username}</div>

            <div>{slide.content.body}</div>
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
