import NotFound from "@splash/pages/NotFound";
import Splash from "@splash/pages/Splash";
import React, { useMemo } from "react";
import {
  createBrowserRouter,
  RouterProvider,
  RouteObject,
} from "react-router-dom"; // react-router-dom is the standard for web projects
import { AdminApp } from "../apps/admin/Admin";
import { BeamerApp } from "../apps/beamer/BeamerApp";

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
        path: "/beamer/:id?",
        element: <BeamerApp />,
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
