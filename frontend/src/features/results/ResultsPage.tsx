import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, Download, Loader2, Pencil, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { FullScreenLoader } from "@/components/FullScreenLoader";
import { cn } from "@/lib/utils";
import { formatAnswer } from "@/lib/formatAnswer";
import { themeStyle } from "@/lib/themes";
import { useForm } from "@/api/forms";
import { csvUrl, useAnalytics, useDeleteResponses, useFunnel, useResponses } from "@/api/results";
import { useDocumentTitle } from "@/lib/useDocumentTitle";
import type { AnswerValue, FormAnalytics, FormFunnel, Question, ResponseItem } from "@/types/forms";

export function ResultsPage() {
  const { id = "" } = useParams();
  const { data: form, isLoading } = useForm(id);
  useDocumentTitle("Hasil", form?.title);
  const analytics = useAnalytics(id);
  const funnel = useFunnel(id);
  const responses = useResponses(id);
  const deleteResponses = useDeleteResponses(id);
  const [tab, setTab] = useState<"summary" | "responses">("summary");
  const [confirmClear, setConfirmClear] = useState(false);

  if (isLoading || !form) return <FullScreenLoader />;

  const responseCount = analytics.data?.response_count ?? 0;

  return (
    <div style={themeStyle(form.settings.theme, form.settings.font)} className="font-sans min-h-dvh bg-background">
      <div className="border-b bg-card">
        <div className="mx-auto flex max-w-[980px] flex-wrap items-center gap-3 px-6 py-3">
          <Button variant="ghost" size="icon" asChild>
            <Link to="/" aria-label="Kembali ke survei">
              <ArrowLeft />
            </Link>
          </Button>
          <div className="min-w-0">
            <h1 className="font-display truncate text-lg text-foreground">{form.title}</h1>
            <p className="text-xs text-muted-foreground">
              {analytics.data?.response_count ?? 0} respons
            </p>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <Button variant="outline" size="sm" className="rounded-lg" asChild>
              <Link to={`/forms/${id}`}>
                <Pencil /> Edit
              </Link>
            </Button>
            {responseCount > 0 && (
              <Button
                variant="outline"
                size="sm"
                className="rounded-lg text-destructive hover:text-destructive"
                onClick={() => setConfirmClear(true)}
                disabled={deleteResponses.isPending}
              >
                {deleteResponses.isPending ? <Loader2 className="animate-spin" /> : <Trash2 />} Hapus data
              </Button>
            )}
            <Button size="sm" className="rounded-lg" asChild>
              <a href={csvUrl(id)} download>
                <Download /> Unduh CSV
              </a>
            </Button>
          </div>
        </div>
        <div className="mx-auto flex max-w-[980px] gap-1 px-6">
          <TabButton active={tab === "summary"} onClick={() => setTab("summary")}>
            Ringkasan
          </TabButton>
          <TabButton active={tab === "responses"} onClick={() => setTab("responses")}>
            Respons
          </TabButton>
        </div>
      </div>

      <main className="mx-auto max-w-[980px] px-6 py-6">
        {tab === "summary" ? (
          <div className="space-y-4">
            {funnel.isLoading ? (
              <p className="text-muted-foreground">Memuat…</p>
            ) : funnel.data ? (
              <FunnelSection data={funnel.data} />
            ) : null}

            {analytics.isLoading ? (
              <p className="text-muted-foreground">Memuat…</p>
            ) : analytics.data && analytics.data.response_count > 0 ? (
              <AnalyticsView data={analytics.data} />
            ) : (
              <EmptyState />
            )}
          </div>
        ) : responses.isLoading ? (
          <p className="text-muted-foreground">Memuat…</p>
        ) : responses.data && responses.data.length > 0 ? (
          <ResponsesTable responses={responses.data} questions={form.questions ?? []} />
        ) : (
          <EmptyState />
        )}
      </main>

      <ConfirmDialog
        open={confirmClear}
        onOpenChange={setConfirmClear}
        title="Hapus semua data respons?"
        description={`${responseCount} respons beserta jawabannya akan dihapus permanen. Pertanyaan dan pengaturan survei tetap utuh. Tindakan ini tidak bisa dibatalkan.`}
        confirmText="Hapus data"
        cancelText="Batal"
        destructive
        onConfirm={() => {
          deleteResponses.mutate();
          setConfirmClear(false);
        }}
      />
    </div>
  );
}

function TabButton({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "relative -mb-px rounded-t-lg px-4 py-2 text-sm font-medium transition-colors",
        active
          ? "border border-b-0 border-border bg-background text-foreground"
          : "text-muted-foreground hover:text-foreground",
      )}
    >
      {children}
    </button>
  );
}

function EmptyState() {
  return (
    <Card className="rounded-2xl">
      <CardContent className="py-16 text-center text-muted-foreground">Belum ada respons.</CardContent>
    </Card>
  );
}

/** FunnelSection shows the drop-off funnel: one horizontal bar per question,
 *  width proportional to reached/starts. `starts` counts every opened session
 *  (completed + abandoned); guard against divide-by-zero when nobody has
 *  started yet. */
