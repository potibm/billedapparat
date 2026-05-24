import { required, SelectInput } from "react-admin";

export const StatusSelectInput = () => (
  <SelectInput
    source="status"
    label="Status"
    choices={[
      { id: "active", name: "Active" },
      { id: "hidden", name: "Hidden" },
    ]}
    validate={required()}
  />
);
