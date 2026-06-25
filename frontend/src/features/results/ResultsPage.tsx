import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, Download, Pencil } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FullScreenLoader } from "@/components/FullScreenLoader";
import { cn } from "@/lib/utils";
import { formatAnswer } from "@/lib/formatAnswer";
import { useForm } from "@/api/forms";
import { csvUrl, useAnalytics, useResponses } from "@/api/results";
import type { AnswerValue, FormAnalytics, Question, ResponseItem } from "@/types/forms";

export function ResultsPage() {
  const { id = "" } = useParams();
  const { data: form, isLoading } = useForm(id);
  const analytics = useAnalytics(id);
  const responses = useResponses(id);
  const [tab, setTab] = useState<"summary" | "responses">("summary");

  if (isLoading || !form) return <FullScreenLoader />;

  return (
    <div className="min-h-dvh bg-muted/30">
      <div className="border-b bg-background">
        <div className="container flex flex-wrap items-center gap-3 py-3">
          <Button variant="ghost" size="icon" asChild>
            <Link to="/" aria-label="Back to forms">
              <ArrowLeft />
            </Link>
          </Button>
          <div className="min-w-0">
            <h1 className="truncate text-base font-semibold">{form.title}</h1>
            <p className="text-xs text-muted-foreground">
              {analytics.data?.response_count ?? 0} response
              {(analytics.data?.response_count ?? 0) === 1 ? "" : "s"}
            </p>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <Button variant="outline" size="sm" asChild>
              <Link to={`/forms/${id}`}>
                <Pencil /> Edit
              </Link>
            </Button>
            <Button size="sm" asChild>
              <a href={csvUrl(id)} download>
                <Download /> Download CSV
              </a>
            </Button>
          </div>
        </div>
        <div className="container flex gap-1 pb-2">
          <TabButton active={tab === "summary"} onClick={() => setTab("summary")}>
            Summary
          </TabButton>
          <TabButton active={tab === "responses"} onClick={() => setTab("responses")}>
            Responses
          </TabButton>
        </div>
      </div>

      <main className="container py-6">
        {tab === "summary" ? (
          analytics.isLoading ? (
            <p className="text-muted-foreground">Loading…</p>
          ) : analytics.data && analytics.data.response_count > 0 ? (
            <AnalyticsView data={analytics.data} />
          ) : (
            <EmptyState />
          )
        ) : responses.isLoading ? (
          <p className="text-muted-foreground">Loading…</p>
        ) : responses.data && responses.data.length > 0 ? (
          <ResponsesTable responses={responses.data} questions={form.questions ?? []} />
        ) : (
          <EmptyState />
        )}
      </main>
    </div>
  );
}

function TabButton({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
        active ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-accent",
      )}
    >
      {children}
    </button>
  );
}

function EmptyState() {
  return (
    <Card>
      <CardContent className="py-16 text-center text-muted-foreground">No responses yet.</CardContent>
    </Card>
  );
}

function AnalyticsView({ data }: { data: FormAnalytics }) {
  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-3">
        <Stat label="Responses" value={String(data.response_count)} />
        <Stat label="Completion" value={`${Math.round(data.completion_rate * 100)}%`} />
        <Stat label="Questions" value={String(data.questions.length)} />
      </div>

      {data.questions.map((q) => (
        <Card key={q.question_id}>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">{q.title || "Untitled"}</CardTitle>
            <p className="text-sm text-muted-foreground">
              {q.answered} answer{q.answered === 1 ? "" : "s"}
              {q.average !== undefined ? ` · average ${q.average.toFixed(2)}` : ""}
            </p>
          </CardHeader>
          <CardContent>
            {q.options && q.options.length > 0 ? (
              <BarList options={q.options} />
            ) : (
              <p className="text-sm text-muted-foreground">
                {q.answered > 0 ? "Open-ended — see the Responses tab." : "No answers yet."}
              </p>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function BarList({ options }: { options: { label: string; count: number }[] }) {
  const max = Math.max(1, ...options.map((o) => o.count));
  return (
    <div className="space-y-2">
      {options.map((o, i) => (
        <div key={i} className="flex items-center gap-3 text-sm">
          <span className="w-32 shrink-0 truncate">{o.label}</span>
          <div className="h-5 flex-1 overflow-hidden rounded bg-muted">
            <div className="h-full rounded bg-primary" style={{ width: `${(o.count / max) * 100}%` }} />
          </div>
          <span className="w-8 shrink-0 text-right tabular-nums text-muted-foreground">{o.count}</span>
        </div>
      ))}
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <Card>
      <CardContent className="py-4">
        <p className="text-sm text-muted-foreground">{label}</p>
        <p className="text-2xl font-semibold">{value}</p>
      </CardContent>
    </Card>
  );
}

function ResponsesTable({ responses, questions }: { responses: ResponseItem[]; questions: Question[] }) {
  const cols = useMemo(() => questions.filter((q) => q.type !== "statement"), [questions]);
  const byId = useMemo(() => new Map(cols.map((q) => [q.id, q])), [cols]);

  return (
    <Card>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="border-b bg-muted/50">
            <tr>
              <th className="whitespace-nowrap px-4 py-2 text-left font-medium">Submitted</th>
              {cols.map((q) => (
                <th key={q.id} className="whitespace-nowrap px-4 py-2 text-left font-medium">
                  {q.title || "Untitled"}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {responses.map((r) => {
              const answers = new Map(r.answers.map((a) => [a.question_id, a.value as AnswerValue]));
              return (
                <tr key={r.id} className="border-b last:border-0">
                  <td className="whitespace-nowrap px-4 py-2 text-muted-foreground">
                    {new Date(r.submitted_at).toLocaleString()}
                  </td>
                  {cols.map((q) => (
                    <td key={q.id} className="px-4 py-2">
                      {formatAnswer(byId.get(q.id), answers.get(q.id))}
                    </td>
                  ))}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
