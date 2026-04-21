import { StatusChipField } from "@admin/components/fields/StatusChipField";
import {
  List,
  Datagrid,
  TextField,
  EditButton,
  DeleteButton,
  NumberField,
} from "react-admin";

export const NewsList = () => (
  <List title="News Screens">
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <TextField source="content.title" label="Title" />

      <NumberField source="display_options.priority" label="Priority" />

      <StatusChipField source="status" />

      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export default NewsList;
