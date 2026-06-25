import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { AnimatePresence, motion } from "framer-motion";
import { ArrowLeft, ArrowRight, Check, CornerDownLeft, Loader2, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { BrandMark } from "@/components/BrandMark";
import { FullScreenLoader } from "@/components/FullScreenLoader";
import { QuestionScreen } from "@/features/runner/QuestionScreen";
import { usePublicForm, useSubmitResponse, type SubmitAnswer } from "@/api/public";
import { themeStyle } from "@/lib/themes";
import { assetUrl, homePath } from "@/lib/paths";
import { firstQuestionId, nextQuestionId, reachablePath } from "@/lib/logicEngine";
import type { AnswerValue, LogicRule, Question } from "@/types/forms";

type Answers = Record<string, AnswerValue | undefined>;

// Question-enter transition (Soft Studio): fade + 22px rise, ~290ms ease-out.
const ENTER = { duration: 0.29, ease: [0.2, 0.7, 0.3, 1] as const };

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
      setError("Pertanyaan ini wajib diisi.");
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

  // `next` closes over `answers`; a delayed auto-advance (after picking a choice/
  // rating) must call the LATEST `next` — once the just-picked answer has applied
  // — otherwise it sees stale state and falsely reports the question unanswered
  // ("wajib diisi" until you click a second time). Route every timer through a ref.
  const nextRef = useRef(next);
  nextRef.current = next;
  const scheduleAdvance = useCallback((delay = 170) => {
    window.setTimeout(() => nextRef.current(), delay);
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
            scheduleAdvance();
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
            scheduleAdvance();
          }
        } else if (current.type === "yes_no" && i < 2) {
          setAnswer(current.id, i === 0);
          scheduleAdvance();
        }
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [started, done, current, next, answers]);

  const rootStyle = themeStyle(form?.settings.theme, form?.settings.font);

  if (isLoading) return <FullScreenLoader />;
  if (isError || !form)
    return (
      <Centered>
        <p className="font-display text-2xl text-foreground">Survei ini tidak tersedia.</p>
        <p className="text-body">Mungkin belum dipublikasikan atau tautannya keliru.</p>
      </Centered>
    );

  const path = reachablePath(questions, rules, answers);
  const showProgress = form.settings.show_progress && started && !done && path.length > 0;
  const progress = path.length ? (Math.min(history.length, path.length) / path.length) * 100 : 0;
  const isLast = current ? nextQuestionId(questions, rules, answers, current.id) === null : false;

  return (
    <div style={rootStyle} className="font-sans relative flex min-h-dvh flex-col overflow-hidden bg-background text-foreground">
      {showProgress && (
        <div className="h-1 w-full bg-primary-soft">
          <div className="h-full bg-primary transition-all duration-300" style={{ width: `${progress}%` }} />
        </div>
      )}

      <div className="flex flex-1 items-center justify-center px-6 py-10">
        <div className="w-full max-w-[600px]">
          {!started ? (
            <Welcome
              title={form.settings.welcome_title || form.title}
              description={form.settings.welcome_description || form.description}
              banner={assetUrl(form.settings.banner_url)}
              logo={assetUrl(form.settings.logo_url)}
              startLabel={form.settings.start_button_text?.trim() || "Mulai"}
              questionCount={questions.length}
              onStart={start}
              empty={questions.length === 0}
            />
          ) : done ? (
            <ThankYou
              title={form.settings.thank_you_title || "Makasih, sudah terkirim!"}
              description={form.settings.thank_you_description || "Jawabanmu sudah kami catat."}
              logo={assetUrl(form.settings.logo_url)}
            />
          ) : (
            <AnimatePresence mode="wait">
              <motion.div
                key={currentId}
                initial={{ opacity: 0, y: 22 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -18 }}
                transition={ENTER}
              >
                {current && (
                  <QuestionScreen
                    question={current}
                    value={answers[current.id]}
                    onChange={(v) => setAnswer(current.id, v)}
                    onAdvance={() => scheduleAdvance()}
                    step={Math.min(history.length, path.length || history.length)}
                    total={Math.max(path.length, history.length)}
                  />
                )}

                {error && <p className="mt-4 text-sm text-destructive">{error}</p>}
                {submit.isError && <p className="mt-4 text-sm text-destructive">{(submit.error as Error).message}</p>}

                <div className="mt-9 flex items-center gap-4">
                  <Button onClick={next} size="lg" className="h-12 rounded-xl px-6 text-base" disabled={submit.isPending}>
                    {submit.isPending ? <Loader2 className="animate-spin" /> : isLast ? <Send /> : null}
                    {isLast ? "Kirim" : "OK"}
                    {!isLast && !submit.isPending && <ArrowRight />}
                  </Button>
                  <span className="hidden items-center gap-1 text-xs text-muted-foreground sm:flex">
                    tekan Enter <CornerDownLeft className="size-3" />
                  </span>
                  <Button variant="ghost" size="sm" className="ml-auto text-muted-foreground" onClick={back}>
                    <ArrowLeft /> Kembali
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
  banner,
  logo,
  startLabel,
  questionCount,
  onStart,
  empty,
}: {
  title: string;
  description?: string;
  banner?: string;
  logo?: string;
  startLabel: string;
  questionCount: number;
  onStart: () => void;
  empty: boolean;
}) {
  const mins = Math.max(1, Math.round(questionCount * 0.4));
  return (
    <div className="text-center">
      {banner && (
        <img src={banner} alt="" className="soft-card-shadow mb-7 max-h-56 w-full rounded-2xl border object-cover" />
      )}
      {logo ? (
        <img src={logo} alt="" className="mx-auto mb-5 size-16 rounded-2xl border object-cover" />
      ) : (
        <BrandMark size={56} className="mx-auto mb-5" />
      )}
      <div className="label-eyebrow text-brand">Survei</div>
      <h1 className="font-display mt-3 text-[34px] leading-[1.1] text-foreground sm:text-[42px]">{title}</h1>
      {description && <p className="text-body mx-auto mt-4 max-w-md text-lg">{description}</p>}
      {empty ? (
        <p className="text-body mt-8">Survei ini belum punya pertanyaan.</p>
      ) : (
        <>
          <Button size="lg" className="mt-8 h-12 rounded-xl px-7 text-base" onClick={onStart}>
            {startLabel} <ArrowRight />
          </Button>
          <p className="mt-5 text-[13px] text-muted-foreground">
            {questionCount} pertanyaan · ±{mins} menit · anonim
          </p>
        </>
      )}
    </div>
  );
}

function ThankYou({ title, description, logo }: { title: string; description?: string; logo?: string }) {
  return (
    <div className="relative text-center">
      <span className="animate-floaty pointer-events-none absolute -left-2 top-4 size-10 rounded-2xl bg-brand/20" />
      <span
        className="animate-floaty pointer-events-none absolute -right-1 top-20 size-7 rounded-full bg-primary/15"
        style={{ animationDelay: "1.2s" }}
      />
      {logo ? (
        <img src={logo} alt="" className="mx-auto mb-6 size-16 rounded-2xl border object-cover" />
      ) : (
        <div className="mx-auto mb-6 grid size-16 place-items-center rounded-full bg-primary">
          <Check className="size-8 text-primary-foreground" strokeWidth={2.5} />
        </div>
      )}
      <h1 className="font-display text-[32px] leading-tight text-foreground sm:text-[38px]">{title}</h1>
      {description && <p className="text-body mt-3 text-lg">{description}</p>}

      {/* Soft cross-promo: respondents can become creators. */}
      <div className="mt-10 border-t border-border pt-6">
        <p className="text-sm text-muted-foreground">Dibuat dengan txsurvey — alat survei gratis.</p>
        <a
          href={homePath}
          className="mt-2 inline-flex items-center gap-1 text-sm font-semibold text-primary transition-opacity hover:opacity-80"
        >
          Bikin surveimu sendiri <ArrowRight className="size-4" />
        </a>
      </div>
    </div>
  );
}

function Centered({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center gap-2 bg-background p-6 text-center">
      {children}
    </div>
  );
}
