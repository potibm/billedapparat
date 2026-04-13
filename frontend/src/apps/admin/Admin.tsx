import { Admin, Resource } from "react-admin";
import jsonServerProvider from "ra-data-json-server";
import slides from "./resources/slides";
import { MyTheme, MyDarkTheme } from "./theme/MyTheme";
import { MyLayout } from "./theme/MyLayout";

const dataProvider = jsonServerProvider("/api/admin");

export const AdminApp = () => (
  <Admin
    basename="/admin"
    dataProvider={dataProvider}
    theme={MyTheme}
    darkTheme={MyDarkTheme}
    layout={MyLayout}
    title="Billedapparat Admin"
  >
    <Resource name="slides" {...slides} />
  </Admin>
);
