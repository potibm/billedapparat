import MarkdownField from "@admin/components/fields/MarkdownField";
import {
  Show,
  TextField,
  SimpleShowLayout,
  NumberField,
  BooleanField,
} from "react-admin";

export const NewsEntityShow = () => (
  <Show title="Show News Entities">
    <SimpleShowLayout>
      <NumberField source="id" label="ID" />

      <TextField source="source" label="Source" />
      <TextField source="external_id" label="External ID" />

      <TextField source="title" label="Title" />

      <MarkdownField source="body" label="Body" showRaw={true} />

      <BooleanField source="is_urgent" />

      <BooleanField source="is_hidden" />
    </SimpleShowLayout>
  </Show>
);
