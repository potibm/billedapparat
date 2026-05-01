import { Admin, Resource } from "react-admin";
import sponsors from "./resources/sponsors";
import news from "./resources/news";
import socialMedia from "./resources/social-media";
import socialText from "./resources/social-text";
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
    <Resource name="news" {...news} />
    <Resource name="social-media" {...socialMedia} />
    <Resource name="social-text" {...socialText} />
  </Admin>
);
