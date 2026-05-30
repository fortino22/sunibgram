'use client';

import { useState } from 'react';
import { Check } from 'lucide-react';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import {
  PrimaryButton,
  SettingsPanel,
  SuccessBanner,
} from '@/components/settings/SettingsPanel';
import { useActivateTrial, useSubscriptionStatus } from '@/hooks/useSubscription';

const PREMIUM_PERKS: string[] = [
  'Verified badge on your profile, posts, and replies',
  'Verified badge in chats, communities, and notifications',
  'Priority ranking in trending users and discover suggestions',
  'Higher placement in followers and following lists',
];

export default function SubscriptionSettingsPage() {
  const { data: subscription, isLoading } = useSubscriptionStatus();
  const { mutate, isPending, error } = useActivateTrial();
  const [success, setSuccess] = useState<string | null>(null);

  function handleActivate() {
    setSuccess(null);
    mutate(undefined, {
      onSuccess: (res) => setSuccess(res.message ?? 'Trial activated.'),
    });
  }

  const isPremium = subscription?.isPremium ?? false;
  const isTrial = subscription?.isTrial ?? false;
  const endDate = subscription?.endDate ?? null;

  return (
    <SettingsPanel
      title="Premium"
      description="Manage your subscription and unlock premium features."
    >
      <div style={{ maxWidth: 640 }}>
        <SuccessBanner message={success} />
        <ErrorAlert message={error ? (error as Error).message : null} />

        {isLoading ? (
          <SkeletonStatus />
        ) : (
          <>
            <StatusCard
              isPremium={isPremium}
              isTrial={isTrial}
              endDate={endDate}
            />

            <div style={{ marginTop: 24 }}>
              <div style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)', marginBottom: 12 }}>
                What you get with Premium
              </div>
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {PREMIUM_PERKS.map((perk) => (
                  <li
                    key={perk}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 10,
                      padding: '10px 0',
                      borderBottom: '1px solid var(--border)',
                      fontSize: 14,
                      color: 'var(--text-secondary)',
                    }}
                  >
                    <Check size={16} color="var(--accent)" />
                    {perk}
                  </li>
                ))}
              </ul>
            </div>

            {!isPremium && (
              <div style={{ marginTop: 24, display: 'flex', justifyContent: 'flex-end' }}>
                <PrimaryButton type="button" onClick={handleActivate} disabled={isPending}>
                  {isPending ? 'Activating…' : 'Activate free trial'}
                </PrimaryButton>
              </div>
            )}
          </>
        )}
      </div>
    </SettingsPanel>
  );
}

function StatusCard({
  isPremium,
  isTrial,
  endDate,
}: {
  isPremium: boolean;
  isTrial: boolean;
  endDate: string | null;
}) {
  const formattedEnd = endDate
    ? new Date(endDate).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    : null;

  return (
    <div
      style={{
        padding: 18,
        borderRadius: 14,
        border: '1px solid var(--border)',
        background: 'var(--bg-subtle)',
        display: 'flex',
        alignItems: 'center',
        gap: 14,
      }}
    >
      <div
        style={{
          width: 44,
          height: 44,
          borderRadius: '50%',
          background: isPremium ? 'var(--accent)' : 'var(--border)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: isPremium ? '#fff' : 'var(--text-tertiary)',
          flexShrink: 0,
        }}
      >
        <Check size={20} strokeWidth={2.5} />
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 16, fontWeight: 700, color: 'var(--text)' }}>
          {isPremium ? (isTrial ? 'Premium (Trial)' : 'Premium') : 'Not subscribed'}
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-tertiary)', marginTop: 2 }}>
          {isPremium
            ? formattedEnd
              ? `Active until ${formattedEnd}`
              : 'Active'
            : 'Activate your free trial to unlock premium features.'}
        </div>
      </div>
    </div>
  );
}

function SkeletonStatus() {
  return (
    <div
      style={{
        padding: 18,
        borderRadius: 14,
        border: '1px solid var(--border)',
        background: 'var(--bg-subtle)',
        display: 'flex',
        alignItems: 'center',
        gap: 14,
      }}
    >
      <div style={{ width: 44, height: 44, borderRadius: '50%', background: 'var(--border)' }} />
      <div style={{ flex: 1 }}>
        <div style={{ width: 140, height: 14, background: 'var(--border)', borderRadius: 4, marginBottom: 6 }} />
        <div style={{ width: 200, height: 12, background: 'var(--border)', borderRadius: 4 }} />
      </div>
    </div>
  );
}
