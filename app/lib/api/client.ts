import { z } from "zod";

// API base URL from environment variable
// Set NEXT_PUBLIC_API_URL in .env.local or docker-compose
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

// Ping response schema
export const pingResponseSchema = z.object({
  message: z.string(),
});

export type PingResponse = z.infer<typeof pingResponseSchema>;

// Permission schema
export const permissionSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string(),
  resource: z.string(),
  action: z.string(),
});

export type Permission = z.infer<typeof permissionSchema>;

// Role schema
export const roleSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string(),
  is_system: z.boolean(),
  permissions: z.array(permissionSchema),
});

export type Role = z.infer<typeof roleSchema>;

// User schema
export const userSchema = z.object({
  id: z.number(),
  email: z.string(),
  name: z.string(),
  picture: z.string(),
  role: z.string(),
  permissions: z.array(z.string()).nullable().optional().transform((val) => val ?? []),
});

export type User = z.infer<typeof userSchema>;

// Auth response schema
export const authResponseSchema = z.object({
  access_token: z.string(),
  refresh_token: z.string(),
  access_token_expires_in: z.number(),
  token_type: z.string(),
  user: userSchema,
});

export type AuthResponse = z.infer<typeof authResponseSchema>;

// Google config schema
export const googleConfigSchema = z.object({
  client_id: z.string(),
  redirect_uri: z.string(),
});

export type GoogleConfig = z.infer<typeof googleConfigSchema>;

// Helper to get auth header
const getAuthHeader = (): Record<string, string> => {
  if (typeof globalThis.window === "undefined") return {};
  const authStorage = localStorage.getItem("auth-storage");
  if (!authStorage) return {};
  try {
    const parsed = JSON.parse(authStorage);
    const token = parsed?.state?.tokens?.access_token;
    if (token) {
      return { Authorization: `Bearer ${token}` };
    }
  } catch {
    return {};
  }
  return {};
};

// API client
export const api = {
  ping: async (): Promise<PingResponse> => {
    const response = await fetch(`${API_BASE_URL}/ping`, {
      headers: getAuthHeader(),
    });
    if (!response.ok) {
      throw new Error(`API error: ${response.status}`);
    }
    const data = await response.json();
    return pingResponseSchema.parse(data);
  },

  // Google OAuth
  getGoogleConfig: async (): Promise<GoogleConfig> => {
    const response = await fetch(`${API_BASE_URL}/auth/google/config`);
    if (!response.ok) {
      throw new Error(`API error: ${response.status}`);
    }
    const data = await response.json();
    return googleConfigSchema.parse(data);
  },

  googleAuth: async (accessToken: string): Promise<AuthResponse> => {
    const response = await fetch(`${API_BASE_URL}/auth/google`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ access_token: accessToken }),
      credentials: "include",
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `API error: ${response.status}`);
    }
    const data = await response.json();
    return authResponseSchema.parse(data);
  },

  // Refresh token
  refresh: async (): Promise<AuthResponse> => {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: "POST",
      credentials: "include",
    });
    if (!response.ok) {
      throw new Error(`Refresh failed: ${response.status}`);
    }
    const data = await response.json();
    return authResponseSchema.parse(data);
  },

  // Logout
  logout: async (): Promise<void> => {
    await fetch(`${API_BASE_URL}/auth/logout`, {
      method: "POST",
      credentials: "include",
    });
  },

  // Admin: List all permissions
  listPermissions: async (): Promise<Permission[]> => {
    const response = await fetch(`${API_BASE_URL}/admin/permissions`, {
      headers: getAuthHeader(),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `API error: ${response.status}`);
    }
    const data = await response.json();
    return z.array(permissionSchema).parse(data);
  },

  // Admin: List all roles
  listRoles: async (): Promise<Role[]> => {
    const response = await fetch(`${API_BASE_URL}/admin/roles`, {
      headers: getAuthHeader(),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `API error: ${response.status}`);
    }
    const data = await response.json();
    return z.array(roleSchema).parse(data);
  },

  // Admin: Get a single role
  getRole: async (id: number): Promise<Role> => {
    const response = await fetch(`${API_BASE_URL}/admin/roles/${id}`, {
      headers: getAuthHeader(),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `API error: ${response.status}`);
    }
    const data = await response.json();
    return roleSchema.parse(data);
  },

  // Admin: Create a new role
  createRole: async (name: string, description: string): Promise<Role> => {
    const response = await fetch(`${API_BASE_URL}/admin/roles`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...getAuthHeader(),
      },
      body: JSON.stringify({ name, description }),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `API error: ${response.status}`);
    }
    const data = await response.json();
    return roleSchema.parse(data);
  },

  // Admin: Update a role
  updateRole: async (id: number, name: string, description: string): Promise<Role> => {
    const response = await fetch(`${API_BASE_URL}/admin/roles/${id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        ...getAuthHeader(),
      },
      body: JSON.stringify({ name, description }),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `API error: ${response.status}`);
    }
    const data = await response.json();
    return roleSchema.parse(data);
  },

  // Admin: Delete a role
  deleteRole: async (id: number): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/admin/roles/${id}`, {
      method: "DELETE",
      headers: getAuthHeader(),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `API error: ${response.status}`);
    }
  },

  // Admin: Set role permissions
  setRolePermissions: async (roleId: number, permissionIds: number[]): Promise<Role> => {
    const response = await fetch(`${API_BASE_URL}/admin/roles/${roleId}/permissions`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        ...getAuthHeader(),
      },
      body: JSON.stringify({ permission_ids: permissionIds }),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `API error: ${response.status}`);
    }
    const data = await response.json();
    return roleSchema.parse(data);
  },
};
