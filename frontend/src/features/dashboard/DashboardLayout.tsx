import { Link, Outlet } from "react-router-dom";
import { LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useLogout, useMe } from "@/api/auth";

export function DashboardLayout() {
  const { data: user } = useMe();
  const logout = useLogout();

  return (
    <div className="min-h-dvh bg-muted/30">
      <header className="sticky top-0 z-10 flex items-center justify-between border-b bg-background px-6 py-3">
        <Link to="/" className="text-lg font-semibold">
          txsurvey
        </Link>
        <div className="flex items-center gap-3">
          {user && <span className="hidden text-sm text-muted-foreground sm:inline">{user.email}</span>}
          <Button variant="outline" size="sm" onClick={() => logout.mutate()} disabled={logout.isPending}>
            <LogOut /> Sign out
          </Button>
        </div>
      </header>
      <Outlet />
    </div>
  );
}
