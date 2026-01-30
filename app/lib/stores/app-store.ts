import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

interface AppState {
  // Theme
  theme: "light" | "dark" | "system";
  setTheme: (theme: "light" | "dark" | "system") => void;

  // Locale
  locale: string;
  setLocale: (locale: string) => void;

  // Counter (Zustand demo)
  count: number;
  increment: () => void;
  decrement: () => void;
  reset: () => void;

  // Ping history
  pingHistory: Array<{ message: string; timestamp: Date }>;
  addPing: (message: string) => void;
  clearPingHistory: () => void;
}

export const useAppStore = create<AppState>()(
  devtools(
    persist(
      (set) => ({
        // Theme
        theme: "system",
        setTheme: (theme) => set({ theme }),

        // Locale
        locale: "en",
        setLocale: (locale) => set({ locale }),

        // Counter (Zustand demo)
        count: 0,
        increment: () => set((state) => ({ count: state.count + 1 })),
        decrement: () => set((state) => ({ count: state.count - 1 })),
        reset: () => set({ count: 0 }),

        // Ping history
        pingHistory: [],
        addPing: (message) =>
          set((state) => ({
            pingHistory: [
              { message, timestamp: new Date() },
              ...state.pingHistory.slice(0, 9), // Keep last 10
            ],
          })),
        clearPingHistory: () => set({ pingHistory: [] }),
      }),
      {
        name: "chattycathy-storage",
      }
    )
  )
);
