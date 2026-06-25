import { Navigate, Outlet } from "react-router-dom";
import { FullScreenLoader } from "@/components/FullScreenLoader";
import { useMe } from "@/api/auth";

/** RequireAuth guards creator routes: redirects to /login when not signed in. */
export function RequireAuth() {
  const { data: user, isLoading } = useMe();
  if (isLoading) return <FullScreenLoader />;
  if (!user) return <Navigate to="/login" replace />;
  return <Outlet />;
}
