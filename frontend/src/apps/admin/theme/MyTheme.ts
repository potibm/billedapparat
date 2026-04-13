import { defaultTheme } from "react-admin";
import type { ThemeOptions } from "@mui/material";

export const MyTheme: ThemeOptions = {
  ...defaultTheme,
  palette: {
    ...defaultTheme.palette,
    primary: {
      main: "#00838F",
    },
    secondary: {
      main: "#00F5FF",
      contrastText: "#111827",
    },
    error: {
      main: "#FF3366",
    },
  },
  components: {
    ...defaultTheme.components,
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: "4px",
          fontWeight: 600,
        },
      },
    },
  },
};

export const MyDarkTheme: ThemeOptions = {
  ...defaultTheme,
  palette: {
    mode: "dark",
    primary: {
      main: "#00F5FF",
    },
    secondary: {
      main: "#111827",
      contrastText: "#00F5FF",
    },
    error: {
      main: "#FF3366",
    },
    background: {
      default: "#0f172a",
      paper: "#1e293b",
    },
  },
  components: {
    ...defaultTheme.components,
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: "4px",
          fontWeight: 600,
        },
      },
    },
  },
};
