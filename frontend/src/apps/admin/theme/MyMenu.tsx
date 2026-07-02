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
import { useAppConfig } from "@core/config/useConfig";

const MenuGroup = ({ title }: { title: string }) => (
  <Box sx={{ mt: 2, mb: 1, pl: 2 }}>
    <Typography variant="caption" color="textSecondary">
      {title}
    </Typography>
  </Box>
);

const MenuExternalLink = ({ href, label }: { href: string; label: string }) => (
  <MenuItem
    component="a"
    href={href}
    target="_blank"
    rel="noopener noreferrer"
    sx={{ pl: 2 }}
  >
    <ListItemIcon>
      <LaunchIcon fontSize="small" />
    </ListItemIcon>
    <ListItemText primary={label} />
  </MenuItem>
);

const MenuDiv = () => <Divider sx={{ my: 1 }} />;

export const MyMenu = () => {
  const { admin_urls, version } = useAppConfig();

  const timetableURL = admin_urls?.timetable || null;
  const newsURL = admin_urls?.news || null;

  return (
    <Box sx={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <Box sx={{ flex: 1 }}>
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
          {timetableURL && (
            <MenuExternalLink href={timetableURL} label="Edit" />
          )}
          <Menu.ResourceItem name="timetable-slides" />
          <Menu.ResourceItem name="timetable" />
          <MenuDiv />

          {/* --- GROUP: NEWS --- */}
          <MenuGroup title="News" />
          {newsURL && <MenuExternalLink href={newsURL} label="Edit" />}
          <Menu.ResourceItem name="news-slides" />
          <Menu.ResourceItem name="news" />
          <MenuDiv />
        </Menu>
      </Box>

      <Box sx={{ p: 2, textAlign: "center" }}>
        <Typography variant="caption" color="text.secondary">
          Version: {version}
        </Typography>
      </Box>
    </Box>
  );
};
