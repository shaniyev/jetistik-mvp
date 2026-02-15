import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Search, Loader2, AlertCircle } from "lucide-react";
import { useLanguage } from "@/i18n/LanguageContext";

type Status = "idle" | "loading" | "error" | "empty" | "rate-limit";

export function HeroSection() {
  const [iin, setIin] = useState("");
  const [status, setStatus] = useState<Status>("idle");
  const [errorMsg, setErrorMsg] = useState("");
  const navigate = useNavigate();
  const { t } = useLanguage();

  const validate = (v: string) => /^\d{12}$/.test(v);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate(iin)) {
      setStatus("error");
      setErrorMsg(t("hero.errorIin"));
      return;
    }

    setStatus("loading");
    setErrorMsg("");

    try {
      await new Promise((r) => setTimeout(r, 1200));
      navigate(`/my?iin=${iin}`);
    } catch {
      setStatus("error");
      setErrorMsg(t("hero.errorGeneric"));
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const v = e.target.value.replace(/\D/g, "").slice(0, 12);
    setIin(v);
    if (status !== "idle" && status !== "loading") setStatus("idle");
  };

  return (
    <section className="relative overflow-hidden bg-gradient-to-b from-primary/5 to-background py-20 md:py-32">
      <div className="container max-w-3xl text-center">
        <h1 className="mb-4 text-3xl font-bold leading-tight tracking-tight text-foreground md:text-5xl">
          {t("hero.title")}
        </h1>
        <p className="mx-auto mb-10 max-w-xl text-base text-muted-foreground md:text-lg">
          {t("hero.subtitle")}
        </p>

        <form
          onSubmit={handleSubmit}
          className="mx-auto flex max-w-md flex-col items-center gap-3 sm:flex-row"
        >
          <div className="relative w-full">
            <Input
              value={iin}
              onChange={handleChange}
              placeholder={t("hero.inputPlaceholder")}
              inputMode="numeric"
              maxLength={12}
              className="h-12 pr-4 text-base"
              disabled={status === "loading"}
            />
          </div>
          <Button
            type="submit"
            size="lg"
            className="h-12 w-full gap-2 sm:w-auto"
            disabled={status === "loading"}
          >
            {status === "loading" ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Search className="h-4 w-4" />
            )}
            {t("hero.submit")}
          </Button>
        </form>

        {status === "error" && errorMsg && (
          <p className="mt-3 flex items-center justify-center gap-1.5 text-sm text-destructive">
            <AlertCircle className="h-4 w-4" />
            {errorMsg}
          </p>
        )}
        {status === "rate-limit" && (
          <p className="mt-3 text-sm text-destructive">{t("hero.rateLimit")}</p>
        )}
        {status === "empty" && (
          <p className="mt-3 text-sm text-muted-foreground">{t("hero.empty")}</p>
        )}

        <p className="mt-4 text-xs text-muted-foreground">{t("hero.privacy")}</p>
      </div>
    </section>
  );
}
