import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, BarChart3, Check, ChevronDown, Copy, Eye, Loader2, Palette, Plus, Send, Undo2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Card } from "@/components/ui/card";
import { FullScreenLoader } from "@/components/FullScreenLoader";
import { SortableQuestionList } from "@/features/builder/SortableQuestionList";
import { QuestionEditor } from "@/features/builder/QuestionEditor";
import { DesignDialog } from "@/features/builder/DesignDialog";
import { cn } from "@/lib/utils";
import { QUESTION_TYPES, typeDef } from "@/lib/questionTypes";
import { themeStyle } from "@/lib/themes";
import { runnerPath, runnerUrl } from "@/lib/paths";
import {
  useAddQuestion,
  useForm,
  usePublishForm,
  useReorderQuestions,
  useUpdateForm,
} from "@/api/forms";
import type { Question, QuestionType } from "@/types/forms";

export function BuilderPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const { data: form, isLoading, isError } = useForm(id);

  const updateForm = useUpdateForm(id);
  const publish = usePublishForm(id);
  const addQuestion = useAddQuestion(id);
  const reorder = useReorderQuestions(id);

  const [title, setTitle] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [themeOpen, setThemeOpen] = useState(false);
  // On mobile the list and editor are separate panes (toggled); desktop shows both.
  const [mobilePane, setMobilePane] = useState<"list" | "editor">("list");

  const selectQuestion = (qid: string | null) => {
    setSelectedId(qid);
    if (qid) setMobilePane("editor");
  };

  const questions: Question[] = useMemo(() => form?.questions ?? [], [form]);

  useEffect(() => {
    if (form) setTitle(form.title);
  }, [form?.title]); // eslint-disable-line react-hooks/exhaustive-deps

  // Keep a valid selection as the question set changes.
  useEffect(() => {
    if (questions.length === 0) {
      setSelectedId(null);
    } else if (!questions.some((q) => q.id === selectedId)) {
      setSelectedId(questions[0].id);
    }
  }, [questions, selectedId]);

  if (isLoading) return <FullScreenLoader />;
  if (isError || !form)
    return (
      <main className="container max-w-3xl py-20 text-center">
        <p className="text-muted-foreground">Form not found.</p>
        <Button variant="outline" className="mt-4" onClick={() => navigate("/")}>
          <ArrowLeft /> Back to forms
        </Button>
      </main>
    );

  const saveTitle = () => {
    if (title.trim() && title !== form.title) {
      updateForm.mutate({ title: title.trim(), description: form.description, settings: form.settings });
    }
  };

  const onAdd = (t: QuestionType) => {
    const def = typeDef(t);
    addQuestion.mutate(
      { type: t, title: "", required: false, metadata: { ...def.defaultMetadata } },
      { onSuccess: (q) => selectQuestion(q.id) },
    );
  };

  const publicUrl = runnerUrl(form.slug);
  const copyLink = () => {
    navigator.clipboard.writeText(publicUrl).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    });
  };

  const selected = questions.find((q) => q.id === selectedId) ?? null;
  const isPublished = form.status === "published";

  return (
    <div style={themeStyle(form.settings.theme, form.settings.font)} className="font-sans min-h-dvh bg-background">
      {/* Builder header */}
      <div className="border-b bg-card">
        <div className="container flex flex-wrap items-center gap-3 py-3">
          <Button variant="ghost" size="icon" onClick={() => navigate("/")} aria-label="Back">
            <ArrowLeft />
          </Button>
          <Input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onBlur={saveTitle}
            className="font-display h-9 max-w-xs border-transparent text-lg font-medium hover:border-input focus-visible:border-input"
          />
          <Badge variant={isPublished ? "success" : "muted"}>{form.status}</Badge>

          <div className="ml-auto flex items-center gap-2">
            {isPublished && (
              <Button variant="outline" size="sm" onClick={copyLink}>
                {copied ? <Check /> : <Copy />} {copied ? "Copied" : "Share link"}
              </Button>
            )}
            <Button variant="outline" size="sm" onClick={() => setThemeOpen(true)}>
              <Palette /> Design
            </Button>
            <Button variant="outline" size="sm" asChild>
              <Link to={`/forms/${id}/results`}>
                <BarChart3 /> Results
              </Link>
            </Button>
            <Button variant="outline" size="sm" onClick={() => window.open(runnerPath(form.slug), "_blank")}>
              <Eye /> Preview
            </Button>
            {isPublished ? (
              <Button variant="secondary" size="sm" onClick={() => publish.mutate(false)} disabled={publish.isPending}>
                <Undo2 /> Unpublish
              </Button>
            ) : (
              <Button size="sm" onClick={() => publish.mutate(true)} disabled={publish.isPending}>
                {publish.isPending ? <Loader2 className="animate-spin" /> : <Send />} Publish
              </Button>
            )}
          </div>
        </div>
        {publish.isError && (
          <div className="container pb-2 text-sm text-destructive">{(publish.error as Error).message}</div>
        )}
      </div>

      {/* Mobile pane toggle (desktop shows both panes side by side) */}
      <div className="container md:hidden">
        <div className="mt-4 grid grid-cols-2 gap-1 rounded-lg border bg-muted/40 p-1">
          {(["list", "editor"] as const).map((pane) => (
            <button
              key={pane}
              onClick={() => setMobilePane(pane)}
              className={cn(
                "rounded-md py-1.5 text-sm font-medium capitalize transition-colors",
                mobilePane === pane ? "bg-background shadow-sm" : "text-muted-foreground",
              )}
            >
              {pane === "list" ? `Questions${questions.length ? ` (${questions.length})` : ""}` : "Editor"}
            </button>
          ))}
        </div>
      </div>

      {/* Builder body */}
      <main className="container grid gap-[22px] py-6 md:grid-cols-[17.5rem_1fr]">
        <aside className={cn("space-y-3", mobilePane === "editor" && "hidden md:block")}>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="w-full justify-between rounded-xl bg-card" disabled={addQuestion.isPending}>
                <span className="flex items-center gap-2">
                  <Plus /> Tambah pertanyaan
                </span>
                <ChevronDown className="size-4 opacity-50" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-[var(--radix-dropdown-menu-trigger-width)]">
              {QUESTION_TYPES.map((t) => (
                <DropdownMenuItem key={t.type} onSelect={() => onAdd(t.type)}>
                  {t.label}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>

          {questions.length === 0 ? (
            <p className="rounded-xl border border-dashed bg-card/50 p-4 text-center text-sm text-muted-foreground">
              Tambahkan pertanyaan pertamamu.
            </p>
          ) : (
            <SortableQuestionList
              questions={questions}
              selectedId={selectedId}
              onSelect={selectQuestion}
              onReorder={(ids) => reorder.mutate(ids)}
            />
          )}
        </aside>

        <section className={cn(mobilePane === "list" && "hidden md:block")}>
          {selected ? (
            <Card className="rounded-2xl p-6 sm:p-[26px]">
              <QuestionEditor
                formId={id}
                question={selected}
                questions={questions}
                rules={form.logic_rules ?? []}
                onDeleted={() => {
                  setSelectedId(null);
                  setMobilePane("list");
                }}
              />
            </Card>
          ) : (
            <Card className="flex items-center justify-center rounded-2xl p-16 text-center text-sm text-muted-foreground">
              Pilih atau tambah pertanyaan untuk mengeditnya.
            </Card>
          )}
        </section>
      </main>

      <DesignDialog form={form} open={themeOpen} onOpenChange={setThemeOpen} />
    </div>
  );
}
