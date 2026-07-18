import { lazy } from "react";
import { createBrowserRouter } from "react-router-dom";
import { routerBasename } from "@/lib/paths";
import { RequireAuth } from "@/features/auth/RequireAuth";

const LoginPage = lazy(() => import("@/features/auth/LoginPage").then((module) => ({ default: module.LoginPage })));
const LegalPage = lazy(() => import("@/features/legal/LegalPage").then((module) => ({ default: module.LegalPage })));
const RunnerPage = lazy(() => import("@/features/runner/RunnerPage").then((module) => ({ default: module.RunnerPage })));
const DashboardLayout = lazy(() => import("@/features/dashboard/DashboardLayout").then((module) => ({ default: module.DashboardLayout })));
const FormsListPage = lazy(() => import("@/features/dashboard/FormsListPage").then((module) => ({ default: module.FormsListPage })));
const TemplatesPage = lazy(() => import("@/features/templates/TemplatesPage").then((module) => ({ default: module.TemplatesPage })));
const BuilderPage = lazy(() => import("@/features/builder/BuilderPage").then((module) => ({ default: module.BuilderPage })));
const ResultsPage = lazy(() => import("@/features/results/ResultsPage").then((module) => ({ default: module.ResultsPage })));
const PublicLayout = lazy(() => import("@/features/marketing/PublicLayout").then((module) => ({ default: module.PublicLayout })));
const PublicTemplatesPage = lazy(() => import("@/features/marketing/PublicTemplatesPage").then((module) => ({ default: module.PublicTemplatesPage })));
const TemplatePreviewPage = lazy(() => import("@/features/marketing/TemplatePreviewPage").then((module) => ({ default: module.TemplatePreviewPage })));
const FeaturePage = lazy(() => import("@/features/marketing/FeaturePage").then((module) => ({ default: module.FeaturePage })));
const GuidePage = lazy(() => import("@/features/marketing/GuidePage").then((module) => ({ default: module.GuidePage })));
const LandingPage = lazy(() => import("@/features/marketing/LandingPage").then((module) => ({ default: module.LandingPage })));
const NotFoundPage = lazy(() => import("@/features/marketing/NotFoundPage").then((module) => ({ default: module.NotFoundPage })));

export const router = createBrowserRouter(
  [
    { path: "/login", element: <LoginPage /> },
    // Legal docs — public, outside the auth guard.
    { path: "/legal", element: <LegalPage /> },
    // Public runner — anonymous, outside the auth guard.
    { path: "/r/:slug", element: <RunnerPage /> },
    {
      element: <PublicLayout />,
      children: [
        { index: true, element: <LandingPage /> },
        { path: "/contoh-template-survei", element: <PublicTemplatesPage /> },
        { path: "/template-survei-kepuasan-karyawan", element: <TemplatePreviewPage /> },
        { path: "/template-feedback-onboarding", element: <TemplatePreviewPage /> },
        { path: "/template-kepuasan-layanan-it", element: <TemplatePreviewPage /> },
        { path: "/fitur/:featureSlug", element: <FeaturePage /> },
        { path: "/panduan", element: <GuidePage /> },
      ],
    },
    {
      element: <RequireAuth />,
      children: [
        {
          element: <DashboardLayout />,
          children: [
            { path: "/app", element: <FormsListPage /> },
            { path: "/templates", element: <TemplatesPage /> },
          ],
        },
        // The builder and results are full-bleed (own header), outside dashboard chrome.
        { path: "/forms/:id", element: <BuilderPage /> },
        { path: "/forms/:id/results", element: <ResultsPage /> },
      ],
    },
    { path: "*", element: <NotFoundPage /> },
  ],
  // basename carries the deploy subpath (e.g. "/txsurvey"); undefined at root.
  { basename: routerBasename },
);
