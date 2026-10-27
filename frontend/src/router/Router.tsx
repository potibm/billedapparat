import NotFound from "@splash/pages/NotFound";
import Splash from "@splash/pages/Splash";
import React, { lazy, Suspense, useMemo } from "react";
import {
  createBrowserRouter,
  RouterProvider,
  RouteObject,
} from "react-router-dom"; // react-router-dom is the standard for web projects

const AdminApp = lazy(() => import("../apps/admin/Admin").then((m) => ({ default: m.AdminApp })));
const BeamerApp = lazy(() => import("../apps/beamer/BeamerApp").then((m) => ({ default: m.BeamerApp })));

const RouteFallback = () => (
  <div style={{ padding: 24, textAlign: "center" }}>Loading…</div>
);

export const Router: React.FC = () => {
  const router = useMemo(() => {
    const routes: RouteObject[] = [
      {
        path: "/",
        element: <Splash />,
      },
      {
        path: "/admin/*",
        element: (
          <Suspense fallback={<RouteFallback />}>
            <AdminApp />
          </Suspense>
        ),
      },
      {
        path: "/beamer/:id?",
        element: (
          <Suspense fallback={<RouteFallback />}>
            <BeamerApp />
          </Suspense>
        ),
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
