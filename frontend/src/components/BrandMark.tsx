import { cn } from "@/lib/utils";

/** BrandMark is the txsurvey logo mark: a primary rounded square with an inset
 *  rotated square whose top edge carries the theme accent. Pure CSS, no raster
 *  asset. Size is the outer square edge in px. */
export function BrandMark({ size = 52, className }: { size?: number; className?: string }) {
  const inner = Math.round(size * 0.42);
  return (
    <span
      className={cn("inline-grid shrink-0 place-items-center rounded-[28%] bg-primary", className)}
      style={{ width: size, height: size }}
      aria-hidden
    >
      <span
        className="rotate-45 rounded-[3px] border-t-2 border-brand bg-primary-foreground/20"
        style={{ width: inner, height: inner }}
      />
    </span>
  );
}
