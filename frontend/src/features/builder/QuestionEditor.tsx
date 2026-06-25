import { useEffect, useState } from "react";
import { Plus, Trash2, X, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { SimpleSelect } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { QUESTION_TYPES, typeDef } from "@/lib/questionTypes";
import { RulesForQuestion } from "@/features/builder/LogicEditor";
import { useDeleteQuestion, useUpdateQuestion } from "@/api/forms";
import type { LogicRule, Option, Question, QuestionInput, QuestionType } from "@/types/forms";

interface Props {
  formId: string;
  question: Question;
  questions: Question[];
  rules: LogicRule[];
  onDeleted: () => void;
}

function toInput(q: Question): QuestionInput {
  return {
    type: q.type,
    title: q.title,
    description: q.description,
    required: q.required,
    metadata: q.metadata ?? {},
  };
}

export function QuestionEditor({ formId, question, questions, rules, onDeleted }: Props) {
  const [draft, setDraft] = useState<QuestionInput>(toInput(question));
  const update = useUpdateQuestion(formId);
  const del = useDeleteQuestion(formId);

  // Reload the draft whenever a different question is selected.
  useEffect(() => setDraft(toInput(question)), [question.id]); // eslint-disable-line react-hooks/exhaustive-deps

  const def = typeDef(draft.type);
  const meta = draft.metadata ?? {};
  const setMeta = (patch: Partial<typeof meta>) => setDraft((d) => ({ ...d, metadata: { ...d.metadata, ...patch } }));

  const onTypeChange = (type: QuestionType) => {
    // Switching type resets metadata to that type's defaults.
    setDraft((d) => ({ ...d, type, metadata: { ...typeDef(type).defaultMetadata } }));
  };

  const setOptions = (options: Option[]) => setMeta({ options });
  const dirty = JSON.stringify(draft) !== JSON.stringify(toInput(question));

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">Edit question</h2>
        <Button
          variant="ghost"
          size="sm"
          className="text-destructive"
          onClick={() => {
            if (confirm("Delete this question?")) del.mutate(question.id, { onSuccess: onDeleted });
          }}
        >
          <Trash2 /> Delete
        </Button>
      </div>

      <div className="space-y-2">
        <Label>Type</Label>
        <SimpleSelect
          value={draft.type}
          onValueChange={(v) => onTypeChange(v as QuestionType)}
          options={QUESTION_TYPES.map((t) => ({ value: t.type, label: t.label }))}
        />
      </div>

      <div className="space-y-2">
        <Label>{def.isStatement ? "Statement text" : "Question"}</Label>
        <Input
          value={draft.title}
          placeholder={def.isStatement ? "Tell respondents something…" : "Ask a question…"}
          onChange={(e) => setDraft((d) => ({ ...d, title: e.target.value }))}
        />
      </div>

      <div className="space-y-2">
        <Label>Description (optional)</Label>
        <Textarea
          value={draft.description ?? ""}
          onChange={(e) => setDraft((d) => ({ ...d, description: e.target.value }))}
        />
      </div>

      {def.isChoice && (
        <OptionsEditor options={meta.options ?? []} onChange={setOptions} />
      )}

      {draft.type === "rating" && (
        <div className="space-y-2">
          <Label>Scale</Label>
          <SimpleSelect
            value={String(meta.scale ?? 5)}
            onValueChange={(v) => setMeta({ scale: Number(v) })}
            options={[3, 4, 5, 6, 7, 8, 9, 10].map((n) => ({ value: String(n), label: `1 – ${n}` }))}
          />
        </div>
      )}

      {draft.type === "number" && (
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>Min</Label>
            <Input
              type="number"
              value={meta.min ?? ""}
              onChange={(e) => setMeta({ min: e.target.value === "" ? undefined : Number(e.target.value) })}
            />
          </div>
          <div className="space-y-2">
            <Label>Max</Label>
            <Input
              type="number"
              value={meta.max ?? ""}
              onChange={(e) => setMeta({ max: e.target.value === "" ? undefined : Number(e.target.value) })}
            />
          </div>
        </div>
      )}

      {(draft.type === "short_text" || draft.type === "long_text" || draft.type === "number") && (
        <div className="space-y-2">
          <Label>Placeholder (optional)</Label>
          <Input value={meta.placeholder ?? ""} onChange={(e) => setMeta({ placeholder: e.target.value })} />
        </div>
      )}

      {!def.isStatement && (
        <div className="flex items-center justify-between rounded-md border p-3">
          <Label htmlFor="required">Required</Label>
          <Switch
            id="required"
            checked={!!draft.required}
            onCheckedChange={(v) => setDraft((d) => ({ ...d, required: v }))}
          />
        </div>
      )}

      <div className="flex justify-end">
        <Button
          disabled={!dirty || update.isPending}
          onClick={() => update.mutate({ qid: question.id, input: draft })}
        >
          {update.isPending ? <Loader2 className="animate-spin" /> : null} Save question
        </Button>
      </div>
      {update.isError && (
        <p className="text-sm text-destructive">{(update.error as Error).message}</p>
      )}

      {questions.length > 1 && (
        <RulesForQuestion formId={formId} source={question} questions={questions} rules={rules} />
      )}
    </div>
  );
}

function OptionsEditor({ options, onChange }: { options: Option[]; onChange: (o: Option[]) => void }) {
  const set = (i: number, label: string) =>
    onChange(options.map((o, idx) => (idx === i ? { ...o, label } : o)));
  const add = () => onChange([...options, { id: "", label: `Option ${options.length + 1}` }]);
  const remove = (i: number) => onChange(options.filter((_, idx) => idx !== i));

  return (
    <div className="space-y-2">
      <Label>Options</Label>
      <div className="space-y-2">
        {options.map((o, i) => (
          <div key={i} className="flex items-center gap-2">
            <Input value={o.label} onChange={(e) => set(i, e.target.value)} />
            <Button
              variant="ghost"
              size="icon"
              onClick={() => remove(i)}
              disabled={options.length <= 1}
              aria-label="Remove option"
            >
              <X />
            </Button>
          </div>
        ))}
      </div>
      <Button variant="outline" size="sm" onClick={add}>
        <Plus /> Add option
      </Button>
    </div>
  );
}
