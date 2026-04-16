import {
  FunctionField,
  ChipField,
  RaRecord,
  FunctionFieldProps,
} from "react-admin";

export const StatusChipField = (props: Omit<FunctionFieldProps, "render">) => (
  <FunctionField
    {...props}
    label="Status"
    render={(record: RaRecord) => (
      <ChipField
        source="status"
        record={record}
        sx={{
          backgroundColor: record?.status === "active" ? "#d1fae5" : "#f3f4f6",
          color: record?.status === "active" ? "#065f46" : "#374151",
          fontWeight: "bold",
        }}
      />
    )}
  />
);
