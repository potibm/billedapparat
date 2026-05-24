import { StatusChipField } from "@admin/components/fields/StatusChipField";
import { ImagePreviewField } from "@admin/components/fields/ImagePreviewField";
import {
  List,
  Datagrid,
  TextField,
  DeleteButton,
  NumberField,
} from "react-admin";
import { SocialFilters } from "@admin/components/filters/SocialFilters";
import { SourceField } from "@admin/components/fields/SourceField";

export const SocialMediaList = () => (
  <List
    title="Social with Media"
    filters={SocialFilters}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="show" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <SourceField source="source" label="Source" width={24} height={24} />

      <ImagePreviewField
        source="content.media.local_url"
        label="Logo"
        maxWidth={80}
        maxHeight={45}
        sortable={false}
      />

      <TextField source="author.display_name" label="Author" />

      <TextField source="content.title" label="Name" />

      <StatusChipField source="status" />

      <DeleteButton />
    </Datagrid>
  </List>
);

export default SocialMediaList;
