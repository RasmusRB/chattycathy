import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import { useAppStore } from "../stores/app-store";

export const pingKeys = {
  all: ["ping"] as const,
  latest: () => [...pingKeys.all, "latest"] as const,
};

export function usePing() {
  const addPing = useAppStore((state) => state.addPing);

  return useQuery({
    queryKey: pingKeys.latest(),
    queryFn: async () => {
      const response = await api.ping();
      addPing(response.message);
      return response;
    },
    enabled: false, // Don't auto-fetch, only on demand
  });
}

export function usePingMutation() {
  const queryClient = useQueryClient();
  const addPing = useAppStore((state) => state.addPing);

  return useMutation({
    mutationFn: api.ping,
    onSuccess: (data) => {
      addPing(data.message);
      queryClient.setQueryData(pingKeys.latest(), data);
    },
  });
}
