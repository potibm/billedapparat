import { NumberInput, SearchInput } from "react-admin";
import { ReactElement } from "react";
import { QuickFilter } from "./QuickFilter";

export const ExternalDataFilters: ReactElement[] = [
  <NumberInput key="external_id" label="External ID" source="external_id" />,
  <SearchInput key="q" source="q" alwaysOn />,
  <QuickFilter
    key="hidden"
    source="hidden_yes"
    label="Hidden"
    defaultValue={true}
  />,
  <QuickFilter
    key="not_hidden"
    source="hidden_no"
    label="Shown"
    defaultValue={true}
  />,
];

export const ExternalDataFiltersWithUrgent: ReactElement[] = [
  ...ExternalDataFilters,
  <QuickFilter
    key="urgent"
    source="urgent_yes"
    label="Urgent"
    defaultValue={true}
  />,
  <QuickFilter
    key="not_urgent"
    source="urgent_no"
    label="Not Urgent"
    defaultValue={true}
  />,
];
