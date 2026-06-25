import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { FileText, Loader2, Plus, Sparkles, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ThemePicker } from "@/components/ThemePicker";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { api } from "@/api/client";
import { DEFAULT_THEME_ID } from "@/lib/themes";
import { useCreateForm, useDeleteForm, useForms } from "@/api/forms";
import type { FormListItem, FormStatus } from "@/types/forms";

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

  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [preset, setPreset] = useState(DEFAULT_THEME_ID);
  const [creating, setCreating] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<FormListItem | null>(null);

  const openCreate = () => {
    setTitle("");
    setPreset(DEFAULT_THEME_ID);
    setOpen(true);
  };

  const submit = async () => {
    if (creating) return;
    setCreating(true);
    try {
      const form = await createForm.mutateAsync(title.trim() || "Untitled survey");
      await api(`/forms/${form.id}`, {
        method: "PATCH",
        body: JSON.stringify({
          title: form.title,
          description: "",
          settings: { show_progress: true, theme: { preset } },
        }),
      });
      navigate(`/forms/${form.id}`);
    } finally {
      setCreating(false);
    }
  };

  return (
    <main className="container max-w-5xl py-10">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Your surveys</h1>
        <Button onClick={openCreate}>
          <Plus /> New survey
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
            <p className="text-muted-foreground">No surveys yet. Create your first one.</p>
            <Button onClick={openCreate}>
              <Plus /> New survey
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
                    setDeleteTarget(f);
                  }}
                  aria-label="Delete survey"
                >
                  <Trash2 className="text-destructive" />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New survey</DialogTitle>
            <DialogDescription>Give it a title and pick a theme — you can change both later.</DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label>Title</Label>
            <Input
              autoFocus
              value={title}
              placeholder="e.g. Customer feedback"
              onChange={(e) => setTitle(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") submit();
              }}
            />
          </div>
          <div className="space-y-2">
            <Label>Theme</Label>
            <ThemePicker value={preset} onChange={setPreset} />
          </div>
          <DialogFooter>
            <Button onClick={submit} disabled={creating}>
              {creating ? <Loader2 className="animate-spin" /> : <Sparkles />} Create survey
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(o) => !o && setDeleteTarget(null)}
        title="Delete this survey?"
        description={
          deleteTarget
            ? `"${deleteTarget.title}" and its ${deleteTarget.response_count} response${deleteTarget.response_count === 1 ? "" : "s"} will be permanently deleted.`
            : undefined
        }
        confirmText="Delete"
        destructive
        onConfirm={() => {
          if (deleteTarget) deleteForm.mutate(deleteTarget.id);
          setDeleteTarget(null);
        }}
      />
    </main>
  );
}
