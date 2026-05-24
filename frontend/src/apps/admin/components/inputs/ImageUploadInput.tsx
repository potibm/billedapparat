import { ImageInput, ImageField } from "react-admin";

export const ImageUploadInput = ({
  source = "image_upload",
  label = "Upload screen",
}) => (
  <ImageInput
    source={source}
    label={label}
    placeholder={<p>Drag the slide here or click to upload</p>}
  >
    <ImageField source="src" title="title" />
  </ImageInput>
);
