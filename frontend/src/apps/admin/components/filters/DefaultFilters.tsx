import { NumberInput, SearchInput } from "react-admin";
import { ReactElement } from "react";
import { QuickFilter } from "./QuickFilter";

export const DefaultFilters: ReactElement[] = [
  <NumberInput key="id" label="ID" source="id" />,
  <SearchInput key="q" source="q" alwaysOn />,
  <NumberInput
    key="priority"
    label="Priority"
    source="display_options.priority"
    min={1}
    max={9}
    step={1}
  />,
  <QuickFilter
    key="active"
    source="status_active"
    label="Active"
    defaultValue={true}
  />,
  <QuickFilter
    key="inactive"
    source="status_inactive"
    label="Inactive"
    defaultValue={true}
  />,
];
