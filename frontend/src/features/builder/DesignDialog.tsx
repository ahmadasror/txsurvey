import { useEffect, useRef, useState, type CSSProperties } from "react";
import { ArrowRight, Check, ImagePlus, Loader2 } from "lucide-react";
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
import { BrandMark } from "@/components/BrandMark";
import { cn } from "@/lib/utils";
import { assetUrl } from "@/lib/paths";
import { DEFAULT_FONT_ID, DEFAULT_THEME_ID, FONT_PRESETS, themeStyle } from "@/lib/themes";
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
  const previewTheme = hoverTheme ?? s.theme?.preset ?? DEFAULT_THEME_ID;
  const fontId = s.font ?? DEFAULT_FONT_ID;

  const save = () =>
    update.mutate(
      { title: form.title, description: form.description, settings: s },
      { onSuccess: () => onOpenChange(false) },
    );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] max-w-3xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="font-display text-2xl">Desain</DialogTitle>
          <DialogDescription>Atur tampilan survei untuk responden.</DialogDescription>
        </DialogHeader>

        <div className="grid gap-6 md:grid-cols-[300px_1fr]">
          {/* Live welcome preview */}
          <div className="md:sticky md:top-0 md:self-start">
            <Label className="label-eyebrow mb-2 block text-muted-foreground">Pratinjau</Label>
            <WelcomePreview
              presetId={previewTheme}
              fontId={fontId}
              banner={assetUrl(s.banner_url)}
              logo={assetUrl(s.logo_url)}
              title={s.welcome_title || form.title}
              desc={s.welcome_description}
            />
          </div>

          {/* Controls */}
          <div className="space-y-6">
            <section className="space-y-3">
              <Label className="label-eyebrow text-muted-foreground">Tema</Label>
              <ThemePicker
                value={s.theme?.preset ?? DEFAULT_THEME_ID}
                onChange={(id) => set({ theme: { ...s.theme, preset: id } })}
                onPreview={setHoverTheme}
              />
            </section>

            <section className="space-y-3">
              <Label className="label-eyebrow text-muted-foreground">Font</Label>
              <div className="grid grid-cols-2 gap-2">
                {FONT_PRESETS.map((f) => {
                  const selected = fontId === f.id;
                  return (
                    <button
                      key={f.id}
                      type="button"
                      onClick={() => set({ font: f.id })}
                      className={cn(
                        "flex items-center gap-3 rounded-xl border-2 px-3 py-2.5 text-left transition-colors",
                        selected ? "border-primary bg-primary-soft" : "border-border hover:border-primary/40",
                      )}
                    >
                      <span className="text-2xl leading-none text-foreground" style={{ fontFamily: f.display }}>
                        {f.sample}
                      </span>
                      <span className="text-sm font-medium text-foreground">{f.label}</span>
                    </button>
                  );
                })}
              </div>
            </section>

            <section className="grid gap-4 sm:grid-cols-2">
              <ImageUpload
                label="Banner (sampul)"
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

            <section className="space-y-3">
              <Label className="label-eyebrow text-muted-foreground">Layar sambutan</Label>
              <Input
                placeholder="Judul sambutan"
                value={s.welcome_title ?? ""}
                onChange={(e) => set({ welcome_title: e.target.value })}
              />
              <Textarea
                placeholder="Deskripsi sambutan"
                value={s.welcome_description ?? ""}
                onChange={(e) => set({ welcome_description: e.target.value })}
              />
              <Input
                placeholder="Teks tombol mulai (default: Mulai)"
                value={s.start_button_text ?? ""}
                onChange={(e) => set({ start_button_text: e.target.value })}
              />
            </section>

            <section className="space-y-3">
              <Label className="label-eyebrow text-muted-foreground">Layar terima kasih</Label>
              <Input
                placeholder="Judul terima kasih"
                value={s.thank_you_title ?? ""}
                onChange={(e) => set({ thank_you_title: e.target.value })}
              />
              <Textarea
                placeholder="Pesan terima kasih"
                value={s.thank_you_description ?? ""}
                onChange={(e) => set({ thank_you_description: e.target.value })}
              />
            </section>

            <div className="flex items-center justify-between rounded-xl border p-3">
              <Label htmlFor="progress">Tampilkan progress bar</Label>
              <Switch id="progress" checked={!!s.show_progress} onCheckedChange={(v) => set({ show_progress: v })} />
            </div>
          </div>
        </div>

        {update.isError && <p className="text-sm text-destructive">{(update.error as Error).message}</p>}

        <DialogFooter>
          <Button className="rounded-xl" onClick={save} disabled={update.isPending}>
            {update.isPending ? <Loader2 className="animate-spin" /> : <Check />} Simpan desain
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function WelcomePreview({
  presetId,
  fontId,
  banner,
  logo,
  title,
  desc,
}: {
  presetId: string;
  fontId: string;
  banner?: string;
  logo?: string;
  title: string;
  desc?: string;
}) {
  const style = themeStyle({ preset: presetId }, fontId) as CSSProperties | undefined;
  return (
    <div style={style} className="font-sans overflow-hidden rounded-2xl border bg-background transition-colors">
      <div
        className="h-16 bg-primary-soft bg-cover bg-center"
        style={banner ? { backgroundImage: `url(${banner})` } : undefined}
      />
      <div className="px-4 pb-5">
        <div className="-mt-7 mb-2 flex justify-center">
          {logo ? (
            <img
              src={logo}
              alt=""
              className="size-14 rounded-full border object-cover ring-[3px] ring-background"
            />
          ) : (
            <span className="rounded-full ring-[3px] ring-background">
              <BrandMark size={52} className="rounded-full" />
            </span>
          )}
        </div>
        <div className="text-center">
          <div className="label-eyebrow text-brand">Survei</div>
          <div className="font-display mt-1 text-lg leading-tight text-foreground">{title || "Judul survei"}</div>
          {desc && <p className="text-body mt-1.5 line-clamp-2 text-xs">{desc}</p>}
          <span className="mt-3 inline-flex items-center gap-1 rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground">
            Mulai <ArrowRight className="size-3" />
          </span>
        </div>
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
        <div className={cn("relative overflow-hidden rounded-xl border", className)}>
          <img src={assetUrl(value)} alt={label} className="size-full object-cover" />
          <Button
            type="button"
            variant="secondary"
            size="sm"
            className="absolute right-1.5 top-1.5 h-7"
            onClick={() => onChange("")}
          >
            Hapus
          </Button>
        </div>
      ) : (
        <button
          type="button"
          onClick={() => ref.current?.click()}
          disabled={upload.isPending}
          className={cn(
            "flex w-full flex-col items-center justify-center gap-1.5 rounded-xl border border-dashed text-sm text-muted-foreground hover:bg-accent",
            className ?? "py-6",
          )}
        >
          {upload.isPending ? <Loader2 className="size-5 animate-spin" /> : <ImagePlus className="size-5" />}
          {upload.isPending ? "Mengunggah…" : "Unggah gambar"}
        </button>
      )}
      <input
        ref={ref}
        type="file"
        accept="image/png,image/jpeg,image/webp"
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
