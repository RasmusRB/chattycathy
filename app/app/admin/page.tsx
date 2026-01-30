"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, X, Shield, ChevronDown, ChevronUp, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Navbar } from "@/components/navbar";
import { useAuthStore } from "@/lib/stores/auth-store";
import { api, type Role, type Permission } from "@/lib/api/client";

export default function AdminPage() {
  const t = useTranslations();
  const router = useRouter();
  const queryClient = useQueryClient();
  
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const hasPermission = useAuthStore((state) => state.hasPermission);
  
  const [expandedRole, setExpandedRole] = useState<number | null>(null);
  const [pendingChanges, setPendingChanges] = useState<Record<number, number[]>>({});
  const [isHydrated, setIsHydrated] = useState(false);

  // Wait for client-side hydration
  useEffect(() => {
    setIsHydrated(true);
  }, []);

  // Redirect if not authenticated or not admin
  useEffect(() => {
    if (isHydrated && (!isAuthenticated || !hasPermission("roles:update"))) {
      router.push("/login");
    }
  }, [isHydrated, isAuthenticated, hasPermission, router]);

  // Fetch roles
  const { data: roles, isLoading: rolesLoading } = useQuery({
    queryKey: ["admin", "roles"],
    queryFn: () => api.listRoles(),
    enabled: isHydrated && isAuthenticated && hasPermission("roles:update"),
  });

  // Fetch permissions
  const { data: permissions, isLoading: permissionsLoading } = useQuery({
    queryKey: ["admin", "permissions"],
    queryFn: () => api.listPermissions(),
    enabled: isHydrated && isAuthenticated && hasPermission("roles:update"),
  });

  // Mutation to update role permissions
  const updatePermissionsMutation = useMutation({
    mutationFn: ({ roleId, permissionIds }: { roleId: number; permissionIds: number[] }) =>
      api.setRolePermissions(roleId, permissionIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "roles"] });
    },
  });

  // Group permissions by resource
  const groupedPermissions = permissions?.reduce<Record<string, Permission[]>>((acc, perm) => {
    if (!acc[perm.resource]) {
      acc[perm.resource] = [];
    }
    acc[perm.resource].push(perm);
    return acc;
  }, {}) ?? {};

  // Get current permissions for a role (considering pending changes)
  const getRolePermissionIds = (role: Role): number[] => {
    if (pendingChanges[role.id] !== undefined) {
      return pendingChanges[role.id];
    }
    return role.permissions.map((p) => p.id);
  };

  // Toggle a permission for a role
  const togglePermission = (role: Role, permissionId: number) => {
    const currentIds = getRolePermissionIds(role);
    const newIds = currentIds.includes(permissionId)
      ? currentIds.filter((id) => id !== permissionId)
      : [...currentIds, permissionId];
    
    setPendingChanges((prev) => ({
      ...prev,
      [role.id]: newIds,
    }));
  };

  // Check if a role has pending changes
  const hasPendingChanges = (role: Role): boolean => {
    if (pendingChanges[role.id] === undefined) return false;
    const originalIds = role.permissions.map((p) => p.id).sort();
    const currentIds = [...pendingChanges[role.id]].sort();
    return JSON.stringify(originalIds) !== JSON.stringify(currentIds);
  };

  // Save changes for a role
  const saveChanges = async (role: Role) => {
    const permissionIds = pendingChanges[role.id];
    if (permissionIds === undefined) return;
    
    await updatePermissionsMutation.mutateAsync({ roleId: role.id, permissionIds });
    setPendingChanges((prev) => {
      const newChanges = { ...prev };
      delete newChanges[role.id];
      return newChanges;
    });
  };

  // Cancel changes for a role
  const cancelChanges = (roleId: number) => {
    setPendingChanges((prev) => {
      const newChanges = { ...prev };
      delete newChanges[roleId];
      return newChanges;
    });
  };

  // Loading state
  if (!isHydrated || rolesLoading || permissionsLoading) {
    return (
      <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950">
        <Navbar />
        <main className="mx-auto max-w-4xl px-4 py-8">
          <div className="flex items-center justify-center py-12">
            <Loader2 className="size-8 animate-spin text-zinc-400" />
          </div>
        </main>
      </div>
    );
  }

  // Not authorized
  if (!isAuthenticated || !hasPermission("roles:update")) {
    return null;
  }

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950">
      <Navbar />
      <main className="mx-auto max-w-4xl px-4 py-8">
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-50">
            {t("admin.title")}
          </h1>
          <p className="mt-1 text-zinc-600 dark:text-zinc-400">
            {t("admin.description")}
          </p>
        </div>

        {/* Roles List */}
        <div className="space-y-4">
          {roles?.map((role) => (
            <div
              key={role.id}
              className="rounded-lg border border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900"
            >
              {/* Role Header */}
              <button
                onClick={() => setExpandedRole(expandedRole === role.id ? null : role.id)}
                className="flex w-full items-center justify-between px-4 py-3 text-left"
              >
                <div className="flex items-center gap-3">
                  <Shield className="size-5 text-zinc-500" />
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-zinc-900 dark:text-zinc-50">
                        {role.name}
                      </span>
                      {role.is_system && (
                        <span className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400">
                          {t("admin.systemRole")}
                        </span>
                      )}
                      {hasPendingChanges(role) && (
                        <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs text-amber-700 dark:bg-amber-900/30 dark:text-amber-400">
                          {t("admin.unsavedChanges")}
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-zinc-500 dark:text-zinc-400">
                      {role.description}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-zinc-500">
                    {getRolePermissionIds(role).length} {t("admin.permissions")}
                  </span>
                  {expandedRole === role.id ? (
                    <ChevronUp className="size-5 text-zinc-400" />
                  ) : (
                    <ChevronDown className="size-5 text-zinc-400" />
                  )}
                </div>
              </button>

              {/* Expanded Permissions */}
              {expandedRole === role.id && (
                <div className="border-t border-zinc-200 px-4 py-4 dark:border-zinc-800">
                  {Object.entries(groupedPermissions).map(([resource, perms]) => (
                    <div key={resource} className="mb-4 last:mb-0">
                      <h3 className="mb-2 text-sm font-medium text-zinc-700 capitalize dark:text-zinc-300">
                        {resource}
                      </h3>
                      <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
                        {perms.map((perm) => {
                          const isEnabled = getRolePermissionIds(role).includes(perm.id);
                          return (
                            <button
                              key={perm.id}
                              onClick={() => togglePermission(role, perm.id)}
                              className={`flex items-center gap-2 rounded-md border px-3 py-2 text-left text-sm transition-colors ${
                                isEnabled
                                  ? "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400"
                                  : "border-zinc-200 bg-zinc-50 text-zinc-600 hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-700"
                              }`}
                              title={perm.description}
                            >
                              {isEnabled ? (
                                <Check className="size-4 shrink-0" />
                              ) : (
                                <X className="size-4 shrink-0 opacity-40" />
                              )}
                              <span className="truncate">{perm.action}</span>
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  ))}

                  {/* Save/Cancel Buttons */}
                  {hasPendingChanges(role) && (
                    <div className="mt-4 flex justify-end gap-2 border-t border-zinc-200 pt-4 dark:border-zinc-800">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => cancelChanges(role.id)}
                      >
                        {t("admin.cancel")}
                      </Button>
                      <Button
                        size="sm"
                        onClick={() => saveChanges(role)}
                        disabled={updatePermissionsMutation.isPending}
                      >
                        {updatePermissionsMutation.isPending ? (
                          <Loader2 className="mr-2 size-4 animate-spin" />
                        ) : null}
                        {t("admin.saveChanges")}
                      </Button>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}
