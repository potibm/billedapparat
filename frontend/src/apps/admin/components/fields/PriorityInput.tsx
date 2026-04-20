import { NumberInput } from "react-admin";

export const PriorityInput = () => (
  <NumberInput
    source="display_options.priority"
    label="Priority"
    defaultValue={1}
    min={1}
    max={10}
    step={1}
    type="slider"
    helperText="Higher number = higher visibility (e.g., 10 for main sponsors, 1 for standard)"
  />
);
