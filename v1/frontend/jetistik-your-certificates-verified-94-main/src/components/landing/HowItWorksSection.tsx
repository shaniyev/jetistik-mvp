import { KeyRound, List, Download, QrCode } from "lucide-react";
import { useLanguage } from "@/i18n/LanguageContext";
import { TranslationKey } from "@/i18n/translations";

const steps: { icon: typeof KeyRound; titleKey: TranslationKey; descKey: TranslationKey }[] = [
  { icon: KeyRound, titleKey: "how.step1Title", descKey: "how.step1Desc" },
  { icon: List, titleKey: "how.step2Title", descKey: "how.step2Desc" },
  { icon: Download, titleKey: "how.step3Title", descKey: "how.step3Desc" },
  { icon: QrCode, titleKey: "how.step4Title", descKey: "how.step4Desc" },
];

export function HowItWorksSection() {
  const { t } = useLanguage();

  return (
    <section id="how-it-works" className="bg-muted/50 py-16 md:py-24">
      <div className="container max-w-5xl">
        <h2 className="mb-12 text-center text-2xl font-bold text-foreground md:text-3xl">
          {t("how.title")}
        </h2>

        <div className="grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
          {steps.map((step, i) => (
            <div key={step.titleKey} className="flex flex-col items-center text-center">
              <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-xl bg-primary/10">
                <step.icon className="h-7 w-7 text-primary" />
              </div>
              <span className="mb-1 text-xs font-semibold uppercase tracking-wider text-primary">
                {t("how.step")} {i + 1}
              </span>
              <h3 className="mb-1 text-base font-semibold text-foreground">{t(step.titleKey)}</h3>
              <p className="text-sm text-muted-foreground">{t(step.descKey)}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
