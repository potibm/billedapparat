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
  NumberInput,
} from "react-admin";

export const SponsorEdit = () => (
  <Edit title="Edit Sponsor Screens">
    <SimpleForm>
      <NumberInput source="id" label="ID" disabled />

      <TextInput
        source="content.title"
        label="Name"
        validate={[required()]}
        fullWidth
      />

      <ImageEditPreviewField />

      <ImageUploadInput label="Change Slide" />

      <PriorityInput />

      <BooleanInput
        source="display_options.allow_social_overlay"
        label="Allow social media overlay"
      />

      <StatusSelectInput />
    </SimpleForm>
  </Edit>
);
