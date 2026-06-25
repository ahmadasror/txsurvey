import { createBrowserRouter } from "react-router-dom";
import { LoginPage } from "@/features/auth/LoginPage";
import { RequireAuth } from "@/features/auth/RequireAuth";
import { DashboardLayout } from "@/features/dashboard/DashboardLayout";
import { FormsListPage } from "@/features/dashboard/FormsListPage";
import { BuilderPage } from "@/features/builder/BuilderPage";
import { RunnerPage } from "@/features/runner/RunnerPage";

export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  // Public runner — anonymous, outside the auth guard.
  { path: "/r/:slug", element: <RunnerPage /> },
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
