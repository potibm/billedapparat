import { useRecordContext } from "react-admin";
import { ImagePreviewField } from "./ImagePreviewField";

export const ImageEditPreviewField = () => {
  const record = useRecordContext();
  if (!record || !record.content?.media?.local_url) return null;
  return (
    <div className="mb-4 ml-1">
      <p className="text-gray-300 text-xs mb-1">Current Slide</p>
      <ImagePreviewField
        source="content.media.local_url"
        label="Preview"
        maxWidth={200}
        maxHeight={100}
      />
    </div>
  );
};
