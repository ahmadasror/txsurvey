import { useNavigate } from "react-router-dom";
import { FileText, Plus, Trash2, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useCreateForm, useDeleteForm, useForms } from "@/api/forms";
import type { FormStatus } from "@/types/forms";

const statusVariant: Record<FormStatus, "success" | "muted" | "secondary"> = {
  published: "success",
  draft: "muted",
  closed: "secondary",
};

export function FormsListPage() {
  const navigate = useNavigate();
  const { data: forms, isLoading } = useForms();
  const createForm = useCreateForm();
  const deleteForm = useDeleteForm();

  const onCreate = () =>
    createForm.mutate("Untitled form", { onSuccess: (form) => navigate(`/forms/${form.id}`) });

  return (
    <main className="container max-w-5xl py-10">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Your forms</h1>
        <Button onClick={onCreate} disabled={createForm.isPending}>
          {createForm.isPending ? <Loader2 className="animate-spin" /> : <Plus />} New form
        </Button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </div>
      ) : !forms || forms.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center gap-3 py-16 text-center">
            <FileText className="size-10 text-muted-foreground" />
            <p className="text-muted-foreground">No forms yet. Create your first one.</p>
            <Button onClick={onCreate} disabled={createForm.isPending}>
              <Plus /> New form
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {forms.map((f) => (
            <Card
              key={f.id}
              className="cursor-pointer transition-shadow hover:shadow-md"
              onClick={() => navigate(`/forms/${f.id}`)}
            >
              <CardHeader className="flex-row items-start justify-between space-y-0">
                <div className="min-w-0">
                  <CardTitle className="truncate text-lg">{f.title}</CardTitle>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {f.question_count} question{f.question_count === 1 ? "" : "s"} · {f.response_count} response
                    {f.response_count === 1 ? "" : "s"}
                  </p>
                </div>
                <Badge variant={statusVariant[f.status]}>{f.status}</Badge>
              </CardHeader>
              <CardContent className="flex justify-end pt-0">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={(e) => {
                    e.stopPropagation();
                    if (confirm(`Delete "${f.title}"?`)) deleteForm.mutate(f.id);
                  }}
                  aria-label="Delete form"
                >
                  <Trash2 className="text-destructive" />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </main>
  );
}
