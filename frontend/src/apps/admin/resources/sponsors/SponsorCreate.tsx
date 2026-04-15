import {
  Create,
  SimpleForm,
  TextInput,
  ImageInput,
  ImageField,
  SelectInput,
  required,
  NumberInput,
} from "react-admin";

export const SponsorCreate = () => (
  <Create title="Neuen Sponsor hinzufügen">
    <SimpleForm
      defaultValues={{ content: { type: "sponsor" }, status: "active" }}
    >
      <TextInput
        source="content.text"
        label="Name des Sponsors"
        validate={[required()]}
        fullWidth
      />

      {/* Das Drag & Drop Feld */}
      <ImageInput
        source="image_upload"
        label="Logo hochladen"
        placeholder={<p>Logo hierher ziehen oder klicken</p>}
      >
        <ImageField source="src" title="title" />
      </ImageInput>

      <NumberInput
        source="display_options.priority"
        label="Priorität"
        defaultValue={1}
        min={1}
        max={10}
        step={1}
        helperText="Höhere Zahl = höhere Sichtbarkeit (z.B. 10 für Hauptsponsoren, 0 für Standard)"
      />

      <SelectInput
        source="status"
        choices={[
          { id: "active", name: "Aktiv" },
          { id: "hidden", name: "Ausgeblendet" },
        ]}
      />
    </SimpleForm>
  </Create>
);

export default SponsorCreate;
