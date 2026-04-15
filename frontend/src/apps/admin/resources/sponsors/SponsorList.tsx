import {
  List,
  Datagrid,
  TextField,
  ImageField,
  EditButton,
  DeleteButton,
  ChipField,
  FunctionField,
  NumberField,
} from "react-admin";

export const SponsorList = () => (
  <List title="Sponsoren">
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <ImageField
        source="media_url_original"
        label="Logo"
        sx={{
          "& img": {
            maxWidth: 60,
            maxHeight: 60,
            objectFit: "contain",
            borderRadius: "4px",
            backgroundColor: "#f3f4f6", // Leichter Hintergrund, falls das Logo weiß/transparent ist
            padding: "4px",
          },
        }}
      />

      <TextField source="content.text" label="Name" />

      <NumberField source="display_options.priority" label="Prio" />

      <FunctionField
        label="Status"
        render={(record: any) => (
          <ChipField
            source="status"
            record={record}
            sx={{
              backgroundColor:
                record.status === "active" ? "#d1fae5" : "#f3f4f6",
              color: record.status === "active" ? "#065f46" : "#374151",
              fontWeight: "bold",
            }}
          />
        )}
      />

      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export default SponsorList;
