import { Link, Outlet } from "react-router-dom";
import { ArrowRight } from "lucide-react";
import { BrandMark } from "@/components/BrandMark";

export function PublicLayout() {
  return (
    <div className="min-h-dvh bg-background text-foreground">
      <header className="border-b bg-card/95 backdrop-blur">
        <div className="mx-auto flex min-h-16 max-w-[1080px] items-center justify-between gap-5 px-6">
          <Link to="/" className="flex items-center gap-2.5" aria-label="txsurvey public home">
            <BrandMark size={32} />
            <span className="font-display text-xl">txsurvey</span>
          </Link>
          <nav className="hidden items-center gap-6 text-sm text-muted-foreground md:flex" aria-label="Navigasi publik">
            <Link className="hover:text-foreground" to="/contoh-template-survei">Template</Link>
            <Link className="hover:text-foreground" to="/fitur/logika-bercabang">Logika bercabang</Link>
            <Link className="hover:text-foreground" to="/fitur/survei-anonim">Survei anonim</Link>
            <Link className="hover:text-foreground" to="/panduan">Panduan</Link>
          </nav>
          <Link
            to="/login"
            className="inline-flex h-10 items-center gap-2 rounded-xl bg-primary px-4 text-sm font-semibold text-primary-foreground"
          >
            Buat survei <ArrowRight className="size-4" />
          </Link>
        </div>
      </header>

      <Outlet />

      <footer className="mt-16 border-t bg-card">
        <div className="mx-auto flex max-w-[1080px] flex-col gap-5 px-6 py-9 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <BrandMark size={26} />
            <span>txsurvey — survei yang terasa seperti ngobrol.</span>
          </div>
          <div className="flex flex-wrap gap-5">
            <Link className="hover:text-foreground" to="/contoh-template-survei">Template</Link>
            <Link className="hover:text-foreground" to="/panduan">FAQ</Link>
            <Link className="hover:text-foreground" to="/legal">Privasi & ketentuan</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
