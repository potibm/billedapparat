import { ImageField, ImageFieldProps } from "react-admin";

export const ImageListPreviewField = (props: ImageFieldProps) => (
  <ImageField
    {...props}
    sx={{
      "& img": {
        maxWidth: "80px",
        maxHeight: "45px",
        objectFit: "contain",
        backgroundColor: "#f3f4f6",
        padding: "4px",
        borderRadius: "4px",
      },
    }}
  />
);
