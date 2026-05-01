import {
  List,
  Datagrid,
  TextField,
  EditButton,
  DeleteButton,
  NumberField,
} from "react-admin";

export const FilterRulesList = () => (
  <List title="Filter Rules" sort={{ field: "id", order: "DESC" }}>
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <TextField source="source" label="Source" />

      <TextField source="type" label="Type" />

      <TextField source="value" label="Value" />

      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export default FilterRulesList;