function FunnelSection({ data }: { data: FormFunnel }) {
  return (
    <Card className="rounded-2xl">
      <CardHeader className="pb-3">
        <CardTitle className="font-display text-lg font-medium">Funnel</CardTitle>
        <p className="text-sm text-muted-foreground">
          {data.starts === 0
            ? "Belum ada sesi dibuka."
            : `${data.starts} mulai · ${data.completed} selesai · ${Math.round(
                (data.completed / data.starts) * 100,
              )}% penyelesaian`}
        </p>
      </CardHeader>
      <CardContent>
        {data.starts === 0 ? (
          <p className="text-sm text-muted-foreground">Belum ada data.</p>
        ) : (
          <div className="space-y-2.5">
            {data.steps.map((s) => {
              const pct = (s.reached / data.starts) * 100;
              return (
                <div key={s.question_id} className="flex items-center gap-3 text-sm">
                  <span className="w-32 shrink-0 truncate text-foreground" title={s.title || "Tanpa judul"}>
                    {s.title || "Tanpa judul"}
                  </span>
                  <div className="h-6 flex-1 overflow-hidden rounded-lg bg-primary-soft">
                    <div className="h-full rounded-lg bg-primary" style={{ width: `${pct}%` }} />
                  </div>
                  <span className="w-24 shrink-0 text-right tabular-nums text-muted-foreground">
                    {s.reached} · {Math.round(pct)}%
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function AnalyticsView({ data }: { data: FormAnalytics }) {
  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-3">
        <Stat label="Respons" value={String(data.response_count)} />
        <Stat label="Penyelesaian" value={`${Math.round(data.completion_rate * 100)}%`} accent />
        <Stat label="Pertanyaan" value={String(data.questions.length)} />
      </div>

      {data.questions.map((q) => (
        <Card key={q.question_id} className="rounded-2xl">
          <CardHeader className="pb-3">
            <CardTitle className="font-display text-lg font-medium">{q.title || "Tanpa judul"}</CardTitle>
            <p className="text-sm text-muted-foreground">
              {q.answered} jawaban
            </p>
          </CardHeader>
          <CardContent>
            {q.options && q.options.length > 0 ? (
              <BarList options={q.options} />
            ) : q.average !== undefined ? (
              <div className="flex items-baseline gap-2">
                <span className="font-display text-4xl text-primary">{q.average.toFixed(2)}</span>
                <span className="text-sm text-muted-foreground">rata-rata</span>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                {q.answered > 0 ? "Jawaban terbuka — lihat tab Respons." : "Belum ada jawaban."}
              </p>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function BarList({ options }: { options: { label: string; count: number }[] }) {
  const total = Math.max(1, options.reduce((n, o) => n + o.count, 0));
  const max = Math.max(1, ...options.map((o) => o.count));
  // Fill opacity decreases by rank (most-picked is strongest), purely visual.
  const ranked = [...options].sort((a, b) => b.count - a.count);
  const rankOf = (o: { label: string; count: number }) => ranked.indexOf(o);

  return (
    <div className="space-y-2.5">
      {options.map((o, i) => {
        const opacity = Math.max(0.4, 1 - rankOf(o) * 0.13);
        return (
          <div key={i} className="flex items-center gap-3 text-sm">
            <span className="w-32 shrink-0 truncate text-foreground">{o.label}</span>
            <div className="h-6 flex-1 overflow-hidden rounded-lg bg-primary-soft">
              <div
                className="h-full rounded-lg bg-primary"
                style={{ width: `${(o.count / max) * 100}%`, opacity }}
              />
            </div>
            <span className="w-16 shrink-0 text-right tabular-nums text-muted-foreground">
              {o.count} · {Math.round((o.count / total) * 100)}%
            </span>
          </div>
        );
      })}
    </div>
  );
}

function Stat({ label, value, accent }: { label: string; value: string; accent?: boolean }) {
  return (
    <Card className="rounded-2xl">
      <CardContent className="py-5">
        <p className="label-eyebrow text-muted-foreground">{label}</p>
        <p className={cn("font-display mt-1 text-3xl", accent ? "text-primary" : "text-foreground")}>{value}</p>
      </CardContent>
    </Card>
  );
}

function ResponsesTable({ responses, questions }: { responses: ResponseItem[]; questions: Question[] }) {
  const cols = useMemo(() => questions.filter((q) => q.type !== "statement"), [questions]);
  const byId = useMemo(() => new Map(cols.map((q) => [q.id, q])), [cols]);

  return (
    <Card className="overflow-hidden rounded-2xl">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="border-b bg-background">
            <tr>
              <th className="whitespace-nowrap px-4 py-2.5 text-left font-medium text-muted-foreground">Dikirim</th>
              {cols.map((q) => (
                <th key={q.id} className="whitespace-nowrap px-4 py-2.5 text-left font-medium text-muted-foreground">
                  {q.title || "Tanpa judul"}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {responses.map((r) => {
              const answers = new Map(r.answers.map((a) => [a.question_id, a.value as AnswerValue]));
              return (
                <tr key={r.id} className="border-b last:border-0">
                  <td className="whitespace-nowrap px-4 py-2.5 text-muted-foreground">
                    {new Date(r.submitted_at).toLocaleString()}
                  </td>
                  {cols.map((q) => (
                    <td key={q.id} className="px-4 py-2.5 text-foreground">
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
