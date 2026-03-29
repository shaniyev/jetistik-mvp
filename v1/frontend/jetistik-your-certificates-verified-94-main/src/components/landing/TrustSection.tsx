import { Card, CardContent } from "@/components/ui/card";
import { ShieldCheck, Lock, FileText } from "lucide-react";
import { useLanguage } from "@/i18n/LanguageContext";
import { TranslationKey } from "@/i18n/translations";

const items: { icon: typeof ShieldCheck; titleKey: TranslationKey; descKey: TranslationKey }[] = [
  { icon: ShieldCheck, titleKey: "trust.authenticityTitle", descKey: "trust.authenticityDesc" },
  { icon: Lock, titleKey: "trust.securityTitle", descKey: "trust.securityDesc" },
  { icon: FileText, titleKey: "trust.convenienceTitle", descKey: "trust.convenienceDesc" },
];

export function TrustSection() {
  const { t } = useLanguage();

  return (
    <section className="bg-background py-16 md:py-24">
      <div className="container max-w-5xl">
        <div className="grid gap-6 md:grid-cols-3">
          {items.map((item) => (
            <Card
              key={item.titleKey}
              className="border-border/60 bg-card shadow-sm transition-shadow hover:shadow-md"
            >
              <CardContent className="flex flex-col items-center p-8 text-center">
                <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                  <item.icon className="h-6 w-6 text-primary" />
                </div>
                <h3 className="mb-2 text-lg font-semibold text-foreground">{t(item.titleKey)}</h3>
                <p className="text-sm leading-relaxed text-muted-foreground">{t(item.descKey)}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
