'use client';

import { useQuery } from '@tanstack/react-query';
import { getOnboardingStatus } from '@/lib/services/profile';
import { useAuthStore } from '@/stores/authStore';

// Sizes the first-time follow gate. `requiredFollows` is the target capped at
// the followable pool (0 when no one is followable → gate is skippable).
export function useOnboardingStatus() {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['onboarding', 'status'],
    queryFn: () => {
      if (!token) throw new Error('Not authenticated');
      return getOnboardingStatus(token);
    },
    enabled: token !== null,
  });
}
