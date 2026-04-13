import NotFound from "@splash/pages/NotFound";
import Splash from "@splash/pages/Splash";
import React, { useMemo } from "react";
import {
  createBrowserRouter,
  RouterProvider,
  RouteObject,
} from "react-router-dom"; // react-router-dom ist für Webprojekte Standard
import { AdminApp } from "../apps/admin/Admin";

export const Router: React.FC = () => {
  const router = useMemo(() => {
    const routes: RouteObject[] = [
      {
        path: "/",
        element: <Splash />,
      },
      {
        path: "/admin/*",
        element: <AdminApp />,
      },
      {
        path: "*",
        element: <NotFound />,
      },
    ];

    return createBrowserRouter(routes);
  }, []);

  return <RouterProvider router={router} />;
};

export default Router;
