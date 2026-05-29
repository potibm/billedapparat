import {
  List,
  Datagrid,
  TextField,
  EditButton,
  DeleteButton,
  NumberField,
} from "react-admin";
import { DefaultFilters } from "@admin/components/filters/DefaultFilters";
import { ImagePreviewField } from "@admin/components/fields/ImagePreviewField";
import { StatusToggleField } from "@admin/components/fields/StatusToggleField";

export const SponsorList = () => (
  <List
    title="Sponsor Screens"
    filters={DefaultFilters}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <ImagePreviewField
        source="content.media.local_url"
        label="Logo"
        maxWidth={80}
        maxHeight={45}
        sortable={false}
      />

      <TextField source="content.title" label="Name" />

      <NumberField source="display_options.priority" label="Priority" />

      <StatusToggleField source="status" inactiveValue="hidden" />

      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export default SponsorList;
