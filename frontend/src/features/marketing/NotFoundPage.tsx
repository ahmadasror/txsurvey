import { Link } from "react-router-dom";
import { ArrowRight } from "lucide-react";
import { usePageMetadata } from "@/lib/usePageMetadata";

export function NotFoundPage() {
  usePageMetadata({
    title: "Halaman Tidak Ditemukan",
    description: "Halaman yang kamu cari tidak tersedia.",
    robots: "noindex, nofollow",
  });

  return (
    <main className="flex min-h-dvh flex-col items-center justify-center bg-background px-6 text-center">
      <div className="label-eyebrow text-brand">404</div>
      <h1 className="font-display mt-4 text-4xl">Halaman tidak ditemukan.</h1>
      <p className="mt-3 text-muted-foreground">Tautannya mungkin keliru atau halaman sudah dipindahkan.</p>
      <Link className="mt-7 inline-flex items-center gap-2 font-semibold text-primary" to="/">
        Kembali ke txsurvey <ArrowRight className="size-4" />
      </Link>
    </main>
  );
}
