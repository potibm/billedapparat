import {
  List,
  Datagrid,
  TextField,
  ShowButton,
  NumberField,
} from "react-admin";
import { DefaultFilters } from "@admin/components/filters/DefaultFilters";

export const TimetableSlideList = () => (
  <List
    title="Timetable Slides"
    filters={DefaultFilters}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="show" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <TextField source="content.title" label="Title" />

      <ShowButton />
    </Datagrid>
  </List>
);
