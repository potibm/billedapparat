import { StatusChipField } from "@admin/components/fields/StatusChipField";
import { ImagePreviewField } from "@admin/components/fields/ImagePreviewField";
import {
  List,
  Datagrid,
  TextField,
  DeleteButton,
  NumberField,
} from "react-admin";
import { DefaultFilters } from "@admin/components/filters/DefaultFilters";

export const SocialMediaList = () => (
  <List
    title="Social with Media"
    filters={DefaultFilters}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="show" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <ImagePreviewField
        source="content.media.local_url"
        label="Logo"
        maxWidth={80}
        maxHeight={45}
        sortable={false}
      />

      <TextField source="content.title" label="Name" />

      <StatusChipField source="status" />

      <DeleteButton />
    </Datagrid>
  </List>
);

export default SocialMediaList;
