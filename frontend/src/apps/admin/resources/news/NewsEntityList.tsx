import {
  List,
  Datagrid,
  TextField,
  ShowButton,
  NumberField,
  BooleanField,
} from "react-admin";
import { ExternalDataFiltersWithUrgent } from "@admin/components/filters/ExternalDataFilters";

export const NewsEntityList = () => (
  <List
    title="News Entities"
    filters={ExternalDataFiltersWithUrgent}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="show" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <TextField source="source" label="Source" />
      <TextField source="external_id" label="External ID" />

      <TextField source="title" label="Title" />

      <BooleanField source="is_urgent" />

      <BooleanField source="is_hidden" />

      <ShowButton />
    </Datagrid>
  </List>
);
