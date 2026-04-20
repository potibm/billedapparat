import React from "react";
import {
  SimpleForm,
  TextInput,
  Create,
  required,
  SelectInput,
  CreateProps,
} from "react-admin";

export const SlideCreate: React.FC<CreateProps> = () => {
  return (
    <Create>
      <SimpleForm>
        <SelectInput
          source="content.type"
          label="Typ"
          validate={[required()]}
          choices={[
            { id: "sponsor", name: "Sponsor" },
            { id: "social", name: "Social Media" },
            { id: "news", name: "News / Ankündigung" },
          ]}
        />
        <TextInput
          source="author.displayName"
          label="Autor / Firmenname"
          fullWidth
        />
        <TextInput
          source="content.text"
          label="Textinhalt"
          multiline
          fullWidth
          resettable
        />
        <TextInput
          source="mediaUrlOriginal"
          label="Bild-URL (Optional)"
          fullWidth
        />
        <SelectInput
          source="status"
          label="Status"
          defaultValue="active"
          choices={[
            { id: "pending", name: "Wartet auf Freigabe" },
            { id: "active", name: "Aktiv (Wird angezeigt)" },
            { id: "hidden", name: "Versteckt" },
          ]}
        />
      </SimpleForm>
    </Create>
  );
};
