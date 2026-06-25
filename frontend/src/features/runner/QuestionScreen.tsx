import { Check } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { SimpleSelect } from "@/components/ui/select";
import { cn } from "@/lib/utils";
import type { AnswerValue, Question } from "@/types/forms";

interface Props {
  question: Question;
  value: AnswerValue | undefined;
  onChange: (v: AnswerValue) => void;
  onAdvance: () => void;
  step: number;
  total: number;
}

const letter = (i: number) => String.fromCharCode(65 + i);
const pad = (n: number) => String(n).padStart(2, "0");

const fieldClass =
  "rounded-xl border-input bg-card text-lg focus-visible:ring-4 focus-visible:ring-primary/15 focus-visible:border-primary";

export function QuestionScreen({ question: q, value, onChange, onAdvance, step, total }: Props) {
  const meta = q.metadata ?? {};

  return (
    <div className="w-full">
      <div className="label-eyebrow flex items-center gap-1.5 text-primary">
        <span className="tabular-nums">{pad(step)}</span>
        <span aria-hidden className="opacity-40">→</span>
        <span className="text-muted-foreground">dari {total}</span>
      </div>

      <h1 className="font-display mt-3 text-[27px] leading-[1.2] text-foreground sm:text-[32px]">
        {q.title || "Pertanyaan tanpa judul"}
        {q.required && <span className="ml-1 text-brand">*</span>}
      </h1>
      {q.description && <p className="text-body mt-2.5 text-[15px] sm:text-base">{q.description}</p>}

      <div className="mt-7">
        {q.type === "short_text" && (
          <Input
            autoFocus
            className={cn("h-14", fieldClass)}
            placeholder={meta.placeholder || "Tulis jawabanmu…"}
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
          />
        )}

        {q.type === "long_text" && (
          <Textarea
            autoFocus
            className={cn("min-h-36 py-3", fieldClass)}
            placeholder={meta.placeholder || "Tulis jawabanmu…"}
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
          />
        )}

        {q.type === "email" && (
          <Input
            autoFocus
            type="email"
            className={cn("h-14", fieldClass)}
            placeholder={meta.placeholder || "nama@contoh.com"}
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
          />
        )}

        {q.type === "number" && (
          <Input
            autoFocus
            type="number"
            className={cn("h-14", fieldClass)}
            placeholder={meta.placeholder || "Tulis angka…"}
            value={value === undefined ? "" : String(value)}
            onChange={(e) => onChange(e.target.value === "" ? "" : Number(e.target.value))}
          />
        )}

        {q.type === "date" && (
          <Input
            autoFocus
            type="date"
            className={cn("h-14 w-auto", fieldClass)}
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
          />
        )}

        {q.type === "yes_no" && (
          <div className="flex flex-col gap-3 sm:flex-row">
            {[
              { label: "Ya", v: true },
              { label: "Tidak", v: false },
            ].map((opt, i) => (
              <ChoiceButton
                key={opt.label}
                selected={value === opt.v}
                badge={letter(i)}
                className="sm:flex-1"
                onClick={() => {
                  onChange(opt.v);
                  onAdvance();
                }}
              >
                {opt.label}
              </ChoiceButton>
            ))}
          </div>
        )}

        {(q.type === "multiple_choice" || q.type === "dropdown") &&
          (q.type === "dropdown" ? (
            <SimpleSelect
              className={cn("h-14", fieldClass)}
              value={(value as string) ?? ""}
              onValueChange={(v) => onChange(v)}
              placeholder="Pilih…"
              options={(meta.options ?? []).map((o) => ({ value: o.id, label: o.label }))}
            />
          ) : (
            <div className="space-y-2.5">
              {(meta.options ?? []).map((o, i) => (
                <ChoiceButton
                  key={o.id}
                  selected={value === o.id}
                  badge={letter(i)}
                  onClick={() => {
                    onChange(o.id);
                    onAdvance();
                  }}
                >
                  {o.label}
                </ChoiceButton>
              ))}
            </div>
          ))}

        {q.type === "checkboxes" && (
          <div className="space-y-2.5">
            {(meta.options ?? []).map((o, i) => {
              const arr = Array.isArray(value) ? (value as string[]) : [];
              const checked = arr.includes(o.id);
              return (
                <ChoiceButton
                  key={o.id}
                  selected={checked}
                  badge={letter(i)}
                  onClick={() => onChange(checked ? arr.filter((x) => x !== o.id) : [...arr, o.id])}
                >
                  {o.label}
                </ChoiceButton>
              );
            })}
          </div>
        )}

        {q.type === "rating" && (
          <div className="flex flex-wrap gap-2.5">
            {Array.from({ length: meta.scale ?? 5 }, (_, i) => i + 1).map((n) => (
              <button
                key={n}
                onClick={() => {
                  onChange(n);
                  onAdvance();
                }}
                className={cn(
                  "font-display grid size-12 place-items-center rounded-xl border text-lg transition-all",
                  value === n
                    ? "border-primary bg-primary text-primary-foreground ring-4 ring-primary/15"
                    : "border-input bg-card hover:-translate-y-0.5 hover:border-primary/50",
                )}
              >
                {n}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function ChoiceButton({
  children,
  selected,
  badge,
  className,
  onClick,
}: {
  children: React.ReactNode;
  selected: boolean;
  badge?: string;
  className?: string;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex w-full items-center gap-3 rounded-2xl border px-4 py-3.5 text-left text-base transition-all",
        selected
          ? "border-primary bg-primary-soft ring-4 ring-primary/15"
          : "border-input bg-card hover:-translate-y-0.5 hover:border-primary/50",
        className,
      )}
    >
      {badge && (
        <span
          className={cn(
            "grid size-7 shrink-0 place-items-center rounded-lg border text-xs font-semibold transition-colors",
            selected ? "border-primary bg-primary text-primary-foreground" : "border-input text-muted-foreground",
          )}
        >
          {badge}
        </span>
      )}
      <span className="flex-1 text-foreground">{children}</span>
      {selected && <Check className="size-4 shrink-0 text-primary" />}
    </button>
  );
}
