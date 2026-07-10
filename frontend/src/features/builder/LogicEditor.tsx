import { useEffect, useState } from "react";
import { CornerDownRight, GitBranch, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { SimpleSelect } from "@/components/ui/select";
import { useAddLogicRule, useDeleteLogicRule, useUpdateLogicRule } from "@/api/forms";
import type { LogicAction, LogicOperator, LogicRule, LogicRuleInput, Question } from "@/types/forms";

const OPERATORS: { value: LogicOperator; label: string }[] = [
  { value: "eq", label: "is" },
  { value: "neq", label: "is not" },
  { value: "gt", label: ">" },
  { value: "gte", label: "≥" },
  { value: "lt", label: "<" },
  { value: "lte", label: "≤" },
  { value: "contains", label: "contains" },
  { value: "not_contains", label: "doesn’t contain" },
  { value: "is_empty", label: "is empty" },
  { value: "is_not_empty", label: "is not empty" },
];

const ACTIONS: { value: LogicAction; label: string }[] = [
  { value: "jump_to", label: "Lompat ke" },
  { value: "show", label: "Tampilkan" },
  { value: "hide", label: "Sembunyikan" },
  { value: "end_form", label: "Akhiri form" },
];

const needsValue = (op: LogicOperator) =>
  op !== "is_empty" && op !== "is_not_empty" && op !== "always";

function defaultValue(source: Question): unknown {
  if (source.type === "multiple_choice" || source.type === "checkboxes" || source.type === "dropdown")
    return source.metadata.options?.[0]?.id ?? "";
  if (source.type === "yes_no") return true;
  if (source.type === "rating" || source.type === "number") return 0;
  return "";
}

interface Props {
  formId: string;
  source: Question;
  questions: Question[];
}

/** RulesForQuestion renders + edits the logic rules whose source is `source`. */
export function RulesForQuestion({
  formId,
  source,
  questions,
  rules,
}: Props & { rules: LogicRule[] }) {
  const add = useAddLogicRule(formId);
  const rulesForSource = rules.filter((r) => r.source_question_id === source.id);
  const others = questions.filter((q) => q.id !== source.id);
  const forward = others.filter((q) => q.position > source.position);

  const onAdd = () => {
    const target = (forward[0] ?? others[0])?.id;
    const input: LogicRuleInput = {
      source_question_id: source.id,
      operator: "eq",
      compare_value: defaultValue(source),
      action: target ? "jump_to" : "end_form",
      target_question_id: target ?? null,
      priority: rulesForSource.length,
    };
    add.mutate(input);
  };

  // Unconditional jump: route straight to a later question, no condition to
  // reason about. Only makes sense when a later question exists to jump to.
  const onAddDirect = () => {
    const target = forward[0]?.id;
    if (!target) return;
    add.mutate({
      source_question_id: source.id,
      operator: "always",
      compare_value: null,
      action: "jump_to",
      target_question_id: target,
      priority: rulesForSource.length,
    });
  };

  return (
    <div className="space-y-3 rounded-md border p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2 text-sm font-semibold">
          <GitBranch className="size-4" /> Logika
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={onAddDirect}
            disabled={add.isPending || forward.length === 0}
            title={forward.length === 0 ? "Tidak ada pertanyaan setelah ini" : undefined}
          >
            <CornerDownRight /> Lompat langsung
          </Button>
          <Button variant="outline" size="sm" onClick={onAdd} disabled={add.isPending}>
            <Plus /> Tambah aturan
          </Button>
        </div>
      </div>
      {rulesForSource.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          Belum ada aturan. “Lompat langsung” meneruskan ke pertanyaan tertentu tanpa syarat;
          “Tambah aturan” bercabang, menampilkan/menyembunyikan, atau mengakhiri form berdasarkan jawaban.
        </p>
      ) : (
        <div className="space-y-2">
          {rulesForSource.map((rule) => (
            <RuleRow key={rule.id} formId={formId} rule={rule} source={source} questions={questions} />
          ))}
        </div>
      )}
      {add.isError && <p className="text-sm text-destructive">{(add.error as Error).message}</p>}
    </div>
  );
}

