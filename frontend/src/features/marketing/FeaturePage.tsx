import { Link, useParams } from "react-router-dom";
import { ArrowRight, GitBranch, ShieldCheck } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { usePageMetadata } from "@/lib/usePageMetadata";

interface FeatureContent {
  slug: string;
  eyebrow: string;
  title: string;
  description: string;
  intro: string;
  icon: LucideIcon;
  benefits: Array<{ title: string; body: string }>;
  steps: string[];
}

const FEATURES: FeatureContent[] = [
  {
    slug: "logika-bercabang",
    eyebrow: "Conditional logic",
    title: "Survei dengan logika bercabang",
    description: "Tampilkan pertanyaan yang relevan berdasarkan jawaban responden dengan alur lompat yang aman dan mudah disusun.",
    intro: "Tidak semua responden perlu melihat pertanyaan yang sama. Logika bercabang membuat satu survei terasa lebih singkat karena setiap orang hanya melewati jalur yang sesuai dengan jawabannya.",
    icon: GitBranch,
    benefits: [
      { title: "Lebih relevan", body: "Pertanyaan lanjutan hanya muncul ketika jawaban sebelumnya memang membutuhkannya." },
      { title: "Lebih cepat selesai", body: "Responden tidak dipaksa membaca bagian yang tidak berhubungan dengan pengalaman mereka." },
      { title: "Data lebih bersih", body: "Jawaban dari pertanyaan yang dilewati tidak ikut masuk ke hasil dan analitik." },
    ],
    steps: [
      "Buat pertanyaan sumber dan pertanyaan tujuan.",
      "Pilih kondisi jawaban yang memicu lompatan.",
      "Arahkan lompatan ke pertanyaan yang posisinya lebih akhir.",
      "Preview alur sebelum membagikan tautan survei.",
    ],
  },
  {
    slug: "survei-anonim",
    eyebrow: "Privasi responden",
    title: "Survei anonim untuk feedback yang lebih jujur",
    description: "Kumpulkan respons tanpa meminta identitas responden dan batasi akses hasil hanya kepada pemilik survei.",
    intro: "Feedback sensitif lebih mudah diberikan ketika responden memahami apa yang dikumpulkan. txsurvey tidak meminta akun, email, atau nama untuk mengisi survei publik secara default.",
    icon: ShieldCheck,
    benefits: [
      { title: "Tanpa login responden", body: "Tautan publik dapat langsung dibuka dan diisi tanpa membuat akun." },
      { title: "Hasil owner-only", body: "Daftar respons, analitik, dan ekspor CSV berada di area creator yang membutuhkan session." },
      { title: "Transparan", body: "Jelaskan tujuan dan kebijakan follow-up di halaman pembuka sebelum responden mulai." },
    ],
    steps: [
      "Jelaskan tujuan survei pada deskripsi pembuka.",
      "Hindari pertanyaan identitas jika tidak benar-benar diperlukan.",
      "Gunakan pertanyaan opsional untuk feedback terbuka.",
      "Bagikan hasil hanya kepada orang yang berwenang.",
    ],
  },
];

export function FeaturePage() {
  const { featureSlug } = useParams();
  const feature = FEATURES.find((item) => item.slug === featureSlug);

  usePageMetadata({
    title: feature?.title || "Fitur",
    description: feature?.description || "Fitur txsurvey untuk membuat survei yang lebih relevan.",
    robots: feature ? "index, follow" : "noindex, nofollow",
    path: `fitur/${featureSlug || ""}`,
  });

  if (!feature) {
    return (
      <main className="mx-auto max-w-[760px] px-6 py-20 text-center">
        <h1 className="font-display text-4xl">Fitur tidak ditemukan</h1>
        <Link className="mt-6 inline-flex items-center gap-2 font-semibold text-primary" to="/panduan">
          Buka panduan <ArrowRight className="size-4" />
        </Link>
      </main>
    );
  }

  const Icon = feature.icon;
  return (
    <main>
      <section className="mx-auto max-w-[900px] px-6 pb-12 pt-16 sm:pt-20">
        <span className="grid size-14 place-items-center rounded-2xl bg-primary-soft text-primary"><Icon className="size-7" /></span>
        <div className="label-eyebrow mt-7 text-brand">{feature.eyebrow}</div>
        <h1 className="font-display mt-4 max-w-3xl text-[40px] leading-tight sm:text-[52px]">{feature.title}</h1>
        <p className="text-body mt-5 max-w-3xl text-lg leading-relaxed">{feature.intro}</p>
      </section>

      <section className="mx-auto grid max-w-[1080px] gap-5 px-6 md:grid-cols-3">
        {feature.benefits.map((benefit) => (
          <article key={benefit.title} className="rounded-2xl border bg-card p-6">
            <h2 className="font-display text-2xl">{benefit.title}</h2>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">{benefit.body}</p>
          </article>
        ))}
      </section>

      <section className="mx-auto mt-16 max-w-[900px] px-6">
        <div className="rounded-3xl bg-primary-soft px-7 py-9 sm:px-10">
          <h2 className="font-display text-3xl text-primary">Cara memulai</h2>
          <ol className="mt-6 grid gap-4 sm:grid-cols-2">
            {feature.steps.map((step, index) => (
              <li key={step} className="flex gap-3 rounded-xl bg-card p-4 text-sm leading-relaxed">
                <span className="font-display text-brand">{String(index + 1).padStart(2, "0")}</span>
                <span>{step}</span>
              </li>
            ))}
          </ol>
        </div>
        <div className="mt-10 flex flex-wrap items-center gap-5">
          <Link className="inline-flex h-12 items-center gap-2 rounded-xl bg-primary px-6 font-semibold text-primary-foreground" to="/login">
            Coba txsurvey <ArrowRight className="size-4" />
          </Link>
          <Link className="font-semibold text-primary" to="/contoh-template-survei">Lihat contoh template</Link>
        </div>
      </section>
    </main>
  );
}
