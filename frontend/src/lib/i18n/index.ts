import { writable, derived } from "svelte/store";
import { browser } from "$app/environment";
import kz from "./kz";
import ru from "./ru";
import en from "./en";

export type Language = "kz" | "ru" | "en";
export type TranslationKey = keyof typeof en;

const translations: Record<Language, Record<string, string>> = { kz, ru, en };

function getInitialLanguage(): Language {
  if (browser) {
    const stored = localStorage.getItem("jetistik-lang");
    if (stored && stored in translations) return stored as Language;
  }
  return "kz";
}

export const language = writable<Language>(getInitialLanguage());

if (browser) {
  language.subscribe((lang) => {
    localStorage.setItem("jetistik-lang", lang);
    document.documentElement.lang = lang === "kz" ? "kk" : lang;
  });
}

export const t = derived(language, ($lang) => {
  return (key: TranslationKey): string => {
    return translations[$lang]?.[key] ?? translations.kz[key] ?? key;
  };
});

export function setLanguage(lang: Language) {
  language.set(lang);
}
