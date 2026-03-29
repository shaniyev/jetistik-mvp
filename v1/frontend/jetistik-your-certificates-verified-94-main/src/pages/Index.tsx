import { Header } from "@/components/landing/Header";
import { HeroSection } from "@/components/landing/HeroSection";
import { TrustSection } from "@/components/landing/TrustSection";
import { HowItWorksSection } from "@/components/landing/HowItWorksSection";
import { QRVerificationSection } from "@/components/landing/QRVerificationSection";
import { FAQSection } from "@/components/landing/FAQSection";
import { Footer } from "@/components/landing/Footer";
import { useEffect } from "react";
import { useLanguage } from "@/i18n/LanguageContext";

const Index = () => {
  const { t } = useLanguage();

  useEffect(() => {
    document.title = t("seo.title");
    const meta = document.querySelector('meta[name="description"]');
    if (meta) {
      meta.setAttribute("content", t("seo.description"));
    } else {
      const tag = document.createElement("meta");
      tag.name = "description";
      tag.content = t("seo.description");
      document.head.appendChild(tag);
    }
  }, [t]);

  return (
    <div className="flex min-h-screen flex-col">
      <Header />
      <main className="flex-1">
        <HeroSection />
        <TrustSection />
        <HowItWorksSection />
        <QRVerificationSection />
        <FAQSection />
      </main>
      <Footer />
    </div>
  );
};

export default Index;
