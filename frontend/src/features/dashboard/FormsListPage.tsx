import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { FileText, Loader2, Plus, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
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
import { cn } from "@/lib/utils";
import { api } from "@/api/client";
import { DEFAULT_THEME_ID } from "@/lib/themes";
import { useCreateForm, useDeleteForm, useForms } from "@/api/forms";
import type { FormListItem, FormStatus } from "@/types/forms";

const statusLabel: Record<FormStatus, string> = {
  published: "● Published",
  draft: "Draft",
  closed: "Closed",
};

const statusClass: Record<FormStatus, string> = {
  published: "bg-primary-soft text-primary",
  draft: "bg-muted text-muted-foreground",
  closed: "bg-brand/15 text-brand",
};

/** sevenBars derives a stable 7-value visual motif from the form id. It is
 *  ambient decoration (no axis, no label) — never presented as a real metric. */
function sevenBars(seed: string): number[] {
  let h = 0;
  for (let i = 0; i < seed.length; i++) h = (h * 31 + seed.charCodeAt(i)) >>> 0;
  return Array.from({ length: 7 }, (_, i) => {
    h = (h * 1103515245 + 12345) >>> 0;
    return 0.35 + ((h >> (i % 8)) % 65) / 100;
  });
}

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
      const form = await createForm.mutateAsync(title.trim() || "Survei tanpa judul");
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

  const totalResponses = (forms ?? []).reduce((n, f) => n + f.response_count, 0);

  return (
    <main className="mx-auto max-w-[980px] px-6 py-9">
      <div className="mb-7 flex items-end justify-between gap-4">
        <div>
          <h1 className="font-display text-[28px] leading-tight text-foreground">Surveimu</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {(forms ?? []).length} survei · {totalResponses} respons total
          </p>
        </div>
        <Button className="h-11 rounded-xl" onClick={openCreate}>
          <Plus /> Survei baru
        </Button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </div>
      ) : !forms || forms.length === 0 ? (
        <Card className="rounded-2xl border-dashed bg-card/50">
          <CardContent className="flex flex-col items-center gap-3 py-16 text-center">
            <FileText className="size-10 text-muted-foreground" />
            <p className="text-muted-foreground">Belum ada survei. Buat yang pertama.</p>
            <Button className="rounded-xl" onClick={openCreate}>
              <Plus /> Survei baru
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {forms.map((f) => {
            const isDraft = f.status === "draft";
            return (
              <div
                key={f.id}
                role="button"
                tabIndex={0}
                onClick={() => navigate(`/forms/${f.id}`)}
                onKeyDown={(e) => e.key === "Enter" && navigate(`/forms/${f.id}`)}
                className={cn(
                  "group cursor-pointer rounded-2xl border p-5 transition-all hover:-translate-y-0.5 hover:shadow-[0_12px_26px_rgba(29,38,36,0.10)]",
                  isDraft ? "border-dashed bg-background" : "border-border bg-card",
                )}
              >
                <div className="flex items-start justify-between gap-3">
                  <h2 className="min-w-0 truncate text-[17px] font-bold text-foreground">{f.title}</h2>
                  <span
                    className={cn(
                      "shrink-0 rounded-full px-2.5 py-0.5 text-xs font-medium",
                      statusClass[f.status],
                    )}
                  >
                    {statusLabel[f.status]}
                  </span>
                </div>
                <p className="mt-1 text-sm text-muted-foreground">
                  {f.question_count} pertanyaan · {f.response_count} respons
                </p>

                <div className="mt-4 flex h-9 items-end gap-1" aria-hidden>
                  {sevenBars(f.id).map((h, i) => (
                    <span
                      key={i}
                      className="flex-1 rounded-sm bg-primary"
                      style={{ height: `${Math.round(h * 100)}%`, opacity: 0.18 + h * 0.5 }}
                    />
                  ))}
                </div>

                <div className="mt-4 flex justify-end">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setDeleteTarget(f);
                    }}
                    className="text-sm font-medium text-muted-foreground transition-colors hover:text-destructive"
                  >
                    Hapus
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="font-display text-2xl">Survei baru</DialogTitle>
            <DialogDescription>Beri judul dan pilih tema — keduanya bisa diubah nanti.</DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label>Judul</Label>
            <Input
              autoFocus
              value={title}
              placeholder="mis. Feedback pelanggan"
              onChange={(e) => setTitle(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") submit();
              }}
            />
          </div>
          <div className="space-y-2">
            <Label>Tema</Label>
            <ThemePicker value={preset} onChange={setPreset} />
          </div>
          <DialogFooter>
            <Button className="rounded-xl" onClick={submit} disabled={creating}>
              {creating ? <Loader2 className="animate-spin" /> : <Sparkles />} Buat survei
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(o) => !o && setDeleteTarget(null)}
        title="Hapus survei ini?"
        description={
          deleteTarget
            ? `"${deleteTarget.title}" dan ${deleteTarget.response_count} responsnya akan dihapus permanen.`
            : undefined
        }
        confirmText="Hapus"
        destructive
        onConfirm={() => {
          if (deleteTarget) deleteForm.mutate(deleteTarget.id);
          setDeleteTarget(null);
        }}
      />
    </main>
  );
}
