import { ImagePreviewField } from "@admin/components/fields/ImagePreviewField";
import { Show, SimpleShowLayout, TextField, DateField } from "react-admin";
import { Typography, Box, Divider } from "@mui/material";

export const SocialMediaShow = () => (
  <Show title="View Social Media">
    <SimpleShowLayout>
      <Typography variant="h6" gutterBottom>
        Content
      </Typography>
      <TextField source="id" label="ID" />
      <TextField source="content.title" label="Title" />
      <TextField source="content.body" label="Content" />
      <ImagePreviewField
        source="content.media.local_url"
        label="Media"
        maxWidth={480}
        maxHeight={270}
      />

      <Divider sx={{ my: 2 }} />

      <Typography variant="h6" gutterBottom>
        Author
      </Typography>
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          gap: 3,
          p: 2,
          bgcolor: "action.hover",
          borderRadius: 2,
          border: "1px solid",
          borderColor: "divider",
        }}
      >
        <ImagePreviewField
          source="author.avatar.local_url"
          label="Avatar"
          maxWidth={80}
          maxHeight={80}
        />
        <Box>
          <TextField
            source="author.display_name"
            label="Name"
            emptyText="Unknown Author"
            sx={{ fontWeight: "bold", display: "block" }}
          />
          <TextField
            source="author.external_id"
            label="External ID"
            emptyText="No ID available"
            sx={{ fontSize: "0.8rem", opacity: 0.8 }}
          />
        </Box>
      </Box>

      <Divider sx={{ my: 2 }} />

      <Typography variant="h6" gutterBottom>
        Metadata
      </Typography>
      <TextField
        source="content.language"
        label="Language"
        emptyText="Not specified"
      />
      <TextField source="status" label="Status" />
      <DateField
        source="origin_created_at"
        label="Created at"
        showTime
        emptyText="-"
      />
    </SimpleShowLayout>
  </Show>
);
