"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/lib/api/client";
import { useAuthStore } from "@/lib/stores/auth-store";

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string;
            callback: (response: { credential: string }) => void;
            auto_select?: boolean;
            cancel_on_tap_outside?: boolean;
          }) => void;
          renderButton: (
            element: HTMLElement,
            config: {
              theme?: "outline" | "filled_blue" | "filled_black";
              size?: "large" | "medium" | "small";
              type?: "standard" | "icon";
              text?: "signin_with" | "signup_with" | "continue_with" | "signin";
              shape?: "rectangular" | "pill" | "circle" | "square";
              logo_alignment?: "left" | "center";
              width?: number;
            }
          ) => void;
          prompt: () => void;
          disableAutoSelect: () => void;
        };
        oauth2: {
          initTokenClient: (config: {
            client_id: string;
            scope: string;
            callback: (response: { access_token?: string; error?: string }) => void;
          }) => {
            requestAccessToken: () => void;
          };
        };
      };
    };
  }
}

interface UseGoogleAuthOptions {
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

export function useGoogleAuth(options?: UseGoogleAuthOptions) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [clientId, setClientId] = useState<string | null>(null);
  const [isScriptLoaded, setIsScriptLoaded] = useState(false);

  const setAuth = useAuthStore((state) => state.setAuth);

  // Load Google config and script
  useEffect(() => {
    const loadGoogleScript = async () => {
      try {
        // Get client ID from environment variable
        const googleClientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;
        if (!googleClientId) {
          setError(new Error("NEXT_PUBLIC_GOOGLE_CLIENT_ID not configured"));
          return;
        }
        setClientId(googleClientId);

        // Check if script is already loaded
        if (globalThis.window?.google?.accounts) {
          setIsScriptLoaded(true);
          return;
        }

        // Load Google Identity Services script
        const script = document.createElement("script");
        script.src = "https://accounts.google.com/gsi/client";
        script.async = true;
        script.defer = true;
        script.onload = () => setIsScriptLoaded(true);
        script.onerror = () => setError(new Error("Failed to load Google script"));
        document.body.appendChild(script);
      } catch (err) {
        setError(err instanceof Error ? err : new Error("Failed to load Google config"));
      }
    };

    loadGoogleScript();
  }, []);

  // Handle Google sign-in response
  const handleGoogleResponse = useCallback(
    async (accessToken: string) => {
      setIsLoading(true);
      setError(null);

      try {
        const response = await api.googleAuth(accessToken);
        setAuth(response.user, {
          access_token: response.access_token,
          refresh_token: response.refresh_token,
          access_token_expires_in: response.access_token_expires_in,
          token_type: response.token_type,
        });
        options?.onSuccess?.();
      } catch (err) {
        const error = err instanceof Error ? err : new Error("Google auth failed");
        setError(error);
        options?.onError?.(error);
      } finally {
        setIsLoading(false);
      }
    },
    [setAuth, options]
  );

  // Initialize Google Sign-In button
  const renderGoogleButton = useCallback(
    (elementId: string) => {
      if (!isScriptLoaded || !clientId || !globalThis.window?.google) return;

      const element = document.getElementById(elementId);
      if (!element) return;

      // Use OAuth2 token client for access token flow
      const tokenClient = globalThis.window.google.accounts.oauth2.initTokenClient({
        client_id: clientId,
        scope: "email profile",
        callback: (response) => {
          if (response.error) {
            setError(new Error(response.error));
            options?.onError?.(new Error(response.error));
            return;
          }
          if (response.access_token) {
            handleGoogleResponse(response.access_token);
          }
        },
      });

      // Create custom button that triggers the token client
      element.innerHTML = "";
      const button = document.createElement("button");
      button.className =
        "flex items-center justify-center gap-3 w-full px-6 py-3 bg-white border border-zinc-300 rounded-lg shadow-sm hover:bg-zinc-50 transition-colors font-medium text-zinc-700 dark:bg-zinc-800 dark:border-zinc-700 dark:hover:bg-zinc-700 dark:text-zinc-200";
      button.innerHTML = `
        <svg width="20" height="20" viewBox="0 0 24 24">
          <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
          <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
          <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
          <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
        </svg>
        <span>Sign in with Google</span>
      `;
      button.onclick = () => tokenClient.requestAccessToken();
      element.appendChild(button);
    },
    [isScriptLoaded, clientId, handleGoogleResponse, options]
  );

  return {
    isLoading,
    error,
    isReady: isScriptLoaded && !!clientId,
    renderGoogleButton,
  };
}
