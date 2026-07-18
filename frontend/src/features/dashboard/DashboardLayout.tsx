import { useState } from "react";
import { Link, NavLink, Outlet } from "react-router-dom";
import { LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { BrandMark } from "@/components/BrandMark";
import { cn } from "@/lib/utils";
import { useLogout, useMe } from "@/api/auth";

export function DashboardLayout() {
  const { data: user } = useMe();
  const logout = useLogout();

  return (
    <div className="min-h-dvh bg-background">
      <header className="sticky top-0 z-10 border-b bg-card">
        <div className="mx-auto flex h-[60px] max-w-[980px] items-center gap-6 px-6">
          <Link to="/app" className="flex items-center gap-2.5">
            <BrandMark size={30} />
            <span className="font-display text-lg text-foreground">txsurvey</span>
          </Link>
          <nav className="flex items-center gap-1 text-sm">
            <NavItem to="/app" end>
              Survei
            </NavItem>
            <NavItem to="/templates">Template</NavItem>
          </nav>
          <div className="ml-auto flex items-center gap-3">
            {user && <span className="hidden text-sm text-muted-foreground sm:inline">{user.email}</span>}
            <Avatar user={user} />
            <Button
              variant="ghost"
              size="icon"
              onClick={() => logout.mutate()}
              disabled={logout.isPending}
              aria-label="Keluar"
            >
              <LogOut />
            </Button>
          </div>
        </div>
      </header>
      <Outlet />
    </div>
  );
}

function NavItem({ to, end, children }: { to: string; end?: boolean; children: React.ReactNode }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        cn(
          "rounded-lg px-3 py-1.5 font-medium transition-colors",
          isActive ? "bg-primary-soft text-primary" : "text-muted-foreground hover:text-foreground",
        )
      }
    >
      {children}
    </NavLink>
  );
}

function Avatar({ user }: { user?: { name: string; email: string; picture_url: string } | null }) {
  const [broken, setBroken] = useState(false);
  if (user?.picture_url && !broken) {
    return (
      <img
        src={user.picture_url}
        alt=""
        // Google's lh3 CDN rejects hot-linked requests that carry a referrer.
        referrerPolicy="no-referrer"
        onError={() => setBroken(true)}
        className="size-8 rounded-full border object-cover"
      />
    );
  }
  const initial = (user?.name || user?.email || "?").trim().charAt(0).toUpperCase();
  return (
    <span className="grid size-8 place-items-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">
      {initial}
    </span>
  );
}
