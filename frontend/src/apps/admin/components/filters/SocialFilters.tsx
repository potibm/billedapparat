import {
  FilterList,
  FilterListItem,
  NumberInput,
  SearchInput,
} from "react-admin";
import { ReactElement } from "react";
import { QuickFilter } from "./QuickFilter";

export const SocialFilters: ReactElement[] = [
  <NumberInput key="id" label="ID" source="id" />,
  <SearchInput key="q" source="q" alwaysOn />,
  <FilterList key="source" label="Source" icon={<></>}>
    <FilterListItem label="Bluesky" value={{ source: "bluesky" }} />
    <FilterListItem label="Mastodon" value={{ source: "mastodon" }} />
  </FilterList>,
  <QuickFilter
    key="active"
    source="status_active"
    label="Active"
    defaultValue={true}
  />,
  <QuickFilter
    key="pending"
    source="status_pending"
    label="Pending"
    defaultValue={true}
  />,
  <QuickFilter
    key="deleted"
    source="status_deleted"
    label="Deleted"
    defaultValue={true}
  />,
];
