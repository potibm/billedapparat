import { Slide } from "../../types/slide.schema";
import { TextSlide } from "./TextSlide";

export const UrgentSlide = ({ slide }: { slide: Slide }) => (
  <TextSlide slide={slide} variant="urgent" fallbackPrefix="Urgent:" />
);
