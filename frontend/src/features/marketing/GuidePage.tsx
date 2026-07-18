import { Link } from "react-router-dom";
import { ArrowRight, BookOpen } from "lucide-react";
import { usePageMetadata } from "@/lib/usePageMetadata";

const STEPS = [
  { title: "Pilih tujuan", body: "Tentukan keputusan apa yang ingin kamu ambil dari hasil survei, bukan sekadar daftar pertanyaan." },
  { title: "Susun pertanyaan", body: "Mulai dari yang mudah, gunakan bahasa singkat, lalu tandai hanya pertanyaan penting sebagai wajib." },
  { title: "Atur alur", body: "Tambahkan logika bercabang ketika kelompok responden membutuhkan pertanyaan lanjutan yang berbeda." },
  { title: "Bagikan dan baca hasil", body: "Publish, kirim tautan publik, lalu pantau respons, ringkasan, dan funnel penyelesaian." },
];

const FAQ = [
  { question: "Apakah responden harus memiliki akun?", answer: "Tidak. Survei yang sudah dipublikasikan dapat diisi langsung melalui tautan publik tanpa login." },
  { question: "Apakah survei bersifat anonim?", answer: "Secara default txsurvey tidak meminta identitas responden. Hindari menambahkan pertanyaan nama atau email jika anonimitas memang dijanjikan." },
  { question: "Bisakah pertanyaan dilewati berdasarkan jawaban?", answer: "Bisa. Gunakan logika bercabang untuk melompat ke pertanyaan yang posisinya lebih akhir berdasarkan kondisi jawaban." },
  { question: "Bagaimana cara mengekspor jawaban?", answer: "Pemilik survei dapat membuka halaman hasil dan mengunduh seluruh jawaban sebagai CSV." },
  { question: "Apakah tautan survei bisa berubah?", answer: "Slug dapat disesuaikan selama survei masih draft. Setelah dipublikasikan, slug dikunci agar tautan yang sudah dibagikan tidak rusak." },
  { question: "Siapa yang bisa melihat hasil?", answer: "Hasil dan analitik hanya tersedia di area creator yang dilindungi session login." },
];

export function GuidePage() {
  usePageMetadata({
    title: "Panduan Membuat Survei Online",
    description: "Panduan singkat membuat, membagikan, dan membaca hasil survei online dengan pertanyaan relevan dan privasi yang jelas.",
    robots: "index, follow",
    path: "panduan",
  });

  return (
    <main>
      <section className="mx-auto max-w-[900px] px-6 pb-12 pt-16 sm:pt-20">
        <span className="grid size-14 place-items-center rounded-2xl bg-primary-soft text-primary"><BookOpen className="size-7" /></span>
        <div className="label-eyebrow mt-7 text-brand">Panduan txsurvey</div>
        <h1 className="font-display mt-4 max-w-3xl text-[40px] leading-tight sm:text-[52px]">Buat survei singkat yang menghasilkan jawaban berguna.</h1>
        <p className="text-body mt-5 max-w-3xl text-lg leading-relaxed">
          Alur sederhana dari menentukan tujuan sampai membaca hasil—cukup untuk survei internal, feedback pelanggan, dan evaluasi layanan.
        </p>
      </section>

      <section className="mx-auto grid max-w-[1080px] gap-5 px-6 sm:grid-cols-2">
        {STEPS.map((step, index) => (
          <article key={step.title} className="rounded-2xl border bg-card p-6">
            <span className="font-display text-lg text-brand">{String(index + 1).padStart(2, "0")}</span>
            <h2 className="font-display mt-3 text-2xl">{step.title}</h2>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{step.body}</p>
          </article>
        ))}
      </section>

      <section className="mx-auto mt-16 max-w-[900px] px-6">
        <h2 className="font-display text-3xl">Pertanyaan umum</h2>
        <div className="mt-6 divide-y rounded-2xl border bg-card px-6">
          {FAQ.map((item) => (
            <details key={item.question} className="group py-5">
              <summary className="cursor-pointer list-none pr-7 font-semibold marker:hidden">{item.question}</summary>
              <p className="mt-3 max-w-3xl text-sm leading-relaxed text-muted-foreground">{item.answer}</p>
            </details>
          ))}
        </div>
        <div className="mt-10 flex flex-wrap gap-5">
          <Link className="inline-flex h-12 items-center gap-2 rounded-xl bg-primary px-6 font-semibold text-primary-foreground" to="/login">
            Buat survei gratis <ArrowRight className="size-4" />
          </Link>
          <Link className="inline-flex items-center font-semibold text-primary" to="/contoh-template-survei">Lihat template siap pakai</Link>
        </div>
      </section>
    </main>
  );
}
