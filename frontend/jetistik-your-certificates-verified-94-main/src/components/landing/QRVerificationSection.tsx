import { CheckCircle, XCircle, HelpCircle } from "lucide-react";
import { useLanguage } from "@/i18n/LanguageContext";
import { TranslationKey } from "@/i18n/translations";

const statuses: {
  icon: typeof CheckCircle;
  label: string;
  descKey: TranslationKey;
  color: string;
  bg: string;
}[] = [
  { icon: CheckCircle, label: "VALID", descKey: "qr.validDesc", color: "text-success", bg: "bg-success/10" },
  { icon: XCircle, label: "REVOKED", descKey: "qr.revokedDesc", color: "text-destructive", bg: "bg-destructive/10" },
  { icon: HelpCircle, label: "NOT_FOUND", descKey: "qr.notFoundDesc", color: "text-neutral", bg: "bg-neutral/10" },
];

export function QRVerificationSection() {
  const { t } = useLanguage();

  return (
    <section id="verification" className="bg-background py-16 md:py-24">
      <div className="container max-w-3xl text-center">
        <h2 className="mb-4 text-2xl font-bold text-foreground md:text-3xl">
          {t("qr.title")}
        </h2>
        <p className="mx-auto mb-10 max-w-lg text-muted-foreground">
          {t("qr.desc")}
        </p>

        <div className="grid gap-4 sm:grid-cols-3">
          {statuses.map((s) => (
            <div
              key={s.label}
              className={`flex flex-col items-center gap-2 rounded-lg border p-6 ${s.bg}`}
            >
              <s.icon className={`h-8 w-8 ${s.color}`} />
              <span className={`text-sm font-bold ${s.color}`}>{s.label}</span>
              <span className="text-xs text-muted-foreground">{t(s.descKey)}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
