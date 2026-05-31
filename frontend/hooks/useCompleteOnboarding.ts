'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { completeOnboarding } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

// Marks first-time onboarding complete. On success the profile is refetched so
// the gate (onboarding_completed) flips and the blocking modal unmounts.
export function useCompleteOnboarding() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => {
      if (!token) throw new Error('Not authenticated');
      return completeOnboarding(token);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
    },
  });
}
