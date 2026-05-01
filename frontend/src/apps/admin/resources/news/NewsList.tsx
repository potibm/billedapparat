import { StatusChipField } from "@admin/components/fields/StatusChipField";
import {
  List,
  Datagrid,
  TextField,
  EditButton,
  DeleteButton,
  NumberField,
} from "react-admin";
import { DefaultFilters } from "@admin/components/filters/DefaultFilters";

export const NewsList = () => (
  <List
    title="News Screens"
    filters={DefaultFilters}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <TextField source="content.title" label="Title" />

      <NumberField source="display_options.priority" label="Priority" />

      <StatusChipField source="status" />

      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export default NewsList;
