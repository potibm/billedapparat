import { MediaUploadInput } from "@admin/components/inputs/MediaUploadInput";
import { PriorityInput } from "@admin/components/inputs/PriorityInput";
import { StatusSelectInput } from "@admin/components/inputs/StatusSelectInput";
import {
  Create,
  SimpleForm,
  TextInput,
  required,
  BooleanInput,
} from "react-admin";

export const SponsorCreate = () => (
  <Create title="Add Sponsor Screen">
    <SimpleForm
      defaultValues={{ content: { type: "sponsor" }, status: "active" }}
    >
      <TextInput
        source="content.title"
        label="Name of the Sponsor"
        validate={[required()]}
        fullWidth
      />

      <MediaUploadInput label="Upload Slide" />

      <PriorityInput />

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
