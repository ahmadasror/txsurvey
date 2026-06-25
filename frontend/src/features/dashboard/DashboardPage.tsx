import { LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useLogout, useMe } from "@/api/auth";

export function DashboardPage() {
  const { data: user } = useMe();
  const logout = useLogout();

  return (
    <div className="min-h-dvh bg-muted/30">
      <header className="flex items-center justify-between border-b bg-background px-6 py-4">
        <span className="text-lg font-semibold">txsurvey</span>
        <div className="flex items-center gap-3">
          {user && <span className="text-sm text-muted-foreground">{user.email}</span>}
          <Button variant="outline" size="sm" onClick={() => logout.mutate()} disabled={logout.isPending}>
            <LogOut /> Sign out
          </Button>
        </div>
      </header>

      <main className="container py-10">
        <Card>
          <CardHeader>
            <CardTitle>Your forms</CardTitle>
            <CardDescription>Phase 2 will list and let you build forms here.</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Signed in as <span className="font-medium text-foreground">{user?.name || user?.email}</span>.
            </p>
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
