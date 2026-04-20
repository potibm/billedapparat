import { useRecordContext } from "react-admin";

export const ImageEditPreviewField = () => {
  const record = useRecordContext();
  if (!record || !record.media_url_original) return null;
  return (
    <div className="mb-4 ml-1">
      <p className="text-gray-300 text-xs mb-1">Current Slide</p>
      <img
        src={record.media_url_original}
        alt="Current Slide"
        className="max-w-[200px] max-h-[100px] object-contain bg-gray-100 p-2 rounded"
      />
    </div>
  );
};
