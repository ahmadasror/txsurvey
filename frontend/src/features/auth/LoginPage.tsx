import { useState } from "react";
import { Link, Navigate } from "react-router-dom";
import { Check, Loader2 } from "lucide-react";
import { BrandMark } from "@/components/BrandMark";
import { FullScreenLoader } from "@/components/FullScreenLoader";
import { loginUrl } from "@/api/client";
import { useMe } from "@/api/auth";
import { useDocumentTitle } from "@/lib/useDocumentTitle";

const VALUE_PROPS = [
  "Satu pertanyaan per layar — terasa seperti ngobrol.",
  "Logika lompat & cabang, bukan formulir kaku.",
  "Lima tema hangat, tipografi editorial.",
  "Ekspor CSV, analitik jujur, tanpa metrik palsu.",
];

export function LoginPage() {
  useDocumentTitle("Masuk");
  const { data: user, isLoading } = useMe();
  const [connecting, setConnecting] = useState(false);

  if (isLoading) return <FullScreenLoader />;
  if (user) return <Navigate to="/" replace />;

  const connect = () => {
    setConnecting(true);
    window.location.href = loginUrl();
  };

  return (
    <div className="flex min-h-dvh flex-wrap bg-background">
      {/* Brand panel */}
      <section className="relative flex flex-[1.1_1_420px] flex-col justify-between overflow-hidden bg-primary px-8 py-10 text-primary-foreground sm:px-12">
        {/* Floating geometry for depth */}
        <span className="animate-floaty pointer-events-none absolute -right-10 top-16 size-40 rounded-full bg-white/5" />
        <span
          className="animate-floaty pointer-events-none absolute right-24 top-44 size-16 rounded-3xl bg-brand/30"
          style={{ animationDelay: "1.5s" }}
        />
        <span className="pointer-events-none absolute -bottom-16 -left-12 size-56 rounded-full bg-black/10" />

        <div className="relative z-10 flex items-center gap-2.5">
          <BrandMark size={34} />
          <span className="font-display text-xl">txsurvey</span>
        </div>

        <div className="relative z-10 max-w-md">
          <div className="label-eyebrow text-brand">Survei yang diisi sampai habis</div>
          <h1 className="font-display mt-4 text-[40px] leading-[1.05] sm:text-[46px]">
            Bikin survei yang benar-benar diisi orang.
          </h1>
          <p className="mt-5 text-[15.5px] leading-relaxed text-primary-foreground/80">
            Bukan formulir membosankan. Satu pertanyaan per layar, transisi halus, dan logika cabang yang
            membuat tiap responden merasa diperhatikan.
          </p>
          <ul className="mt-7 space-y-3.5">
            {VALUE_PROPS.map((v) => (
              <li key={v} className="flex items-start gap-3">
                <span className="mt-0.5 grid size-5 shrink-0 place-items-center rounded bg-white/16">
                  <Check className="size-3.5" />
                </span>
                <span className="text-[15.5px] text-primary-foreground/90">{v}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="relative z-10 flex items-center gap-3 text-sm text-primary-foreground/75">
          <div className="flex -space-x-2">
            {["#D98E5A", "#E7EFE9", "#6B8F71"].map((c, i) => (
              <span key={i} className="size-7 rounded-full border-2 border-primary" style={{ background: c }} />
            ))}
          </div>
          Dipakai tim kecil untuk feedback yang jujur.
        </div>
      </section>

      {/* Sign-in card */}
      <section className="flex flex-[0.9_1_380px] items-center justify-center px-6 py-10">
        <div className="w-full max-w-[380px]">
          <BrandMark size={52} className="mb-6" />
          <h2 className="font-display text-[28px] leading-tight text-foreground">Masuk ke txsurvey</h2>
          <p className="mt-2 text-[15px] text-muted-foreground">
            Masuk atau daftar dengan akun Google — gratis.
          </p>

          <button
            onClick={connect}
            disabled={connecting}
            className="mt-7 flex h-[52px] w-full items-center justify-center gap-3 rounded-xl border bg-card text-[15px] font-medium text-foreground transition-all hover:-translate-y-0.5 hover:border-foreground/25 hover:shadow-[0_12px_26px_rgba(29,38,36,0.10)] disabled:opacity-70"
          >
            {connecting ? (
              <>
                <Loader2 className="size-5 animate-spin" /> Menghubungkan…
              </>
            ) : (
              <>
                <GoogleG /> Lanjut dengan Google
              </>
            )}
          </button>

          <div className="my-6 flex items-center gap-3 text-xs text-muted-foreground">
            <span className="h-px flex-1 bg-border" />
            gratis & cepat
            <span className="h-px flex-1 bg-border" />
          </div>

          <div className="rounded-xl bg-primary-soft px-4 py-3 text-[13.5px] text-primary">
            Gratis untuk tim kecil. Kuota pendaftaran terbatas untuk tahap ini.
          </div>

          <p className="mt-6 text-xs leading-relaxed text-muted-foreground">
            Dengan masuk, kamu menyetujui{" "}
            <Link to="/legal" className="font-medium text-primary underline-offset-2 hover:underline">
              Ketentuan Layanan
            </Link>{" "}
            dan{" "}
            <Link to="/legal" className="font-medium text-primary underline-offset-2 hover:underline">
              Kebijakan Privasi
            </Link>
            .
          </p>
        </div>
      </section>
    </div>
  );
}

/** Google's 4-color "G" — keep branding intact on the sign-in button. */
function GoogleG() {
  return (
    <svg className="size-5" viewBox="0 0 48 48" aria-hidden>
      <path
        fill="#EA4335"
        d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"
      />
      <path
        fill="#4285F4"
        d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"
      />
      <path
        fill="#FBBC05"
        d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"
      />
      <path
        fill="#34A853"
        d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"
      />
    </svg>
  );
}
