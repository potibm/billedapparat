import { useInput, InputProps } from "react-admin";
import MDEditor, { MDEditorProps } from "@uiw/react-md-editor";
import {
  FormControl,
  FormHelperText,
  InputLabel,
  Box,
  useTheme,
} from "@mui/material";

export interface MarkdownInputProps
  extends
    InputProps,
    Omit<
      MDEditorProps,
      "onChange" | "value" | "id" | "defaultValue" | "onBlur"
    > {}

export const MarkdownInput = ({
  source,
  label,
  validate,
  helperText,
  ...props
}: MarkdownInputProps) => {
  const {
    id,
    field,
    fieldState: { isTouched, invalid, error },
    isRequired,
  } = useInput({ source, validate, ...props });

  const theme = useTheme();
  const colorMode = theme.palette.mode;

  return (
    <FormControl
      fullWidth
      error={isTouched && invalid}
      margin="normal"
      className="ra-input-markdown"
    >
      <Box mb={1}>
        <InputLabel
          shrink
          htmlFor={id}
          required={isRequired}
          error={isTouched && invalid}
        >
          {label || source}
        </InputLabel>
      </Box>

      <Box mt={2} data-color-mode={colorMode}>
        <MDEditor
          id={id}
          value={field.value || ""}
          onChange={field.onChange}
          onBlur={field.onBlur}
          height={400}
          {...props}
        />
      </Box>

      <FormHelperText error={isTouched && invalid}>
        {(isTouched && error?.message) || helperText}
      </FormHelperText>
    </FormControl>
  );
};

export default MarkdownInput;
