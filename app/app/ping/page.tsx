"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { usePingMutation } from "@/lib/hooks/use-ping";
import { useAppStore } from "@/lib/stores/app-store";
import { useAuthStore } from "@/lib/stores/auth-store";
import { Button } from "@/components/ui/button";
import { Navbar } from "@/components/navbar";

export default function PingPage() {
  const t = useTranslations();
  const router = useRouter();
  const { mutate: ping, data, isPending, error } = usePingMutation();
  const pingHistory = useAppStore((state) => state.pingHistory);
  const clearPingHistory = useAppStore((state) => state.clearPingHistory);

  // Auth state
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  // Zustand counter demo
  const count = useAppStore((state) => state.count);
  const increment = useAppStore((state) => state.increment);
  const decrement = useAppStore((state) => state.decrement);
  const reset = useAppStore((state) => state.reset);

  // Redirect if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      router.push("/login");
    }
  }, [isAuthenticated, router]);

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-zinc-950">
      {/* Header */}
      <Navbar showAuth />

      {/* Main content */}
      <main className="flex flex-1 flex-col items-center gap-8 px-8 py-16">
        <div className="w-full max-w-2xl">
          <div className="text-center">
            <h2 className="text-3xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
              {t("ping.title")}
            </h2>
            <p className="mt-2 text-lg text-zinc-600 dark:text-zinc-400">
              {t("ping.description")}
            </p>
          </div>

          <div className="mt-8 flex flex-col items-center gap-4">
            <Button
              onClick={() => ping()}
              disabled={isPending}
              size="lg"
              className="min-w-[150px]"
            >
              {isPending ? t("common.loading") : t("ping.button")}
            </Button>

            {error && (
              <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-400">
                {t("common.error")}: {error.message}
              </div>
            )}

            {data && (
              <div className="rounded-lg border border-green-200 bg-green-50 p-6 dark:border-green-800 dark:bg-green-950">
                <p className="text-sm font-medium text-green-700 dark:text-green-400">
                  {t("ping.response")}
                </p>
                <p className="mt-1 text-lg font-semibold text-green-800 dark:text-green-300">
                  {data.message}
                </p>
              </div>
            )}
          </div>

          {/* Counter Section */}
          <div className="mt-12 rounded-xl border border-zinc-200 bg-white p-8 dark:border-zinc-800 dark:bg-zinc-900">
            <h3 className="text-center text-xl font-semibold text-zinc-900 dark:text-zinc-50">
              {t("counter.title")}
            </h3>
            <p className="mt-1 text-center text-sm text-zinc-500 dark:text-zinc-400">
              {t("counter.description")}
            </p>

            <div className="mt-6 flex flex-col items-center gap-6">
              <div className="text-6xl font-bold tabular-nums text-zinc-900 dark:text-zinc-50">
                {count}
              </div>

              <div className="flex gap-3">
                <Button
                  onClick={decrement}
                  variant="outline"
                  size="lg"
                  className="text-lg font-bold min-w-[60px]"
                >
                  −
                </Button>
                <Button
                  onClick={reset}
                  variant="outline"
                  size="lg"
                  className="min-w-[80px]"
                >
                  {t("counter.reset")}
                </Button>
                <Button
                  onClick={increment}
                  variant="outline"
                  size="lg"
                  className="text-lg font-bold min-w-[60px]"
                >
                  +
                </Button>
              </div>
            </div>
          </div>

          {/* Ping History */}
          {pingHistory.length > 0 && (
            <div className="mt-8 rounded-xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-900">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-zinc-900 dark:text-zinc-50">
                  {t("ping.history")}
                </h3>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearPingHistory}
                  className="text-xs"
                >
                  {t("common.clear")}
                </Button>
              </div>
              <ul className="mt-4 space-y-2">
                {pingHistory.map((item, index) => (
                  <li
                    key={index}
                    className="flex items-center justify-between rounded-lg bg-zinc-50 px-4 py-2 text-sm dark:bg-zinc-800"
                  >
                    <span className="text-zinc-700 dark:text-zinc-300">
                      {item.message}
                    </span>
                    <span className="text-xs text-zinc-400">
                      {new Date(item.timestamp).toLocaleTimeString()}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
