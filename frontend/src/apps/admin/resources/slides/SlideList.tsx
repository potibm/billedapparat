import {
  List,
  Datagrid,
  TextField,
  DateField,
  SelectField,
  EditButton,
  DeleteButton,
} from "react-admin";

export const SlideList = () => (
  <List>
    <Datagrid rowClick="edit">
      <TextField source="id" />
      <SelectField
        source="content.type"
        label="Typ"
        choices={[
          { id: "sponsor", name: "Sponsor" },
          { id: "social", name: "Social Media" },
          { id: "news", name: "News" },
        ]}
      />
      <TextField source="author.displayName" label="Autor" />
      <TextField source="status" />
      <DateField source="origin_created_at" label="Datum" showTime />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);
