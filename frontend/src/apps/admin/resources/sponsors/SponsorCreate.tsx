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

export const SponsorCreate = () => (
  <Create title="Add Sponsor Screen">
    <SimpleForm
      defaultValues={{ content: { type: "sponsor" }, status: "active" }}
    >
      <TextInput
        source="content.text"
        label="Name of the Sponsor"
        validate={[required()]}
        fullWidth
      />

      <ImageUploadInput label="Upload Slide" />

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
