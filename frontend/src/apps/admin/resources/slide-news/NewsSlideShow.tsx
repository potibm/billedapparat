import MarkdownField from "@admin/components/fields/MarkdownField";
import { StatusChipField } from "@admin/components/fields/StatusChipField";
import { Show, TextField, SimpleShowLayout, NumberField } from "react-admin";

export const NewsSlideShow = () => (
  <Show title="Show News Slides">
    <SimpleShowLayout>
      <NumberField source="id" label="ID" />

      <TextField source="content.title" label="Title" />

      <MarkdownField source="content.body" label="Body" showRaw={true} />

      <NumberField source="display_options.priority" label="Priority" />

      <StatusChipField source="status" />
    </SimpleShowLayout>
  </Show>
);
