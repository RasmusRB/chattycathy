import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

export const locales = ["en", "es", "fr"] as const;
export type Locale = (typeof locales)[number];

export const localeNames: Record<Locale, string> = {
  en: "English",
  es: "Español",
  fr: "Français",
};

export const localeFlags: Record<Locale, string> = {
  en: "🇺🇸",
  es: "🇪🇸",
  fr: "🇫🇷",
};

interface LocaleState {
  locale: Locale;
  setLocale: (locale: Locale) => void;
}

export const useLocaleStore = create<LocaleState>()(
  devtools(
    persist(
      (set) => ({
        locale: "en",
        setLocale: (locale) => {
          // Set cookie for server-side rendering
          document.cookie = `locale=${locale};path=/;max-age=31536000;samesite=lax`;
          set({ locale });
          // Reload the page to apply the new locale
          window.location.reload();
        },
      }),
      {
        name: "locale-storage",
      }
    ),
    { name: "locale" }
  )
);
