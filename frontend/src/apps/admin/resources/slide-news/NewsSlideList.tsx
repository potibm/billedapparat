import { StatusChipField } from "@admin/components/fields/StatusChipField";
import {
  List,
  Datagrid,
  TextField,
  ShowButton,
  NumberField,
} from "react-admin";
import { DefaultFilters } from "@admin/components/filters/DefaultFilters";

export const NewsSlideList = () => (
  <List
    title="News Slides"
    filters={DefaultFilters}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="show" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <TextField source="content.title" label="Title" />

      <NumberField source="display_options.priority" label="Priority" />

      <StatusChipField source="status" />

      <ShowButton />
    </Datagrid>
  </List>
);
