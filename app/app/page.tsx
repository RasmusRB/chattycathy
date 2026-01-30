"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/lib/stores/auth-store";

export default function Home() {
  const router = useRouter();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  useEffect(() => {
    // Redirect based on auth status
    if (isAuthenticated) {
      router.push("/ping");
    } else {
      router.push("/login");
    }
  }, [isAuthenticated, router]);

  // Show loading while redirecting
  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-zinc-950">
      <div className="flex items-center gap-3 text-zinc-500 dark:text-zinc-400">
        <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-600 dark:border-zinc-600 dark:border-t-zinc-300" />
        <span>Loading...</span>
      </div>
    </div>
  );
}
