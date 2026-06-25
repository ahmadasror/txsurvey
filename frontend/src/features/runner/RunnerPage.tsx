import { useCallback, useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { AnimatePresence, motion } from "framer-motion";
import { ArrowLeft, ArrowRight, Check, CornerDownLeft, Loader2, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FullScreenLoader } from "@/components/FullScreenLoader";
import { QuestionScreen } from "@/features/runner/QuestionScreen";
import { usePublicForm, useSubmitResponse, type SubmitAnswer } from "@/api/public";
import { hexToHslTriple } from "@/lib/theme";
import { firstQuestionId, nextQuestionId, reachablePath } from "@/lib/logicEngine";
import type { AnswerValue, LogicRule, Question } from "@/types/forms";

type Answers = Record<string, AnswerValue | undefined>;

const isEmpty = (v: AnswerValue | undefined): boolean =>
  v === undefined || v === null || v === "" || (Array.isArray(v) && v.length === 0);

export function RunnerPage() {
  const { slug = "" } = useParams();
  const { data: form, isLoading, isError } = usePublicForm(slug);
  const submit = useSubmitResponse(slug);

  const [started, setStarted] = useState(false);
  const [history, setHistory] = useState<string[]>([]); // visited question ids; current = last
  const [done, setDone] = useState(false);
  const [answers, setAnswers] = useState<Answers>({});
  const [error, setError] = useState<string | null>(null);

  const questions: Question[] = useMemo(() => form?.questions ?? [], [form]);
  const rules: LogicRule[] = useMemo(() => form?.logic_rules ?? [], [form]);

  const currentId = history[history.length - 1];
  const current = questions.find((q) => q.id === currentId) ?? null;

  const setAnswer = (qid: string, v: AnswerValue) => {
    setAnswers((a) => ({ ...a, [qid]: v }));
    setError(null);
  };

  const start = () => {
    const first = firstQuestionId(questions, rules, answers);
    if (first) {
      setHistory([first]);
      setStarted(true);
    }
  };

  const doSubmit = useCallback(
    (finalAnswers: Answers) => {
      // Only submit answers to questions actually on the reachable path.
      const reach = new Set(reachablePath(questions, rules, finalAnswers));
      const payload: SubmitAnswer[] = questions
        .filter((q) => q.type !== "statement" && reach.has(q.id) && !isEmpty(finalAnswers[q.id]))
        .map((q) => ({ question_id: q.id, value: finalAnswers[q.id] as AnswerValue }));
      submit.mutate(payload, {
        onSuccess: () => {
          if (form?.settings.redirect_url) window.location.href = form.settings.redirect_url;
          else setDone(true);
        },
      });
    },
    [questions, rules, submit, form],
  );

  const next = useCallback(() => {
    if (!current) return;
    if (current.type !== "statement" && current.required && isEmpty(answers[current.id])) {
      setError("This question is required.");
      return;
    }
    setError(null);
    const nid = nextQuestionId(questions, rules, answers, current.id);
    if (nid === null) doSubmit(answers);
    else setHistory((h) => [...h, nid]);
  }, [current, answers, questions, rules, doSubmit]);

  const back = useCallback(() => {
    setError(null);
    setHistory((h) => (h.length > 1 ? h.slice(0, -1) : (setStarted(false), h)));
  }, []);

  // Keyboard: Enter advances (except in a textarea); digit keys pick options.
  useEffect(() => {
    if (!started || done || !current) return;
    const onKey = (e: KeyboardEvent) => {
      const el = document.activeElement;
      if (e.key === "Enter" && !(el instanceof HTMLTextAreaElement)) {
        e.preventDefault();
        next();
        return;
      }
      const inField = el instanceof HTMLInputElement || el instanceof HTMLSelectElement || el instanceof HTMLTextAreaElement;
      if (!inField && /^[1-9]$/.test(e.key)) {
        const i = Number(e.key) - 1;
        if (current.type === "multiple_choice") {
          const opt = current.metadata.options?.[i];
          if (opt) {
            setAnswer(current.id, opt.id);
            setTimeout(next, 120);
          }
        } else if (current.type === "checkboxes") {
          const opt = current.metadata.options?.[i];
          if (opt) {
            const arr = Array.isArray(answers[current.id]) ? (answers[current.id] as string[]) : [];
            setAnswer(current.id, arr.includes(opt.id) ? arr.filter((x) => x !== opt.id) : [...arr, opt.id]);
          }
        } else if (current.type === "rating") {
          if (i + 1 <= (current.metadata.scale ?? 5)) {
            setAnswer(current.id, i + 1);
            setTimeout(next, 120);
          }
        } else if (current.type === "yes_no" && i < 2) {
          setAnswer(current.id, i === 0);
          setTimeout(next, 120);
        }
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [started, done, current, next, answers]);

  const accent = hexToHslTriple(form?.settings.theme?.accent);
  const rootStyle = accent ? ({ "--primary": accent, "--ring": accent } as React.CSSProperties) : undefined;

  if (isLoading) return <FullScreenLoader />;
  if (isError || !form)
    return (
      <Centered>
        <p className="text-lg font-medium">This form isn’t available.</p>
        <p className="text-muted-foreground">It may be unpublished or the link is wrong.</p>
      </Centered>
    );

  const path = reachablePath(questions, rules, answers);
  const showProgress = form.settings.show_progress && started && !done && path.length > 0;
  const progress = path.length ? (Math.min(history.length, path.length) / path.length) * 100 : 0;
  const isLast = current ? nextQuestionId(questions, rules, answers, current.id) === null : false;

  return (
    <div style={rootStyle} className="flex min-h-dvh flex-col bg-background">
      {showProgress && (
        <div className="h-1.5 w-full bg-muted">
          <div className="h-full bg-primary transition-all duration-300" style={{ width: `${progress}%` }} />
        </div>
      )}

      <div className="flex flex-1 items-center justify-center p-6">
        <div className="w-full max-w-xl">
          {!started ? (
            <Welcome
              title={form.settings.welcome_title || form.title}
              description={form.settings.welcome_description || form.description}
              onStart={start}
              empty={questions.length === 0}
            />
          ) : done ? (
            <ThankYou
              title={form.settings.thank_you_title || "Thank you!"}
              description={form.settings.thank_you_description || "Your response has been recorded."}
            />
          ) : (
            <AnimatePresence mode="wait">
              <motion.div
                key={currentId}
                initial={{ opacity: 0, y: 24 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -24 }}
                transition={{ duration: 0.22, ease: "easeOut" }}
              >
                {current && (
                  <QuestionScreen
                    question={current}
                    value={answers[current.id]}
                    onChange={(v) => setAnswer(current.id, v)}
                    onAdvance={() => setTimeout(next, 120)}
                  />
                )}

                {error && <p className="mt-3 text-sm text-destructive">{error}</p>}
                {submit.isError && <p className="mt-3 text-sm text-destructive">{(submit.error as Error).message}</p>}

                <div className="mt-8 flex items-center gap-3">
                  <Button onClick={next} size="lg" disabled={submit.isPending}>
                    {submit.isPending ? <Loader2 className="animate-spin" /> : isLast ? <Send /> : <ArrowRight />}
                    {isLast ? "Submit" : "OK"}
                  </Button>
                  <span className="hidden items-center gap-1 text-xs text-muted-foreground sm:flex">
                    press Enter <CornerDownLeft className="size-3" />
                  </span>
                  <Button variant="ghost" size="sm" className="ml-auto" onClick={back}>
                    <ArrowLeft /> Back
                  </Button>
                </div>
              </motion.div>
            </AnimatePresence>
          )}
        </div>
      </div>
    </div>
  );
}

function Welcome({
  title,
  description,
  onStart,
  empty,
}: {
  title: string;
  description?: string;
  onStart: () => void;
  empty: boolean;
}) {
  return (
    <div className="text-center">
      <h1 className="text-3xl font-bold sm:text-4xl">{title}</h1>
      {description && <p className="mt-3 text-lg text-muted-foreground">{description}</p>}
      {empty ? (
        <p className="mt-8 text-muted-foreground">This form has no questions yet.</p>
      ) : (
        <Button size="lg" className="mt-8" onClick={onStart}>
          Start <ArrowRight />
        </Button>
      )}
    </div>
  );
}

function ThankYou({ title, description }: { title: string; description?: string }) {
  return (
    <div className="text-center">
      <div className="mx-auto mb-5 flex size-14 items-center justify-center rounded-full bg-primary/10">
        <Check className="size-7 text-primary" />
      </div>
      <h1 className="text-3xl font-bold">{title}</h1>
      {description && <p className="mt-3 text-lg text-muted-foreground">{description}</p>}
    </div>
  );
}

function Centered({ children }: { children: React.ReactNode }) {
  return <div className="flex min-h-dvh flex-col items-center justify-center gap-1 p-6 text-center">{children}</div>;
}
