import { Slide } from "../../types/slide.schema";
import { TextSlide } from "./TextSlide";

export const NewsSlide = ({ slide }: { slide: Slide }) => (
  <TextSlide slide={slide} variant="news" fallbackPrefix="News:" />
);
