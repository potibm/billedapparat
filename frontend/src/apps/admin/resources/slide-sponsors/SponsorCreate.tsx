import { MediaUploadInput } from "@admin/components/inputs/MediaUploadInput";
import { PriorityInput } from "@admin/components/inputs/PriorityInput";
import { StatusSelectInput } from "@admin/components/inputs/StatusSelectInput";
import {
  Create,
  SimpleForm,
  TextInput,
  required,
  BooleanInput,
  NumberInput,
} from "react-admin";

export const SponsorCreate = () => (
  <Create title="Add Sponsor Screen">
    <SimpleForm
      defaultValues={{
        content: { type: "sponsor" },
        status: "active",
        display_options: { duration: 0 },
      }}
    >
      <TextInput
        source="content.title"
        label="Name of the Sponsor"
        validate={[required()]}
        fullWidth
      />

      <MediaUploadInput label="Upload Slide" />

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
        defaultValue={true}
      />

      <StatusSelectInput />
    </SimpleForm>
  </Create>
);

export default SponsorCreate;
