'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getSubscriptionStatus,
  activateTrial,
  type SubscriptionStatus,
} from '@/lib/services/subscriptions';
import { useAuthStore } from '@/stores/authStore';

const ONE_HOUR = 60 * 60 * 1000;

export function useSubscriptionStatus() {
  const token = useAuthStore((s) => s.token);

  return useQuery<SubscriptionStatus>({
    queryKey: ['subscription', 'me'],
    queryFn: () => {
      if (token === null) throw new Error('Authentication token required.');
      return getSubscriptionStatus(token);
    },
    enabled: token !== null,
    staleTime: ONE_HOUR,
    gcTime: ONE_HOUR * 24,
  });
}

export function useActivateTrial() {
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => {
      if (token === null) throw new Error('Authentication token required.');
      return activateTrial(token);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['subscription', 'me'] });
      void queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
    },
  });
}
