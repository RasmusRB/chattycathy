"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useGoogleAuth } from "@/lib/hooks/use-google-auth";
import { useAuthStore } from "@/lib/stores/auth-store";
import { Navbar } from "@/components/navbar";

export default function LoginPage() {
  const t = useTranslations();
  const router = useRouter();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const { isLoading, error, isReady, renderGoogleButton } = useGoogleAuth({
    onSuccess: () => {
      router.push("/ping");
    },
    onError: (error) => {
      console.error("Login failed:", error);
    },
  });

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      router.push("/ping");
    }
  }, [isAuthenticated, router]);

  // Render Google button when ready
  useEffect(() => {
    if (isReady) {
      renderGoogleButton("google-signin-button");
    }
  }, [isReady, renderGoogleButton]);

  return (
    <div className="flex min-h-screen flex-col bg-zinc-50 font-sans dark:bg-zinc-950">
      <Navbar showAuth={false} />
      <main className="flex flex-1 items-center justify-center px-8 py-16">
        <div className="flex w-full max-w-md flex-col items-center gap-8">
          <div className="text-center">
            <h1 className="text-4xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
              {t("common.title")}
            </h1>
            <p className="mt-2 text-lg text-zinc-600 dark:text-zinc-400">
              {t("auth.loginDescription")}
            </p>
          </div>

          <div className="w-full rounded-xl border border-zinc-200 bg-white p-8 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
            <h2 className="mb-6 text-center text-xl font-semibold text-zinc-900 dark:text-zinc-50">
              {t("auth.signIn")}
            </h2>

            {error && (
              <div className="mb-4 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-400">
                {error.message}
              </div>
            )}

            <div className="flex flex-col gap-4">
              {/* Google Sign-In Button Container */}
              <div
                id="google-signin-button"
                className="flex items-center justify-center"
              >
                {!isReady && (
                  <div className="flex items-center justify-center gap-3 w-full px-6 py-3 bg-zinc-100 rounded-lg dark:bg-zinc-800">
                    <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-600 dark:border-zinc-600 dark:border-t-zinc-300" />
                    <span className="text-zinc-500 dark:text-zinc-400">
                      {t("common.loading")}
                    </span>
                  </div>
                )}
              </div>

              {isLoading && (
                <div className="flex items-center justify-center gap-2 text-sm text-zinc-500 dark:text-zinc-400">
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-600 dark:border-zinc-600 dark:border-t-zinc-300" />
                  <span>{t("auth.signingIn")}</span>
                </div>
              )}
            </div>
          </div>

          <p className="text-center text-sm text-zinc-500 dark:text-zinc-400">
            {t("auth.termsNotice")}
          </p>
        </div>
      </main>
    </div>
  );
}
