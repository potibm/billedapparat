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

export const SponsorEdit = () => (
  <Edit title="Edit Sponsor Screens">
    <SimpleForm>
      <TextInput
        source="content.text"
        label="Name"
        validate={[required()]}
        fullWidth
      />

      <ImageEditPreviewField />

      <ImageUploadInput label="Change Slide" />

      <PriorityInput />

      <BooleanInput
        source="content.allowSocialOverlay"
        label="Allow social media overlay"
      />

      <StatusSelectInput />
    </SimpleForm>
  </Edit>
);
