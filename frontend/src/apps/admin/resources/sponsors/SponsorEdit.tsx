import {
  Edit,
  SimpleForm,
  TextInput,
  ImageInput,
  ImageField,
  SelectInput,
  required,
  useRecordContext,
  NumberInput,
} from "react-admin";

// Eine kleine Hilfskomponente, um das aktuelle Logo schön anzuzeigen
const CurrentLogoField = () => {
  const record = useRecordContext();
  if (!record || !record.media_url_original) return null;
  return (
    <div className="mb-4">
      <p className="text-gray-400 text-xs mb-1">Current Logo</p>
      <img
        src={record.media_url_original}
        alt="Current Logo"
        className="max-w-[200px] max-h-[100px] ml-1.5 object-contain bg-gray-100 p-2 rounded"
      />
    </div>
  );
};

export const SponsorEdit = () => (
  <Edit title="Sponsor bearbeiten">
    <SimpleForm>
      <TextInput
        source="content.text"
        label="Name"
        validate={[required()]}
        fullWidth
      />

      <CurrentLogoField />

      <ImageInput
        source="image_upload"
        label="Upload new sponsor slide (overwrite existing)"
        placeholder={<p>Drag and drop a new image here</p>}
      >
        <ImageField source="src" title="title" />
      </ImageInput>

      <NumberInput
        source="display_options.priority"
        label="Priority"
        defaultValue={1}
        min={1}
        max={10}
        step={1}
        helperText="Higher number = higher visibility (e.g., 10 for main sponsors, 1 for standard)"
      />

      <SelectInput
        source="status"
        choices={[
          { id: "active", name: "Active" },
          { id: "hidden", name: "Hidden" },
        ]}
      />
    </SimpleForm>
  </Edit>
);
