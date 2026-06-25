import { useEffect, useRef, useState, type CSSProperties } from "react";
import { Check, ImagePlus, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ThemePicker } from "@/components/ThemePicker";
import { cn } from "@/lib/utils";
import { assetUrl } from "@/lib/paths";
import { presetById } from "@/lib/themes";
import { useUpdateForm, useUploadAsset } from "@/api/forms";
import type { Form, FormSettings } from "@/types/forms";

export function DesignDialog({
  form,
  open,
  onOpenChange,
}: {
  form: Form;
  open: boolean;
  onOpenChange: (o: boolean) => void;
}) {
  const update = useUpdateForm(form.id);
  const [s, setS] = useState<FormSettings>(form.settings);
  const [hoverTheme, setHoverTheme] = useState<string | null>(null);

  useEffect(() => {
    if (open) setS(form.settings);
  }, [open, form.settings]);

  const set = (patch: Partial<FormSettings>) => setS((prev) => ({ ...prev, ...patch }));
  const previewTheme = hoverTheme ?? s.theme?.preset ?? "corporate";

  const save = () =>
    update.mutate(
      { title: form.title, description: form.description, settings: s },
      { onSuccess: () => onOpenChange(false) },
    );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[88vh] max-w-2xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Design</DialogTitle>
          <DialogDescription>Customize how your survey looks to respondents.</DialogDescription>
        </DialogHeader>

        {/* Theme + live preview */}
        <section className="space-y-3">
          <Label>Theme</Label>
          <ThemePicker
            value={s.theme?.preset}
            onChange={(id) => set({ theme: { ...s.theme, preset: id } })}
            onPreview={setHoverTheme}
          />
          <ThemePreview presetId={previewTheme} />
        </section>

        {/* Branding */}
        <section className="grid gap-4 sm:grid-cols-2">
          <ImageUpload
            label="Banner (cover image)"
            value={s.banner_url}
            onChange={(u) => set({ banner_url: u })}
            formId={form.id}
            className="aspect-[16/6]"
          />
          <ImageUpload
            label="Logo"
            value={s.logo_url}
            onChange={(u) => set({ logo_url: u })}
            formId={form.id}
            className="aspect-square max-w-[8rem]"
          />
        </section>

        {/* Welcome screen */}
        <section className="space-y-3">
          <Label className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Welcome screen</Label>
          <Input placeholder="Welcome title" value={s.welcome_title ?? ""} onChange={(e) => set({ welcome_title: e.target.value })} />
          <Textarea
            placeholder="Welcome description"
            value={s.welcome_description ?? ""}
            onChange={(e) => set({ welcome_description: e.target.value })}
          />
          <Input
            placeholder="Start button text (default: Start)"
            value={s.start_button_text ?? ""}
            onChange={(e) => set({ start_button_text: e.target.value })}
          />
        </section>

        {/* Thank-you screen */}
        <section className="space-y-3">
          <Label className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Thank-you screen</Label>
          <Input placeholder="Thank-you title" value={s.thank_you_title ?? ""} onChange={(e) => set({ thank_you_title: e.target.value })} />
          <Textarea
            placeholder="Thank-you message"
            value={s.thank_you_description ?? ""}
            onChange={(e) => set({ thank_you_description: e.target.value })}
          />
        </section>

        <div className="flex items-center justify-between rounded-md border p-3">
          <Label htmlFor="progress">Show progress bar</Label>
          <Switch id="progress" checked={!!s.show_progress} onCheckedChange={(v) => set({ show_progress: v })} />
        </div>

        {update.isError && <p className="text-sm text-destructive">{(update.error as Error).message}</p>}

        <DialogFooter>
          <Button onClick={save} disabled={update.isPending}>
            {update.isPending ? <Loader2 className="animate-spin" /> : <Check />} Save design
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function ThemePreview({ presetId }: { presetId: string }) {
  const vars = presetById(presetId)?.vars as CSSProperties | undefined;
  return (
    <div style={vars} className="rounded-lg border bg-background p-4 transition-colors">
      <div className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Live preview</div>
      <div className="mt-1.5 text-sm font-semibold text-foreground">Seberapa puas kamu?</div>
      <div className="mt-2 space-y-1.5">
        <div className="rounded-md border border-primary bg-primary/10 px-3 py-1.5 text-xs text-foreground">Sangat puas</div>
        <div className="rounded-md border px-3 py-1.5 text-xs text-foreground">Biasa aja</div>
      </div>
      <div className="mt-3 inline-flex rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground">
        OK →
      </div>
    </div>
  );
}

function ImageUpload({
  label,
  value,
  onChange,
  formId,
  className,
}: {
  label: string;
  value?: string;
  onChange: (url: string) => void;
  formId: string;
  className?: string;
}) {
  const upload = useUploadAsset(formId);
  const ref = useRef<HTMLInputElement>(null);

  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      {value ? (
        <div className={cn("relative overflow-hidden rounded-md border", className)}>
          <img src={assetUrl(value)} alt={label} className="size-full object-cover" />
          <Button
            type="button"
            variant="secondary"
            size="sm"
            className="absolute right-1.5 top-1.5 h-7"
            onClick={() => onChange("")}
          >
            Remove
          </Button>
        </div>
      ) : (
        <button
          type="button"
          onClick={() => ref.current?.click()}
          disabled={upload.isPending}
          className={cn(
            "flex w-full flex-col items-center justify-center gap-1.5 rounded-md border border-dashed text-sm text-muted-foreground hover:bg-accent",
            className ?? "py-6",
          )}
        >
          {upload.isPending ? <Loader2 className="size-5 animate-spin" /> : <ImagePlus className="size-5" />}
          {upload.isPending ? "Uploading…" : "Upload image"}
        </button>
      )}
      <input
        ref={ref}
        type="file"
        accept="image/png,image/jpeg,image/webp,image/gif"
        className="hidden"
        onChange={(e) => {
          const f = e.target.files?.[0];
          if (f) upload.mutate(f, { onSuccess: (r) => onChange(r.url) });
          e.currentTarget.value = "";
        }}
      />
      {upload.isError && <p className="text-xs text-destructive">{(upload.error as Error).message}</p>}
    </div>
  );
}
