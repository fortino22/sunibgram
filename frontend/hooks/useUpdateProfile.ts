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
      // Prefix-match invalidation covers ['profile','me'], ['profile', <id>]
      // (own /profile/:id + /about-profile/:id), and the About history list.
      void queryClient.invalidateQueries({ queryKey: ['profile'] });
      void queryClient.invalidateQueries({ queryKey: ['username-history'] });
      // Refetch all thread feeds so embedded author avatars update immediately.
      void queryClient.invalidateQueries({ queryKey: ['threads'] });
    },
  });
}
