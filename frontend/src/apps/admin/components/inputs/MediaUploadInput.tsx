import { FileInput, FileField } from "react-admin";

export const MediaUploadInput = ({
  source = "media_upload",
  label = "Upload media",
}) => (
  <FileInput
    source={source}
    label={label}
    accept={{
      "image/*": [".png", ".jpg", ".jpeg", ".webp"],
      "video/mp4": [".mp4"],
    }}
    placeholder={<p>Drag media here or click to upload</p>}
  >
    <FileField source="src" title="title" />
  </FileInput>
);
