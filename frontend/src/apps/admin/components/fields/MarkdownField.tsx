import { useRecordContext, FieldProps } from "react-admin";
import MDEditor from "@uiw/react-md-editor";
import { Box, Typography, useTheme } from "@mui/material";

export interface MarkdownFieldProps extends FieldProps {
  showRaw?: boolean;
}

export const MarkdownField = ({
  source,
  label,
  showRaw = false,
  className,
  ...props
}: MarkdownFieldProps) => {
  const record = useRecordContext();
  const theme = useTheme();
  const colorMode = theme.palette.mode;

  if (record == null || source == null) return null;

  const value = source
    .split(".")
    .reduce(
      (acc: unknown, part: string) =>
        acc && typeof acc === "object"
          ? (acc as Record<string, unknown>)[part]
          : undefined,
      record,
    ) as string | undefined;

  if (!value) return null;

  return (
    <Box
      sx={{ my: 2 }}
      className={`ra-field-markdown ${className || ""}`}
      {...props}
    >
      <Box sx={{ mb: 1 }}>
        <Typography variant="caption" color="text.secondary">
          {label ?? source}
        </Typography>
      </Box>

      <Box data-color-mode={colorMode}>
        <Box
          sx={{
            p: 2,
            border: 1,
            borderColor: "divider",
            borderRadius: 1,
            bgcolor: "background.paper",
          }}
        >
          <MDEditor.Markdown source={value} />
        </Box>

        {showRaw && (
          <Box sx={{ mt: 2 }}>
            <Typography
              variant="caption"
              color="text.secondary"
              sx={{ display: "block", mb: 1 }}
            >
              Markdown
            </Typography>
            <Box
              component="pre"
              sx={{
                p: 2,
                m: 0,
                border: 1,
                borderColor: "divider",
                borderRadius: 1,
                bgcolor: "action.hover",
                overflowX: "auto",
                fontFamily: "monospace",
                fontSize: "0.875rem",
                whiteSpace: "pre-wrap",
                wordBreak: "break-word",
              }}
            >
              {value}
            </Box>
          </Box>
        )}
      </Box>
    </Box>
  );
};

export default MarkdownField;
