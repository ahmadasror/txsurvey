import { Navigate } from "react-router-dom";
import { LogIn } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { FullScreenLoader } from "@/components/FullScreenLoader";
import { loginUrl } from "@/api/client";
import { useMe } from "@/api/auth";

export function LoginPage() {
  const { data: user, isLoading } = useMe();

  if (isLoading) return <FullScreenLoader />;
  if (user) return <Navigate to="/" replace />;

  return (
    <div className="flex min-h-dvh items-center justify-center bg-muted/30 p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle>txsurvey</CardTitle>
          <CardDescription>Build conversational forms. Sign in to start.</CardDescription>
        </CardHeader>
        <CardContent>
          <Button className="w-full" size="lg" onClick={() => (window.location.href = loginUrl())}>
            <LogIn /> Continue with Google
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
