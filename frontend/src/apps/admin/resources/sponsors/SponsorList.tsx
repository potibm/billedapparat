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

export const SponsorList = () => (
  <List title="Sponsor Screens">
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <ImageListPreviewField source="media_url_original" label="Logo" />

      <TextField source="content.text" label="Name" />

      <NumberField source="display_options.priority" label="Priority" />

      <StatusChipField source="status" />

      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export default SponsorList;
