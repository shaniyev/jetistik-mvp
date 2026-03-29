import { useLanguage } from "@/i18n/LanguageContext";

export function Footer() {
  const { t } = useLanguage();

  return (
    <footer className="border-t bg-background py-10">
      <div className="container flex flex-col items-center gap-4 text-center text-sm text-muted-foreground md:flex-row md:justify-between md:text-left">
        <p>{t("footer.tagline")}</p>
        <div className="flex gap-6">
          <a href="mailto:info@jetistik.kz" className="hover:text-foreground transition-colors">
            info@jetistik.kz
          </a>
          <a href="/privacy" className="hover:text-foreground transition-colors">
            {t("footer.privacy")}
          </a>
        </div>
      </div>
    </footer>
  );
}
