import { createBrowserRouter } from "react-router-dom";
import { LoginPage } from "@/features/auth/LoginPage";
import { RequireAuth } from "@/features/auth/RequireAuth";
import { DashboardLayout } from "@/features/dashboard/DashboardLayout";
import { FormsListPage } from "@/features/dashboard/FormsListPage";
import { BuilderPage } from "@/features/builder/BuilderPage";
import { ResultsPage } from "@/features/results/ResultsPage";
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
      // The builder and results are full-bleed (own header), outside dashboard chrome.
      { path: "/forms/:id", element: <BuilderPage /> },
      { path: "/forms/:id/results", element: <ResultsPage /> },
    ],
  },
]);
