/* eslint-disable @eslint-react/set-state-in-effect */
import { useState, useRef, useEffect } from "react";
import { Admin, Resource } from "react-admin";
import { BrowserRouter } from "react-router-dom";
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
import { authProvider, configureOidc } from "./providers/authProvider";
import { OidcLogin } from "./components/auth/OidcLogin";
import { useAppConfig } from "@core/config/useConfig";

export const AdminApp = () => (
  <BrowserRouter>
    <AppBootstrapper />
  </BrowserRouter>
);

export const AppBootstrapper = () => {
  const appConfig = useAppConfig();
  const isOidcActive = appConfig.auth?.type === "oidc";

  const [isConfiguring, setIsConfiguring] = useState(isOidcActive);

  const hasConfiguredRef = useRef(false);

  useEffect(() => {
    if (!isOidcActive) {
      setIsConfiguring(false);
      return;
    }

    if (
      !hasConfiguredRef.current &&
      appConfig.auth?.authority &&
      appConfig.auth?.client_id
    ) {
      configureOidc(appConfig.auth.authority, appConfig.auth.client_id);
      hasConfiguredRef.current = true;

      setIsConfiguring(false);
    }
  }, [isOidcActive, appConfig.auth]);

  if (isConfiguring) {
    return null;
  }

  return (
    <Admin
      basename="/admin"
      loginPage={isOidcActive ? OidcLogin : false}
      dataProvider={dataProvider}
      authProvider={isOidcActive ? authProvider : undefined}
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
};
