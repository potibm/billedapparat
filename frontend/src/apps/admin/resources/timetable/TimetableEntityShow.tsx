import {
  Show,
  TextField,
  SimpleShowLayout,
  NumberField,
  BooleanField,
  DateField,
} from "react-admin";

export const TimetableEntityShow = () => (
  <Show title="Show Timetable Entities">
    <SimpleShowLayout>
      <NumberField source="id" label="ID" />

      <TextField source="source" label="Source" />
      <TextField source="external_id" label="External ID" />

      <TextField source="title" label="Title" />

      <TextField source="description" label="Description" />

      <DateField source="start_time" label="Start Time" showTime />
      <DateField source="end_time" label="End Time" showTime />

      <TextField source="location.name" label="Location Name" />
      <TextField source="location.address" label="Location Address" />

      <TextField source="category.name" label="Category Name" />
      <TextField source="category.color" label="Category Color" />

      <BooleanField source="hidden" />
    </SimpleShowLayout>
  </Show>
);
