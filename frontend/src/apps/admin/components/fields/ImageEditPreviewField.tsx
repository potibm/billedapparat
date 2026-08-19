import { useRecordContext } from "react-admin";

export const ImageEditPreviewField = () => {
  const record = useRecordContext();
  if (!record || !record.content?.media?.local_url) return null;

  const isVideo = record.content.media.mime_type?.startsWith("video/");

  return (
    <div className="mb-4 ml-1">
      <p className="text-gray-300 text-xs mb-1">Current Slide</p>
      {isVideo ? (
        <video
          src={record.content.media.local_url}
          controls
          muted
          style={{ maxWidth: 200, maxHeight: 100 }}
        />
      ) : (
        <img
          src={record.content.media.local_url}
          alt="Preview"
          style={{ maxWidth: 200, maxHeight: 100 }}
        />
      )}
    </div>
  );
};
