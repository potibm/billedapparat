import React from "react";
import {
  SimpleForm,
  TextInput,
  SelectInput,
  Edit,
  EditProps,
} from "react-admin";

export const SlideEdit: React.FC<EditProps> = () => {
  return (
    <Edit>
      <SimpleForm>
        {/* TextInput mit source="id" ist meist readOnly, 
          damit man den Primärschlüssel nicht versehentlich ändert */}
        <TextInput source="id" disabled />

        <SelectInput
          source="content.type"
          label="Typ"
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
        />
        <TextInput source="mediaUrlOriginal" label="Bild-URL" fullWidth />
        <SelectInput
          source="status"
          label="Status"
          choices={[
            { id: "pending", name: "Wartet auf Freigabe" },
            { id: "active", name: "Aktiv" },
            { id: "hidden", name: "Versteckt" },
          ]}
        />
      </SimpleForm>
    </Edit>
  );
};
