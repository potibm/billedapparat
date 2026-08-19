import { FileInput, FileField } from "react-admin";

export const MediaUploadInput = ({
  source = "media_upload",
  label = "Upload media",
}) => (
  <FileInput
    source={source}
    label={label}
    accept={{ "image/*,video/mp4": [".png", ".jpg", ".jpeg", ".webp", ".mp4"] }}
    placeholder={<p>Drag the slide here or click to upload</p>}
  >
    <FileField source="src" title="title" />
  </FileInput>
);