function RuleRow({
  formId,
  rule,
  source,
  questions,
}: {
  formId: string;
  rule: LogicRule;
  source: Question;
  questions: Question[];
}) {
  const update = useUpdateLogicRule(formId);
  const del = useDeleteLogicRule(formId);

  const toInput = (r: LogicRule): LogicRuleInput => ({
    source_question_id: source.id,
    operator: r.operator,
    compare_value: r.compare_value,
    action: r.action,
    target_question_id: r.target_question_id ?? null,
    priority: r.priority,
  });

  const [draft, setDraft] = useState<LogicRuleInput>(toInput(rule));
  useEffect(() => setDraft(toInput(rule)), [rule.id]); // eslint-disable-line react-hooks/exhaustive-deps

  const others = questions.filter((q) => q.id !== source.id);
  const targets = draft.action === "jump_to" ? others.filter((q) => q.position > source.position) : others;
  const dirty = JSON.stringify(draft) !== JSON.stringify(toInput(rule));

  // Unconditional jump: no operator/value/action to show — just the destination.
  if (rule.operator === "always") {
    return (
      <div className="rounded-md border bg-background p-2">
        <div className="flex flex-wrap items-center gap-2 text-sm">
          <span className="text-muted-foreground">Selalu lompat ke</span>
          <SimpleSelect
            className="h-9 w-44"
            value={draft.target_question_id ?? ""}
            onValueChange={(v) => setDraft((d) => ({ ...d, target_question_id: v }))}
            placeholder="Pilih pertanyaan…"
            options={targets.map((q) => ({ value: q.id, label: q.title || "Tanpa judul" }))}
          />
          <div className="ml-auto flex items-center gap-1">
            <Button size="sm" disabled={!dirty || update.isPending} onClick={() => update.mutate({ rid: rule.id, input: draft })}>
              Simpan
            </Button>
            <Button variant="ghost" size="icon" className="text-destructive" onClick={() => del.mutate(rule.id)} aria-label="Hapus aturan">
              <Trash2 />
            </Button>
          </div>
        </div>
        {update.isError && <p className="mt-1 text-xs text-destructive">{(update.error as Error).message}</p>}
      </div>
    );
  }

  const setValueControl = () => {
    if (!needsValue(draft.operator)) return null;
    const set = (v: unknown) => setDraft((d) => ({ ...d, compare_value: v }));
    if (source.type === "multiple_choice" || source.type === "checkboxes" || source.type === "dropdown") {
      return (
        <SimpleSelect
          className="h-9 w-40"
          value={String(draft.compare_value ?? "")}
          onValueChange={(v) => set(v)}
          placeholder="Value"
          options={(source.metadata.options ?? []).map((o) => ({ value: o.id, label: o.label }))}
        />
      );
    }
    if (source.type === "yes_no") {
      return (
        <SimpleSelect
          className="h-9 w-24"
          value={draft.compare_value === true ? "true" : "false"}
          onValueChange={(v) => set(v === "true")}
          options={[
            { value: "true", label: "Yes" },
            { value: "false", label: "No" },
          ]}
        />
      );
    }
    if (source.type === "rating" || source.type === "number") {
      return (
        <Input
          type="number"
          className="h-9 w-24"
          value={String(draft.compare_value ?? "")}
          onChange={(e) => set(e.target.value === "" ? "" : Number(e.target.value))}
        />
      );
    }
    return (
      <Input
        className="h-9 w-40"
        value={String(draft.compare_value ?? "")}
        onChange={(e) => set(e.target.value)}
      />
    );
  };

  return (
    <div className="rounded-md border bg-background p-2">
      <div className="flex flex-wrap items-center gap-2 text-sm">
        <span className="text-muted-foreground">Jika jawaban</span>
        <SimpleSelect
          className="h-9 w-36"
          value={draft.operator}
          onValueChange={(v) => setDraft((d) => ({ ...d, operator: v as LogicOperator }))}
          options={OPERATORS.map((o) => ({ value: o.value, label: o.label }))}
        />
        {setValueControl()}
        <span className="text-muted-foreground">maka</span>
        <SimpleSelect
          className="h-9 w-32"
          value={draft.action}
          onValueChange={(v) => {
            const action = v as LogicAction;
            setDraft((d) => ({
              ...d,
              action,
              target_question_id: action === "end_form" ? null : (d.target_question_id ?? targets[0]?.id ?? null),
            }));
          }}
          options={ACTIONS.map((a) => ({ value: a.value, label: a.label }))}
        />
        {draft.action !== "end_form" && (
          <SimpleSelect
            className="h-9 w-44"
            value={draft.target_question_id ?? ""}
            onValueChange={(v) => setDraft((d) => ({ ...d, target_question_id: v }))}
            placeholder="Pilih pertanyaan…"
            options={targets.map((q) => ({ value: q.id, label: q.title || "Tanpa judul" }))}
          />
        )}
        <div className="ml-auto flex items-center gap-1">
          <Button size="sm" disabled={!dirty || update.isPending} onClick={() => update.mutate({ rid: rule.id, input: draft })}>
            Simpan
          </Button>
          <Button variant="ghost" size="icon" className="text-destructive" onClick={() => del.mutate(rule.id)} aria-label="Hapus aturan">
            <Trash2 />
          </Button>
        </div>
      </div>
      {update.isError && <p className="mt-1 text-xs text-destructive">{(update.error as Error).message}</p>}
    </div>
  );
}
