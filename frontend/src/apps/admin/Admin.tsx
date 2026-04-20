import { Admin, Resource } from "react-admin";
import slides from "./resources/slides";
import sponsors from "./resources/sponsors";
import { MyTheme, MyDarkTheme } from "./theme/MyTheme";
import { MyLayout } from "./theme/MyLayout";
import { dataProvider } from "./providers/dataProvider";

export const AdminApp = () => (
  <Admin
    basename="/admin"
    dataProvider={dataProvider}
    theme={MyTheme}
    darkTheme={MyDarkTheme}
    layout={MyLayout}
    title="Billedapparat Admin"
  >
    <Resource name="sponsors" {...sponsors} />
    <Resource name="slides" {...slides} />
  </Admin>
);
