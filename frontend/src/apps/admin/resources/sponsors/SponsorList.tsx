import { StatusChipField } from "@admin/components/fields/StatusChipField";
import { ImageListPreviewField } from "@admin/components/fields/ImageListPreviewField";
import {
  List,
  Datagrid,
  TextField,
  EditButton,
  DeleteButton,
  NumberField,
} from "react-admin";
import { DefaultFilters } from "@admin/components/filters/DefaultFilters";

export const SponsorList = () => (
  <List
    title="Sponsor Screens"
    filters={DefaultFilters}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <ImageListPreviewField
        source="content.media.local_url"
        label="Logo"
        sortable={false}
      />

      <TextField source="content.title" label="Name" />

      <NumberField source="display_options.priority" label="Priority" />

      <StatusChipField source="status" />

      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export default SponsorList;
