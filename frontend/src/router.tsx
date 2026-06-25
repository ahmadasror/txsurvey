import { createBrowserRouter } from "react-router-dom";
import { LoginPage } from "@/features/auth/LoginPage";
import { RequireAuth } from "@/features/auth/RequireAuth";
import { DashboardLayout } from "@/features/dashboard/DashboardLayout";
import { FormsListPage } from "@/features/dashboard/FormsListPage";
import { BuilderPage } from "@/features/builder/BuilderPage";

export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    element: <RequireAuth />,
    children: [
      {
        element: <DashboardLayout />,
        children: [{ index: true, path: "/", element: <FormsListPage /> }],
      },
      // The builder is full-bleed (its own header), outside the dashboard chrome.
      { path: "/forms/:id", element: <BuilderPage /> },
    ],
  },
]);
