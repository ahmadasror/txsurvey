import { Check } from "lucide-react";
import { cn } from "@/lib/utils";
import { THEME_PRESETS } from "@/lib/themes";

/** ThemePicker shows the five preset themes as selectable swatch cards.
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
    <div className="grid grid-cols-3 gap-2 sm:grid-cols-5">
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
              "relative flex flex-col items-center gap-1.5 rounded-lg border p-3 text-xs transition-colors",
              selected ? "border-primary ring-2 ring-primary/40" : "hover:bg-accent",
            )}
          >
            {selected && (
              <span className="absolute right-1 top-1 flex size-4 items-center justify-center rounded-full bg-primary text-primary-foreground">
                <Check className="size-3" />
              </span>
            )}
            <span className="size-8 rounded-full shadow-inner" style={{ background: p.swatch }} />
            <span className="text-base leading-none">{p.emoji}</span>
            <span className="font-medium">{p.label}</span>
          </button>
        );
      })}
    </div>
  );
}
