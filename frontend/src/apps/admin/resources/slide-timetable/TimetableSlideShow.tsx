import MarkdownField from "@admin/components/fields/MarkdownField";
import { Show, TextField, SimpleShowLayout, NumberField } from "react-admin";

export const TimetableSlideShow = () => (
  <Show title="Show Timetable Slide">
    <SimpleShowLayout>
      <NumberField source="id" label="ID" />

      <TextField source="content.title" label="Title" />

      <MarkdownField source="content.body" label="Body" showRaw={true} />
    </SimpleShowLayout>
  </Show>
);
