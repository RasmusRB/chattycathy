import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

export interface User {
  id: number;
  email: string;
  name: string;
  picture: string;
  role: string;
  permissions?: string[];
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  access_token_expires_in: number;
  token_type: string;
}

interface AuthState {
  // User
  user: User | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;

  // Actions
  setAuth: (user: User, tokens: AuthTokens) => void;
  setTokens: (tokens: AuthTokens) => void;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
  hasAnyPermission: (...permissions: string[]) => boolean;
  hasAllPermissions: (...permissions: string[]) => boolean;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set, get) => ({
        user: null,
        tokens: null,
        isAuthenticated: false,

        setAuth: (user, tokens) =>
          set({
            user,
            tokens,
            isAuthenticated: true,
          }),

        setTokens: (tokens) =>
          set((state) => ({
            tokens,
            isAuthenticated: state.user !== null,
          })),

        logout: () =>
          set({
            user: null,
            tokens: null,
            isAuthenticated: false,
          }),

        hasPermission: (permission: string) => {
          const user = get().user;
          return user?.permissions?.includes(permission) ?? false;
        },

        hasAnyPermission: (...permissions: string[]) => {
          const user = get().user;
          if (!user?.permissions) return false;
          return permissions.some((p) => user.permissions?.includes(p));
        },

        hasAllPermissions: (...permissions: string[]) => {
          const user = get().user;
          if (!user?.permissions) return false;
          return permissions.every((p) => user.permissions?.includes(p));
        },
      }),
      {
        name: "auth-storage",
        partialize: (state) => ({
          user: state.user,
          tokens: state.tokens,
          isAuthenticated: state.isAuthenticated,
        }),
      }
    ),
    { name: "auth" }
  )
);
