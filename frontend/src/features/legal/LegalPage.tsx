import { useState } from "react";
import { Link } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { BrandMark } from "@/components/BrandMark";
import { cn } from "@/lib/utils";

const UPDATED = "25 Juni 2026";

interface Section {
  title: string;
  body: string[];
  bullets?: string[];
}

const TERMS: Section[] = [
  {
    title: "Penerimaan ketentuan",
    body: [
      "Dengan masuk dan menggunakan txsurvey, kamu menyetujui ketentuan ini. txsurvey adalah alat pembuat survei untuk tim kecil dan saat ini berjalan dalam mode uji terbatas (hanya akun yang diundang).",
    ],
  },
  {
    title: "Penggunaan layanan",
    body: ["Kamu bertanggung jawab atas konten survei yang kamu buat dan jawaban yang kamu kumpulkan."],
    bullets: [
      "Jangan mengumpulkan data pribadi sensitif tanpa dasar yang sah.",
      "Jangan memakai layanan untuk spam, penipuan, atau hal melanggar hukum.",
      "Hormati privasi responden — survei publik bersifat anonim secara default.",
    ],
  },
  {
    title: "Akun & akses",
    body: [
      "Akses diberikan lewat Google Sign-In untuk akun yang diundang. Kamu bertanggung jawab menjaga keamanan akun Google-mu. Kami dapat mencabut akses kapan saja selama masa uji.",
    ],
  },
  {
    title: "Ketersediaan & perubahan",
    body: [
      "Layanan disediakan “sebagaimana adanya” selama masa uji, tanpa jaminan ketersediaan. Fitur dapat berubah, dan ketentuan ini dapat diperbarui — perubahan penting akan kami informasikan.",
    ],
  },
];

const PRIVACY: Section[] = [
  {
    title: "Data yang kami simpan",
    body: ["Kami menyimpan data seminimal mungkin agar layanan berfungsi."],
    bullets: [
      "Identitas pembuat: nama, email, dan foto profil dari Google Sign-In.",
      "Konten survei: pertanyaan, pengaturan, dan logika yang kamu buat.",
      "Jawaban responden: disimpan tanpa identitas pribadi (anonim).",
    ],
  },
  {
    title: "Bagaimana data dipakai",
    body: [
      "Data hanya dipakai untuk menjalankan layanan: menampilkan survei, menyimpan jawaban, dan menghitung ringkasan. Kami tidak menjual data dan tidak memakainya untuk iklan.",
    ],
  },
  {
    title: "Jawaban responden",
    body: [
      "Survei publik tidak meminta email atau identitas responden. Jawaban hanya bisa dilihat oleh pemilik survei dan dapat diekspor sebagai CSV oleh pemilik.",
    ],
  },
  {
    title: "Retensi & penghapusan",
    body: [
      "Menghapus survei akan menghapus jawabannya secara permanen. Kamu dapat meminta penghapusan akun beserta seluruh data terkait lewat formulir di bawah.",
    ],
  },
];

export function LegalPage() {
  const [tab, setTab] = useState<"terms" | "privacy">("terms");
  const sections = tab === "terms" ? TERMS : PRIVACY;

  return (
    <div className="min-h-dvh bg-background">
      <header className="sticky top-0 z-10 border-b bg-card">
        <div className="mx-auto flex h-[60px] max-w-[760px] items-center justify-between px-6">
          <Link to="/login" className="flex items-center gap-2.5">
            <BrandMark size={28} />
            <span className="font-display text-lg text-foreground">txsurvey</span>
          </Link>
          <Button variant="ghost" size="sm" className="text-muted-foreground" asChild>
            <Link to="/login">
              <ArrowLeft /> Kembali ke masuk
            </Link>
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-[760px] px-6 py-10">
        <div className="label-eyebrow text-brand">Dokumen hukum</div>
        <h1 className="font-display mt-3 text-[34px] leading-tight text-foreground">
          {tab === "terms" ? "Ketentuan Layanan" : "Kebijakan Privasi"}
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">Terakhir diperbarui {UPDATED}</p>

        <div className="mt-6 inline-flex rounded-xl border bg-card p-1">
          <SegTab active={tab === "terms"} onClick={() => setTab("terms")}>
            Ketentuan
          </SegTab>
          <SegTab active={tab === "privacy"} onClick={() => setTab("privacy")}>
            Privasi
          </SegTab>
        </div>

        <p className="text-body mt-7 leading-relaxed">
          {tab === "terms"
            ? "Ketentuan singkat ini mengatur penggunaan txsurvey selama masa uji. Ditulis sesederhana mungkin."
            : "Kami percaya survei yang baik dimulai dari rasa percaya. Berikut data yang kami simpan dan cara kami memperlakukannya."}
        </p>

        <div className="mt-8 space-y-9">
          {sections.map((s, i) => (
            <section key={s.title}>
              <div className="flex items-baseline gap-3">
                <span className="font-display text-lg text-brand tabular-nums">{String(i + 1).padStart(2, "0")}</span>
                <h2 className="font-display text-xl text-foreground">{s.title}</h2>
              </div>
              <div className="mt-2 space-y-2 pl-9">
                {s.body.map((p, j) => (
                  <p key={j} className="text-body leading-relaxed">
                    {p}
                  </p>
                ))}
                {s.bullets && (
                  <ul className="mt-2 space-y-1.5">
                    {s.bullets.map((b, j) => (
                      <li key={j} className="flex gap-2.5 text-body">
                        <span className="mt-2 size-1.5 shrink-0 rounded-full bg-brand" />
                        <span className="leading-relaxed">{b}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </section>
          ))}
        </div>

        <ContactCard />

        <p className="mt-8 text-xs leading-relaxed text-muted-foreground">
          Dokumen ini adalah contoh untuk produk masa uji dan bukan nasihat hukum.
        </p>
      </main>
    </div>
  );
}

function SegTab({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "rounded-lg px-4 py-1.5 text-sm font-medium transition-colors",
        active ? "bg-primary-soft text-primary" : "text-muted-foreground hover:text-foreground",
      )}
    >
      {children}
    </button>
  );
}

/** ContactCard collects a question about the legal docs. By design it NEVER asks
 *  for an email; it is a self-contained acknowledgement (no message is stored). */
function ContactCard() {
  const [name, setName] = useState("");
  const [msg, setMsg] = useState("");
  const [sent, setSent] = useState(false);

  return (
    <div className="mt-12 rounded-2xl border bg-card p-6">
      <h3 className="font-display text-lg text-foreground">Ada pertanyaan soal dokumen ini?</h3>
      {sent ? (
        <div className="mt-3">
          <p className="text-body">Terima kasih, pertanyaanmu terkirim.</p>
          <button
            onClick={() => {
              setSent(false);
              setName("");
              setMsg("");
            }}
            className="mt-3 text-sm font-medium text-primary hover:underline"
          >
            Kirim lagi
          </button>
        </div>
      ) : (
        <form
          className="mt-4 space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            if (msg.trim()) setSent(true);
          }}
        >
          <div className="space-y-1.5">
            <Label htmlFor="contact-name">Nama (opsional)</Label>
            <Input id="contact-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="Namamu" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="contact-msg">Pertanyaan</Label>
            <Textarea
              id="contact-msg"
              value={msg}
              onChange={(e) => setMsg(e.target.value)}
              placeholder="Tulis pertanyaanmu…"
              required
            />
          </div>
          <Button type="submit" className="rounded-xl" disabled={!msg.trim()}>
            Kirim pertanyaan
          </Button>
        </form>
      )}
    </div>
  );
}
