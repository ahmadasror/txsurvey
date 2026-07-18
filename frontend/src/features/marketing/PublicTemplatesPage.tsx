import { Link } from "react-router-dom";
import { ArrowRight, Clock3, ListChecks } from "lucide-react";
import { SURVEY_TEMPLATES } from "@/lib/surveyTemplates";
import { presetById } from "@/lib/themes";
import { usePageMetadata } from "@/lib/usePageMetadata";

export function PublicTemplatesPage() {
  usePageMetadata({
    title: "Template Survei Gratis untuk Tim",
    description: "Contoh template survei kepuasan karyawan, feedback onboarding, dan kepuasan layanan IT yang siap disesuaikan.",
    robots: "index, follow",
    path: "contoh-template-survei",
  });

  return (
    <main>
      <section className="mx-auto max-w-[900px] px-6 pb-12 pt-16 text-center sm:pt-20">
        <div className="label-eyebrow text-brand">Template siap pakai</div>
        <h1 className="font-display mt-4 text-[40px] leading-tight sm:text-[52px]">Mulai dari pertanyaan yang sudah teruji.</h1>
        <p className="text-body mx-auto mt-5 max-w-2xl text-lg leading-relaxed">
          Pilih tujuan surveimu, lihat seluruh pertanyaannya, lalu gunakan sebagai titik awal. Semua template bisa diubah setelah masuk.
        </p>
      </section>

      <section className="mx-auto grid max-w-[1080px] gap-5 px-6 md:grid-cols-3" aria-label="Daftar template survei">
        {SURVEY_TEMPLATES.map((template) => {
          const swatch = presetById(template.theme)?.swatch;
          const minutes = Math.max(1, Math.round(template.questions.length * 0.4));
          return (
            <article key={template.id} className="flex flex-col rounded-2xl border bg-card p-6 soft-card-shadow">
              <span className="size-3 rounded-full" style={{ background: swatch }} />
              <h2 className="font-display mt-5 text-2xl">{template.title}</h2>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{template.description}</p>
              <div className="mt-5 flex flex-wrap gap-3 text-xs text-muted-foreground">
                <span className="inline-flex items-center gap-1.5"><ListChecks className="size-4" /> {template.questions.length} pertanyaan</span>
                <span className="inline-flex items-center gap-1.5"><Clock3 className="size-4" /> ±{minutes} menit</span>
              </div>
              <Link className="mt-7 inline-flex items-center gap-2 font-semibold text-primary" to={`/${template.publicPath}`}>
                Lihat template <ArrowRight className="size-4" />
              </Link>
            </article>
          );
        })}
      </section>

      <section className="mx-auto mt-16 max-w-[900px] px-6">
        <div className="rounded-3xl bg-primary px-7 py-10 text-primary-foreground sm:px-12">
          <h2 className="font-display text-3xl">Butuh alur yang berbeda untuk tiap jawaban?</h2>
          <p className="mt-3 max-w-2xl text-primary-foreground/80">
            Pelajari cara logika bercabang melewati pertanyaan yang tidak relevan tanpa membuat responden bingung.
          </p>
          <Link className="mt-5 inline-flex items-center gap-2 font-semibold" to="/fitur/logika-bercabang">
            Pelajari logika bercabang <ArrowRight className="size-4" />
          </Link>
        </div>
      </section>
    </main>
  );
}
