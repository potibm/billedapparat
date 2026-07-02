import { StatusChipField } from "@admin/components/fields/StatusChipField";
import {
  List,
  Datagrid,
  TextField,
  DeleteButton,
  NumberField,
} from "react-admin";
import { SocialFilters } from "@admin/components/filters/SocialFilters";
import { SourceField } from "@admin/components/fields/SourceField";

export const SocialTextList = () => (
  <List
    title="Social with Text"
    filters={SocialFilters}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="show" bulkActionButtons={true}>
      <NumberField source="id" label="ID" />

      <SourceField source="source" label="Source" width={24} height={24} />

      <TextField source="author.display_name" label="Author" />

      <TextField source="content.title" label="Name" />

      <StatusChipField source="status" />

      <DeleteButton />
    </Datagrid>
  </List>
);

export default SocialTextList;
