'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getMyUserSettings,
  updateMyUserSettings,
  type UserSettings,
} from '@/lib/services/userSettings';
import { useAuthStore } from '@/stores/authStore';

export function useUserSettings() {
  const token = useAuthStore((s) => s.token);
  return useQuery<UserSettings>({
    queryKey: ['user-settings', 'me'],
    queryFn: () => getMyUserSettings(token!),
    enabled: !!token,
    staleTime: 60_000,
  });
}

export function useUpdateUserSettings() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { is_private: boolean }) => updateMyUserSettings(params, token!),
    onMutate: async (params) => {
      await queryClient.cancelQueries({ queryKey: ['user-settings', 'me'] });
      const previous = queryClient.getQueryData<UserSettings>(['user-settings', 'me']);
      if (previous) {
        queryClient.setQueryData<UserSettings>(['user-settings', 'me'], {
          ...previous,
          is_private: params.is_private,
        });
      }
      return { previous };
    },
    onError: (_err, _params, ctx) => {
      if (ctx?.previous) {
        queryClient.setQueryData(['user-settings', 'me'], ctx.previous);
      }
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ['user-settings', 'me'] });
    },
  });
}
