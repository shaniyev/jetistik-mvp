import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CheckCircle, Loader2, ArrowLeft } from "lucide-react";
import { Link } from "react-router-dom";
import { useLanguage } from "@/i18n/LanguageContext";

interface FormData {
  fullName: string;
  organization: string;
  phone: string;
}

interface FormErrors {
  fullName?: string;
  organization?: string;
  phone?: string;
}

export default function AdminRequest() {
  const [form, setForm] = useState<FormData>({ fullName: "", organization: "", phone: "" });
  const [errors, setErrors] = useState<FormErrors>({});
  const [status, setStatus] = useState<"idle" | "loading" | "success">("idle");
  const { t } = useLanguage();

  const validate = (): boolean => {
    const e: FormErrors = {};
    const name = form.fullName.trim();
    const org = form.organization.trim();
    const phone = form.phone.trim();

    if (!name) e.fullName = t("admin.errorName");
    else if (name.length < 3) e.fullName = t("admin.errorNameMin");
    else if (name.length > 100) e.fullName = t("admin.errorNameMax");

    if (!org) e.organization = t("admin.errorOrg");
    else if (org.length > 150) e.organization = t("admin.errorOrgMax");

    if (!phone) e.phone = t("admin.errorPhone");
    else if (!/^\+?\d[\d\s\-()]{6,18}\d$/.test(phone)) e.phone = t("admin.errorPhoneInvalid");

    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    setStatus("loading");
    await new Promise((r) => setTimeout(r, 1000));
    setStatus("success");
  };

  const handleChange = (field: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [field]: e.target.value }));
    if (errors[field]) setErrors((prev) => ({ ...prev, [field]: undefined }));
  };

  if (status === "success") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
        <Card className="w-full max-w-md text-center">
          <CardContent className="flex flex-col items-center gap-4 p-10">
            <div className="flex h-14 w-14 items-center justify-center rounded-full bg-success/10">
              <CheckCircle className="h-8 w-8 text-success" />
            </div>
            <h2 className="text-xl font-semibold text-foreground">{t("admin.successTitle")}</h2>
            <p className="text-sm text-muted-foreground">{t("admin.successDesc")}</p>
            <Button variant="outline" asChild className="mt-2">
              <Link to="/">{t("admin.home")}</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold text-foreground">{t("admin.title")}</CardTitle>
          <CardDescription>{t("admin.subtitle")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            <div className="space-y-1.5">
              <Label htmlFor="fullName">{t("admin.fullName")}</Label>
              <Input
                id="fullName"
                placeholder={t("admin.fullNamePlaceholder")}
                value={form.fullName}
                onChange={handleChange("fullName")}
                maxLength={100}
                disabled={status === "loading"}
              />
              {errors.fullName && <p className="text-xs text-destructive">{errors.fullName}</p>}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="organization">{t("admin.organization")}</Label>
              <Input
                id="organization"
                placeholder={t("admin.organizationPlaceholder")}
                value={form.organization}
                onChange={handleChange("organization")}
                maxLength={150}
                disabled={status === "loading"}
              />
              {errors.organization && <p className="text-xs text-destructive">{errors.organization}</p>}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="phone">{t("admin.phone")}</Label>
              <Input
                id="phone"
                type="tel"
                placeholder={t("admin.phonePlaceholder")}
                value={form.phone}
                onChange={handleChange("phone")}
                maxLength={20}
                disabled={status === "loading"}
              />
              {errors.phone && <p className="text-xs text-destructive">{errors.phone}</p>}
            </div>

            <Button type="submit" className="h-11 w-full gap-2" disabled={status === "loading"}>
              {status === "loading" && <Loader2 className="h-4 w-4 animate-spin" />}
              {t("admin.submit")}
            </Button>
          </form>

          <div className="mt-6 text-center">
            <Link to="/" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors">
              <ArrowLeft className="h-3.5 w-3.5" />
              {t("admin.home")}
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
