import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowRight, Headphones, Loader2, Sparkles, Users } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { api } from "@/api/client";
import { presetById } from "@/lib/themes";
import { typeLabel } from "@/lib/questionTypes";
import { useCreateForm } from "@/api/forms";
import { useDocumentTitle } from "@/lib/useDocumentTitle";
import type { QuestionInput, QuestionType } from "@/types/forms";

interface Template {
  id: string;
  icon: LucideIcon;
  theme: string;
  font: string;
  title: string;
  description: string;
  welcomeDesc: string;
  questions: QuestionInput[];
}

const choice = (title: string, labels: string[], required = true): QuestionInput => ({
  type: "multiple_choice",
  title,
  required,
  metadata: { options: labels.map((label) => ({ id: "", label })) },
});
const rating = (title: string, scale: number): QuestionInput => ({ type: "rating", title, required: true, metadata: { scale } });
const yesno = (title: string): QuestionInput => ({ type: "yes_no", title, required: true, metadata: {} });
const long = (title: string, required = false): QuestionInput => ({ type: "long_text", title, required, metadata: {} });

const TEMPLATES: Template[] = [
  {
    id: "pulse",
    icon: Users,
    theme: "pine",
    font: "editorial",
    title: "Pulse Karyawan",
    description: "Ukur kepuasan & loyalitas tim dengan beberapa pertanyaan singkat.",
    welcomeDesc: "Beberapa pertanyaan singkat soal pengalaman kerjamu. Anonim.",
    questions: [
      choice("Sudah berapa lama kamu bergabung?", ["< 6 bulan", "6–12 bulan", "1–3 tahun", "> 3 tahun"]),
      choice("Seberapa puas kamu bekerja di sini?", ["Sangat puas", "Puas", "Biasa aja", "Kurang puas", "Tidak puas"]),
      rating("Seberapa besar kemungkinan kamu merekomendasikan tempat ini ke teman?", 10),
      long("Apa satu hal yang bisa kami perbaiki?"),
      yesno("Boleh kami follow-up jawabanmu?"),
    ],
  },
  {
    id: "onboarding",
    icon: Sparkles,
    theme: "sand",
    font: "editorial",
    title: "Feedback Onboarding",
    description: "Pahami pengalaman karyawan baru di minggu-minggu pertama.",
    welcomeDesc: "Bantu kami memperbaiki proses onboarding untuk yang berikutnya.",
    questions: [
      rating("Seberapa jelas peran & tanggung jawabmu setelah onboarding?", 5),
      choice("Apakah pelatihan yang diberikan cukup?", ["Lebih dari cukup", "Cukup", "Kurang", "Tidak ada pelatihan"]),
      yesno("Apakah mentormu membantu?"),
      long("Bagian onboarding mana yang paling membantu?"),
      long("Apa yang sebaiknya kami tambahkan?"),
    ],
  },
  {
    id: "it-csat",
    icon: Headphones,
    theme: "ink",
    font: "soft",
    title: "Kepuasan Layanan IT",
    description: "Nilai kecepatan & kualitas dukungan tim IT internal.",
    welcomeDesc: "Ceritakan pengalamanmu dengan layanan IT terakhir kali.",
    questions: [
      choice("Layanan IT mana yang kamu gunakan?", ["Helpdesk", "Perbaikan perangkat", "Akses & akun", "Jaringan", "Lainnya"]),
      yesno("Apakah masalahmu terselesaikan?"),
      rating("Seberapa cepat penanganannya?", 5),
      choice("Secara keseluruhan, seberapa puas kamu?", ["Sangat puas", "Puas", "Biasa aja", "Tidak puas"]),
      long("Saran untuk tim IT?"),
    ],
  },
];

const uniqueTypes = (qs: QuestionInput[]): QuestionType[] => [...new Set(qs.map((q) => q.type))];

export function TemplatesPage() {
  useDocumentTitle("Template");
  const navigate = useNavigate();
  const createForm = useCreateForm();
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const use = async (t: Template) => {
    if (busy) return;
    setBusy(t.id);
    setError(null);
    try {
      const form = await createForm.mutateAsync(t.title);
      await api(`/forms/${form.id}`, {
        method: "PATCH",
        body: JSON.stringify({
          title: t.title,
          description: t.description,
          settings: {
            show_progress: true,
            theme: { preset: t.theme },
            font: t.font,
            welcome_title: t.title,
            welcome_description: t.welcomeDesc,
          },
        }),
      });
      // Seed questions in order (positions assigned by the backend).
      for (const q of t.questions) {
        await api(`/forms/${form.id}/questions`, { method: "POST", body: JSON.stringify(q) });
      }
      navigate(`/forms/${form.id}`);
    } catch (e) {
      setError((e as Error).message || "Gagal membuat survei dari template.");
      setBusy(null);
    }
  };

  return (
    <main className="mx-auto max-w-[980px] px-6 py-9">
      <div className="mb-7">
        <h1 className="font-display text-[28px] leading-tight text-foreground">Mulai dari template</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Survei siap pakai — sudah terisi pertanyaan, tema, dan font. Tinggal sesuaikan.
        </p>
      </div>

      {error && <p className="mb-4 text-sm text-destructive">{error}</p>}

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        {TEMPLATES.map((t) => {
          const Icon = t.icon;
          const swatch = presetById(t.theme)?.swatch;
          const mins = Math.max(1, Math.round(t.questions.length * 0.4));
          return (
            <div key={t.id} className="flex flex-col rounded-2xl border border-border bg-card p-5">
              <span
                className="grid size-11 place-items-center rounded-xl text-primary-foreground"
                style={{ background: swatch }}
              >
                <Icon className="size-5" />
              </span>
              <h2 className="font-display mt-4 text-lg text-foreground">{t.title}</h2>
              <p className="mt-1 text-sm text-muted-foreground">{t.description}</p>

              <div className="mt-3 flex flex-wrap gap-1.5">
                {uniqueTypes(t.questions).map((ty) => (
                  <span key={ty} className="rounded-full bg-primary-soft px-2.5 py-0.5 text-xs font-medium text-primary">
                    {typeLabel(ty)}
                  </span>
                ))}
              </div>

              <div className="mt-5 flex items-center justify-between border-t pt-4">
                <span className="text-xs text-muted-foreground">
                  {t.questions.length} pertanyaan · ±{mins} menit
                </span>
                <button
                  onClick={() => use(t)}
                  disabled={!!busy}
                  className="inline-flex items-center gap-1 text-sm font-semibold text-primary transition-opacity hover:opacity-80 disabled:opacity-50"
                >
                  {busy === t.id ? <Loader2 className="size-4 animate-spin" /> : null}
                  Pakai <ArrowRight className="size-4" />
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </main>
  );
}
