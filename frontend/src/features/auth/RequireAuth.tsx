import { Navigate, Outlet } from "react-router-dom";
import { FullScreenLoader } from "@/components/FullScreenLoader";
import { useMe } from "@/api/auth";
import { usePageMetadata } from "@/lib/usePageMetadata";

/** RequireAuth guards creator routes: redirects to /login when not signed in. */
export function RequireAuth() {
  usePageMetadata({
    description: "Area kerja creator txsurvey.",
    robots: "noindex, nofollow",
  });
  const { data: user, isLoading } = useMe();
  if (isLoading) return <FullScreenLoader />;
  if (!user) return <Navigate to="/login" replace />;
  return <Outlet />;
}
