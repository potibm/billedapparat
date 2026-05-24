import { Menu } from "react-admin";
import {
  Box,
  Typography,
  Divider,
  MenuItem,
  ListItemIcon,
  ListItemText,
} from "@mui/material";
import LaunchIcon from "@mui/icons-material/Launch";

const MenuGroup = ({ title }: { title: string }) => (
  <Box sx={{ mt: 2, mb: 1, pl: 2 }}>
    <Typography variant="caption" color="textSecondary">
      {title}
    </Typography>
  </Box>
);

const MenuExternalLink = ({ href, label }: { href: string; label: string }) => (
  <MenuItem component="a" href={href} target="_blank" sx={{ pl: 2 }}>
    <ListItemIcon>
      <LaunchIcon fontSize="small" />
    </ListItemIcon>
    <ListItemText primary={label} />
  </MenuItem>
);

const MenuDiv = () => <Divider sx={{ my: 1 }} />;

export const MyMenu = () => {
  return (
    <Menu>
      <Menu.ResourceItem name="sponsor-slides" />
      <MenuDiv />

      {/* --- GROUP: SOCIAL --- */}
      <MenuGroup title="Social" />
      <Menu.ResourceItem name="social-medias-slides" />
      <Menu.ResourceItem name="social-text-slides" />
      <Menu.ResourceItem name="filter-rules" />
      <MenuDiv />

      {/* --- GROUP: TIMETABLE --- */}
      <MenuGroup title="Timetable" />
      <MenuExternalLink href="https://dein-tidsapparat.url" label="Edit" />
      <Menu.ResourceItem name="timetable" />
      <Menu.ResourceItem name="timetable-slides" />
      <MenuDiv />

      {/* --- GROUP: NEWS --- */}
      <MenuGroup title="News" />
      <MenuExternalLink href="https://dein-protokolapparat.url" label="Edit" />
      <Menu.ResourceItem name="news" />
      <Menu.ResourceItem name="news-slides" />
      <MenuDiv />
    </Menu>
  );
};
