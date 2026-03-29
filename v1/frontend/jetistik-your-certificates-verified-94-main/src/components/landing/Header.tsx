import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Menu, X } from "lucide-react";
import { useLanguage } from "@/i18n/LanguageContext";

export function Header() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const { lang, setLang, t } = useLanguage();

  const navLinks = [
    { label: t("nav.howItWorks"), href: "#how-it-works" },
    { label: t("nav.verification"), href: "#verification" },
    { label: t("nav.faq"), href: "#faq" },
  ];

  return (
    <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80">
      <div className="container flex h-16 items-center justify-between">
        <a href="/" className="text-xl font-bold tracking-tight text-primary">
          Jetistik
        </a>

        {/* Desktop nav */}
        <nav className="hidden items-center gap-8 md:flex">
          {navLinks.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
            >
              {l.label}
            </a>
          ))}
          <Button variant="outline" size="sm" asChild>
            <a href="/admin/login">{t("nav.forOrganizers")}</a>
          </Button>

          {/* Language switcher */}
          <div className="flex items-center rounded-md border border-border text-sm">
            <button
              onClick={() => setLang("kz")}
              className={`px-2.5 py-1 rounded-l-md transition-colors ${lang === "kz" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground"}`}
            >
              ҚАЗ
            </button>
            <button
              onClick={() => setLang("ru")}
              className={`px-2.5 py-1 rounded-r-md transition-colors ${lang === "ru" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground"}`}
            >
              РУС
            </button>
          </div>
        </nav>

        {/* Mobile toggle */}
        <div className="flex items-center gap-3 md:hidden">
          <div className="flex items-center rounded-md border border-border text-xs">
            <button
              onClick={() => setLang("kz")}
              className={`px-2 py-1 rounded-l-md transition-colors ${lang === "kz" ? "bg-primary text-primary-foreground" : "text-muted-foreground"}`}
            >
              ҚАЗ
            </button>
            <button
              onClick={() => setLang("ru")}
              className={`px-2 py-1 rounded-r-md transition-colors ${lang === "ru" ? "bg-primary text-primary-foreground" : "text-muted-foreground"}`}
            >
              РУС
            </button>
          </div>
          <button onClick={() => setMobileOpen(!mobileOpen)} aria-label={t("nav.menu")}>
            {mobileOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
          </button>
        </div>
      </div>

      {/* Mobile nav */}
      {mobileOpen && (
        <nav className="flex flex-col gap-4 border-t bg-background px-6 py-4 md:hidden">
          {navLinks.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="text-sm font-medium text-muted-foreground hover:text-foreground"
              onClick={() => setMobileOpen(false)}
            >
              {l.label}
            </a>
          ))}
          <Button variant="outline" size="sm" className="w-fit" asChild>
            <a href="/admin/login">{t("nav.forOrganizers")}</a>
          </Button>
        </nav>
      )}
    </header>
  );
}
