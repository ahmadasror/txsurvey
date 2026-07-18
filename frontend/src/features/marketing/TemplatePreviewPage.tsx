import { Link, useLocation } from "react-router-dom";
import { ArrowRight, CheckCircle2, Clock3 } from "lucide-react";
import { templateByPath } from "@/lib/surveyTemplates";
import { typeLabel } from "@/lib/questionTypes";
import { usePageMetadata } from "@/lib/usePageMetadata";

export function TemplatePreviewPage() {
  const location = useLocation();
  const path = location.pathname.split("/").filter(Boolean).pop();
  const template = templateByPath(path);

  usePageMetadata({
    title: template?.seoTitle || "Template Survei",
    description: template?.description || "Contoh template survei gratis dari txsurvey.",
    robots: template ? "index, follow" : "noindex, nofollow",
    path: template?.publicPath || path,
  });

  if (!template) {
    return (
      <main className="mx-auto max-w-[760px] px-6 py-20 text-center">
        <h1 className="font-display text-4xl">Template tidak ditemukan</h1>
        <Link className="mt-6 inline-flex items-center gap-2 font-semibold text-primary" to="/contoh-template-survei">
          Lihat semua template <ArrowRight className="size-4" />
        </Link>
      </main>
    );
  }

  const minutes = Math.max(1, Math.round(template.questions.length * 0.4));

  return (
    <main>
      <section className="mx-auto max-w-[900px] px-6 pb-12 pt-16">
        <div className="label-eyebrow text-brand">Template survei gratis</div>
        <h1 className="font-display mt-4 max-w-3xl text-[40px] leading-tight sm:text-[52px]">{template.seoTitle}</h1>
        <p className="text-body mt-5 max-w-2xl text-lg leading-relaxed">{template.description}</p>
        <div className="mt-6 flex flex-wrap gap-4 text-sm text-muted-foreground">
          <span className="inline-flex items-center gap-2"><CheckCircle2 className="size-4 text-primary" /> {template.questions.length} pertanyaan siap edit</span>
          <span className="inline-flex items-center gap-2"><Clock3 className="size-4 text-primary" /> Waktu isi ±{minutes} menit</span>
        </div>
        <Link className="mt-8 inline-flex h-12 items-center gap-2 rounded-xl bg-primary px-6 font-semibold text-primary-foreground" to="/login">
          Gunakan template ini <ArrowRight className="size-4" />
        </Link>
      </section>

      <section className="mx-auto grid max-w-[900px] gap-8 px-6 lg:grid-cols-[1fr_280px]">
        <div>
          <h2 className="font-display text-3xl">Pertanyaan dalam template</h2>
          <ol className="mt-6 space-y-4">
            {template.questions.map((question, index) => (
              <li key={question.title} className="rounded-2xl border bg-card p-5">
                <div className="flex items-start gap-4">
                  <span className="font-display text-lg text-brand">{String(index + 1).padStart(2, "0")}</span>
                  <div>
                    <h3 className="font-medium text-foreground">{question.title}</h3>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {typeLabel(question.type)}{question.required ? " · wajib" : " · opsional"}
                    </p>
                  </div>
                </div>
              </li>
            ))}
          </ol>
        </div>
        <aside className="h-fit rounded-2xl bg-primary-soft p-6">
          <h2 className="font-display text-xl text-primary">Cocok untuk</h2>
          <p className="mt-3 text-sm leading-relaxed text-body">{template.audience}</p>
          <h2 className="font-display mt-7 text-xl text-primary">Cara menggunakan</h2>
          <p className="mt-3 text-sm leading-relaxed text-body">
            Masuk, buat salinan template, sesuaikan bahasa dan tema, lalu bagikan tautan publik kepada responden.
          </p>
        </aside>
      </section>
    </main>
  );
}
