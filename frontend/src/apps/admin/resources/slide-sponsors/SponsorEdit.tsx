import { ImageEditPreviewField } from "@admin/components/fields/ImageEditPreviewField";
import { MediaUploadInput } from "@admin/components/inputs/MediaUploadInput";
import { PriorityInput } from "@admin/components/inputs/PriorityInput";
import { StatusSelectInput } from "@admin/components/inputs/StatusSelectInput";
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

      <MediaUploadInput label="Change Slide" />

      <PriorityInput />

      <NumberInput
        source="display_options.duration"
        label="Duration in seconds (0 = use playlist)"
        step={0.5}
        min={0}
      />

      <BooleanInput
        source="display_options.allow_social_overlay"
        label="Allow social media overlay"
      />

      <StatusSelectInput />
    </SimpleForm>
  </Edit>
);
