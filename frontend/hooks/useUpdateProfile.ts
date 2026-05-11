'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateMyProfile, type UpdateProfileParams } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

export function useUpdateProfile() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: UpdateProfileParams) => updateMyProfile(params, token!),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
      // Refetch all thread feeds so embedded author avatars update immediately.
      void queryClient.invalidateQueries({ queryKey: ['threads'] });
    },
  });
}
