import { ImageUploadInput } from "@admin/components/fields/ImageUploadInput";
import { PriorityInput } from "@admin/components/fields/PriorityInput";
import { StatusSelectInput } from "@admin/components/fields/StatusSelectInput";
import {
  Create,
  SimpleForm,
  TextInput,
  required,
  BooleanInput,
} from "react-admin";

export const NewsCreate = () => (
  <Create title="Add News Screen">
    <SimpleForm
      defaultValues={{ content: { type: "news" }, status: "active" }}
    >
      <TextInput
        source="content.title"
        label="Title"
        validate={[required()]}
        fullWidth
      />

      <TextInput
        source="content.body"
        label="Text"
        validate={[required()]}
        fullWidth
        multiline
      />

      <PriorityInput />

      <StatusSelectInput />
    </SimpleForm>
  </Create>
);

export default NewsCreate;
