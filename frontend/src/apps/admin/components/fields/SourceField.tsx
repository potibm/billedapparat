import { FunctionField, RaRecord, FunctionFieldProps } from "react-admin";
import { SourceIcon } from "@core/sources/SourceIcon";

interface SourceFieldProps extends Omit<FunctionFieldProps, "render"> {
  width?: number | string;
  height?: number | string;
  title?: string;
}

export const SourceField = ({
  width = 16,
  height = 16,
  title,
  ...props
}: SourceFieldProps) => (
  <FunctionField
    {...props}
    label="Source"
    render={(record?: RaRecord) => {
      if (!record || !record.source) return null;

      return (
        <SourceIcon
          source={record.source}
          width={width}
          height={height}
          title={title}
        />
      );
    }}
  />
);
