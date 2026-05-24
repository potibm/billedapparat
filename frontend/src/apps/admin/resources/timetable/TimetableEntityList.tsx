import {
  List,
  Datagrid,
  TextField,
  ShowButton,
  NumberField,
  BooleanField,
  DateField,
} from "react-admin";
import { ExternalDataFilters } from "@admin/components/filters/ExternalDataFilters";

export const TimetableEntityList = () => (
  <List
    title="Timetable Entities"
    filters={ExternalDataFilters}
    sort={{ field: "id", order: "DESC" }}
  >
    <Datagrid rowClick="show" bulkActionButtons={false}>
      <NumberField source="id" label="ID" />

      <TextField source="source" label="Source" />
      <TextField source="external_id" label="External ID" />

      <TextField source="title" label="Title" />

      <DateField source="start_time" label="Start Time" showTime />
      <DateField source="end_time" label="End Time" showTime />

      <BooleanField source="hidden" />

      <ShowButton />
    </Datagrid>
  </List>
);
