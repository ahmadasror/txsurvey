import { cn } from "@/lib/utils";
import { THEME_PRESETS } from "@/lib/themes";

/** ThemePicker shows the five Soft Studio themes as selectable swatch cards.
 *  onPreview fires on hover/focus (and null on leave) so callers can show a
 *  live preview of the hovered theme. */
export function ThemePicker({
  value,
  onChange,
  onPreview,
}: {
  value?: string;
  onChange: (id: string) => void;
  onPreview?: (id: string | null) => void;
}) {
  return (
    <div className="grid grid-cols-5 gap-2">
      {THEME_PRESETS.map((p) => {
        const selected = value === p.id;
        return (
          <button
            key={p.id}
            type="button"
            onClick={() => onChange(p.id)}
            onMouseEnter={() => onPreview?.(p.id)}
            onMouseLeave={() => onPreview?.(null)}
            onFocus={() => onPreview?.(p.id)}
            onBlur={() => onPreview?.(null)}
            aria-pressed={selected}
            className={cn(
              "flex flex-col items-center gap-2 rounded-xl border-2 p-2.5 text-xs transition-all",
              selected ? "border-primary" : "border-border hover:border-primary/40",
            )}
          >
            <span
              className="flex size-9 items-center justify-center rounded-full"
              style={{ background: p.swatch }}
            >
              <span className="size-3 rounded-full" style={{ background: p.accentSwatch }} />
            </span>
            <span className="font-medium text-foreground">{p.label}</span>
          </button>
        );
      })}
    </div>
  );
}
