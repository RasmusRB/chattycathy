import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

export type Theme = "light" | "dark" | "system";

interface ThemeState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

function applyTheme(theme: Theme) {
  if (typeof window === "undefined") return;

  const root = document.documentElement;
  const isDark =
    theme === "dark" ||
    (theme === "system" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches);

  root.classList.remove("light", "dark");
  root.classList.add(isDark ? "dark" : "light");
}

export const useThemeStore = create<ThemeState>()(
  devtools(
    persist(
      (set) => ({
        theme: "system",
        setTheme: (theme) => {
          applyTheme(theme);
          set({ theme });
        },
      }),
      {
        name: "theme-storage",
        onRehydrateStorage: () => (state) => {
          // Apply theme after hydration
          if (state) {
            applyTheme(state.theme);
          }
        },
      }
    ),
    { name: "theme" }
  )
);

// Initialize theme on client side
if (typeof window !== "undefined") {
  // Listen for system theme changes
  const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
  mediaQuery.addEventListener("change", () => {
    const state = useThemeStore.getState();
    if (state.theme === "system") {
      applyTheme("system");
    }
  });

  // Apply initial theme (handles SSR mismatch)
  const stored = localStorage.getItem("theme-storage");
  if (stored) {
    try {
      const { state } = JSON.parse(stored);
      if (state?.theme) {
        applyTheme(state.theme);
      }
    } catch {
      applyTheme("system");
    }
  }
}
