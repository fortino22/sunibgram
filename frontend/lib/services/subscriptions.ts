import { api } from '../api';

// ── Backend response shape ──────────────────────────────────────────────────

interface SubscriptionStatusResponse {
  is_premium: boolean;
  is_trial: boolean;
  end_date?: string;
}

// ── Frontend model shape ─────────────────────────────────────────────────────

export interface SubscriptionStatus {
  isPremium: boolean;
  isTrial: boolean;
  endDate: string | null;
}

function mapStatus(raw: SubscriptionStatusResponse): SubscriptionStatus {
  return {
    isPremium: raw.is_premium,
    isTrial: raw.is_trial,
    endDate: raw.end_date ?? null,
  };
}

// ── API calls ────────────────────────────────────────────────────────────────

export async function getSubscriptionStatus(token: string): Promise<SubscriptionStatus> {
  const raw = await api.get<SubscriptionStatusResponse>('/api/subscriptions/me', token);
  return mapStatus(raw);
}

export function activateTrial(token: string): Promise<{ message: string }> {
  return api.post<{ message: string }>('/api/subscriptions/trial/activate', undefined, token);
}
