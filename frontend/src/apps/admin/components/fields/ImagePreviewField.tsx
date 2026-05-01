import { useState, ReactNode } from "react";
import { ImageField, ImageFieldProps, useRecordContext } from "react-admin";
import { Dialog, DialogContent, IconButton, Box } from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import ImageIcon from "@mui/icons-material/Image";

export interface ImagePreviewFieldProps extends ImageFieldProps {
  maxWidth?: string | number;
  maxHeight?: string | number;
  placeholder?: ReactNode;
}

export const ImagePreviewField = (props: ImagePreviewFieldProps) => {
  const {
    maxWidth = "160px",
    maxHeight = "90px",
    source,
    placeholder,
    ...rest
  } = props;
  const [open, setOpen] = useState(false);
  const record = useRecordContext(props);

  const getSourceValue = (obj: unknown, path: string): string | undefined => {
    const result = path.split(".").reduce((acc: unknown, part: string) => {
      if (acc && typeof acc === "object") {
        return (acc as Record<string, unknown>)[part];
      }
      return undefined;
    }, obj);

    return typeof result === "string" ? result : undefined;
  };

  const imageUrl = record && source ? getSourceValue(record, source) : null;

  if (!imageUrl) {
    return (
      <Box
        sx={{
          width: typeof maxWidth === "number" ? `${maxWidth}px` : maxWidth,
          height: typeof maxHeight === "number" ? `${maxHeight}px` : maxHeight,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: "action.hover",
          borderRadius: "4px",
          borderColor: "divider",
          border: "0px solid",
          color: "text.disabled",
          margin: "4px",
        }}
      >
        {placeholder || <ImageIcon fontSize="small" />}
      </Box>
    );
  }

  return (
    <>
      <Box
        onClick={(e) => {
          e.stopPropagation();
          setOpen(true);
        }}
        sx={{ cursor: "pointer", display: "inline-block" }}
      >
        <ImageField
          {...rest}
          source={source}
          sx={{
            "& img": {
              maxWidth: maxWidth,
              maxHeight: maxHeight,
              objectFit: "contain",
              backgroundColor: "#f3f4f6",
              padding: "4px",
              borderRadius: "4px",
            },
          }}
        />
      </Box>

      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        maxWidth="lg"
        fullWidth={false}
        onClick={(e) => e.stopPropagation()}
      >
        <IconButton
          onClick={(e) => {
            e.stopPropagation();
            setOpen(false);
          }}
          sx={{
            position: "absolute",
            right: 8,
            top: 8,
            color: (theme) => theme.palette.grey[500],
            zIndex: 1,
          }}
        >
          <CloseIcon />
        </IconButton>
        <DialogContent
          sx={{
            p: 1,
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            minWidth: 300,
          }}
        >
          <Box
            component="button"
            onClick={(e) => {
              e.stopPropagation();
              setOpen(false);
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.stopPropagation();
                setOpen(false);
              }
            }}
            sx={{
              background: "none",
              border: "none",
              padding: 0,
              cursor: "pointer",
              "&:focus-visible": {
                outline: "2px solid",
                outlineColor: "primary.main",
                borderRadius: "4px",
              },
            }}
          >
            <img
              src={imageUrl}
              alt="Preview"
              style={{
                maxWidth: "100%",
                maxHeight: "85vh",
                objectFit: "contain",
                display: "block",
              }}
            />
          </Box>
        </DialogContent>
      </Dialog>
    </>
  );
};
