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
}

const letter = (i: number) => String.fromCharCode(65 + i);

export function QuestionScreen({ question: q, value, onChange, onAdvance }: Props) {
  const meta = q.metadata ?? {};

  return (
    <div className="w-full">
      <h1 className="text-2xl font-semibold leading-tight sm:text-3xl">
        {q.title || "Untitled question"}
        {q.required && <span className="ml-1 text-primary">*</span>}
      </h1>
      {q.description && <p className="mt-2 text-base text-muted-foreground">{q.description}</p>}

      <div className="mt-6">
        {q.type === "short_text" && (
          <Input
            autoFocus
            className="h-12 text-lg"
            placeholder={meta.placeholder || "Type your answer…"}
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
          />
        )}

        {q.type === "long_text" && (
          <Textarea
            autoFocus
            className="min-h-32 text-lg"
            placeholder={meta.placeholder || "Type your answer…"}
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
          />
        )}

        {q.type === "email" && (
          <Input
            autoFocus
            type="email"
            className="h-12 text-lg"
            placeholder={meta.placeholder || "name@example.com"}
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
          />
        )}

        {q.type === "number" && (
          <Input
            autoFocus
            type="number"
            className="h-12 text-lg"
            placeholder={meta.placeholder || "Type a number…"}
            value={value === undefined ? "" : String(value)}
            onChange={(e) => onChange(e.target.value === "" ? "" : Number(e.target.value))}
          />
        )}

        {q.type === "date" && (
          <Input
            autoFocus
            type="date"
            className="h-12 w-auto text-lg"
            value={(value as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
          />
        )}

        {q.type === "yes_no" && (
          <div className="flex flex-col gap-3 sm:flex-row">
            {[
              { label: "Yes", v: true },
              { label: "No", v: false },
            ].map((opt) => (
              <ChoiceButton
                key={opt.label}
                selected={value === opt.v}
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
              className="h-12 text-lg"
              value={(value as string) ?? ""}
              onValueChange={(v) => onChange(v)}
              placeholder="Select…"
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
                  icon={checked ? <Check className="size-4" /> : undefined}
                  onClick={() =>
                    onChange(checked ? arr.filter((x) => x !== o.id) : [...arr, o.id])
                  }
                >
                  {o.label}
                </ChoiceButton>
              );
            })}
          </div>
        )}

        {q.type === "rating" && (
          <div className="flex flex-wrap gap-2">
            {Array.from({ length: meta.scale ?? 5 }, (_, i) => i + 1).map((n) => (
              <button
                key={n}
                onClick={() => {
                  onChange(n);
                  onAdvance();
                }}
                className={cn(
                  "flex size-12 items-center justify-center rounded-md border text-lg font-medium transition-colors",
                  value === n
                    ? "border-primary bg-primary text-primary-foreground"
                    : "hover:border-primary hover:bg-accent",
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
  icon,
  onClick,
}: {
  children: React.ReactNode;
  selected: boolean;
  badge?: string;
  icon?: React.ReactNode;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex w-full items-center gap-3 rounded-lg border px-4 py-3 text-left text-base transition-colors",
        selected
          ? "border-primary bg-primary/10 ring-1 ring-primary"
          : "hover:border-primary hover:bg-accent",
      )}
    >
      {badge && (
        <span
          className={cn(
            "flex size-6 shrink-0 items-center justify-center rounded border text-xs font-semibold",
            selected ? "border-primary text-primary" : "text-muted-foreground",
          )}
        >
          {badge}
        </span>
      )}
      <span className="flex-1">{children}</span>
      {icon}
    </button>
  );
}
