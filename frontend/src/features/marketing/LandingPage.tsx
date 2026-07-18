import { Navigate, Link } from "react-router-dom";
import { ArrowRight, BarChart3, GitBranch, MessageSquareText, ShieldCheck } from "lucide-react";
import { useMe } from "@/api/auth";
import { usePageMetadata } from "@/lib/usePageMetadata";
import { SURVEY_TEMPLATES } from "@/lib/surveyTemplates";

const BENEFITS = [
  {
    icon: MessageSquareText,
    title: "Satu pertanyaan per layar",
    body: "Responden fokus pada satu hal, bukan menghadapi formulir panjang yang terasa seperti pekerjaan rumah.",
  },
  {
    icon: GitBranch,
    title: "Logika bercabang",
    body: "Lewati pertanyaan yang tidak relevan dan arahkan setiap responden ke alur yang sesuai jawabannya.",
  },
  {
    icon: BarChart3,
    title: "Hasil yang langsung terbaca",
    body: "Lihat respons, ringkasan, funnel penyelesaian, dan ekspor CSV tanpa spreadsheet manual.",
  },
];

export function LandingPage() {
  const { data: user, isLoading } = useMe();
  usePageMetadata({
    title: "Survei Online dengan Logika Bercabang",
    description: "Buat survei online yang terasa seperti percakapan, lengkap dengan logika bercabang, tema hangat, respons anonim, dan analitik.",
    robots: "index, follow",
    path: "",
    image: "og-image.png",
  });

  if (!isLoading && user) return <Navigate to="/app" replace />;

  return (
    <main>
      <section className="relative overflow-hidden">
        <span className="pointer-events-none absolute -right-16 top-12 size-64 rounded-full bg-primary-soft" />
        <span className="pointer-events-none absolute -left-20 bottom-0 size-52 rounded-full bg-brand/10" />
        <div className="relative mx-auto max-w-[1000px] px-6 pb-20 pt-20 text-center sm:pt-28">
          <div className="label-eyebrow text-brand">Survei yang diisi sampai habis</div>
          <h1 className="font-display mx-auto mt-5 max-w-4xl text-[46px] leading-[1.04] sm:text-[68px]">
            Bikin survei yang terasa seperti ngobrol.
          </h1>
          <p className="text-body mx-auto mt-6 max-w-2xl text-lg leading-relaxed sm:text-xl">
            Satu pertanyaan per layar, alur bercabang yang relevan, dan hasil yang langsung bisa dipakai mengambil keputusan.
          </p>
          <div className="mt-9 flex flex-wrap justify-center gap-4">
            <Link className="inline-flex h-12 items-center gap-2 rounded-xl bg-primary px-6 font-semibold text-primary-foreground" to="/login">
              Buat survei gratis <ArrowRight className="size-4" />
            </Link>
            <Link className="inline-flex h-12 items-center rounded-xl border bg-card px-6 font-semibold text-primary" to="/contoh-template-survei">
              Lihat template
            </Link>
          </div>
          <p className="mt-5 text-sm text-muted-foreground">Gratis untuk tim kecil · tanpa kartu kredit</p>
        </div>
      </section>

      <section className="mx-auto grid max-w-[1080px] gap-5 px-6 md:grid-cols-3" aria-label="Keunggulan txsurvey">
        {BENEFITS.map(({ icon: Icon, title, body }) => (
          <article key={title} className="rounded-2xl border bg-card p-6">
            <span className="grid size-11 place-items-center rounded-xl bg-primary-soft text-primary"><Icon className="size-5" /></span>
            <h2 className="font-display mt-5 text-2xl">{title}</h2>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">{body}</p>
          </article>
        ))}
      </section>

      <section className="mx-auto mt-20 max-w-[1080px] px-6">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <div className="label-eyebrow text-brand">Tidak perlu mulai kosong</div>
            <h2 className="font-display mt-3 text-3xl sm:text-4xl">Template untuk pertanyaan yang umum.</h2>
          </div>
          <Link className="inline-flex items-center gap-2 font-semibold text-primary" to="/contoh-template-survei">
            Semua template <ArrowRight className="size-4" />
          </Link>
        </div>
        <div className="mt-7 grid gap-4 md:grid-cols-3">
          {SURVEY_TEMPLATES.map((template) => (
            <Link key={template.id} to={`/${template.publicPath}`} className="rounded-2xl bg-primary-soft p-5 transition-transform hover:-translate-y-1">
              <h3 className="font-display text-xl text-primary">{template.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-body">{template.description}</p>
            </Link>
          ))}
        </div>
      </section>

      <section className="mx-auto mt-20 max-w-[900px] px-6">
        <div className="rounded-3xl bg-primary px-7 py-12 text-center text-primary-foreground sm:px-12">
          <ShieldCheck className="mx-auto size-8 text-brand" />
          <h2 className="font-display mt-4 text-3xl sm:text-4xl">Feedback jujur dimulai dari rasa aman.</h2>
          <p className="mx-auto mt-4 max-w-2xl text-primary-foreground/80">
            Responden tidak perlu login. Hasil hanya tersedia bagi creator, dan kamu bisa menjelaskan anonimitas sejak layar pembuka.
          </p>
          <Link className="mt-6 inline-flex items-center gap-2 font-semibold" to="/fitur/survei-anonim">
            Pelajari survei anonim <ArrowRight className="size-4" />
          </Link>
        </div>
      </section>
    </main>
  );
}
