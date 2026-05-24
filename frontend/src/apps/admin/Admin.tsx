import { Admin, Resource } from "react-admin";
import sponsors from "./resources/slide-sponsors";
import slideNews from "./resources/slide-news";
import news from "./resources/news";
import socialMedia from "./resources/slide-social-media";
import socialText from "./resources/slide-social-text";
import filterRules from "./resources/filter-rules";
import slideTimetable from "./resources/slide-timetable";
import timetable from "./resources/timetable";
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
    <Resource name="sponsor-slides" {...sponsors} />
    <Resource name="social-medias-slides" {...socialMedia} />
    <Resource name="social-text-slides" {...socialText} />
    <Resource name="news-slides" {...slideNews} />
    <Resource name="news" {...news} />
    <Resource name="timetable-slides" {...slideTimetable} />
    <Resource name="timetable" {...timetable} />
    <Resource name="filter-rules" {...filterRules} />
  </Admin>
);
