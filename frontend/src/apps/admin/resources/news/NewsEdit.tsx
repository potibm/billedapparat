import { ImageEditPreviewField } from "@admin/components/fields/ImageEditPreviewField";
import { ImageUploadInput } from "@admin/components/fields/ImageUploadInput";
import { PriorityInput } from "@admin/components/fields/PriorityInput";
import { StatusSelectInput } from "@admin/components/fields/StatusSelectInput";
import {
  Edit,
  SimpleForm,
  TextInput,
  required,
  BooleanInput,
} from "react-admin";

export const NewsEdit = () => (
  <Edit title="Edit News Screens">
    <SimpleForm>
      <TextInput
        source="content.title"
        label="Title"
        validate={[required()]}
        fullWidth
      />

      <TextInput
        source="content.body"
        label="Body"
        validate={[required()]}
        fullWidth
        multiline
      />

      <PriorityInput />

      <StatusSelectInput />
    </SimpleForm>
  </Edit>
);
