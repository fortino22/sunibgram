'use client';

import { useSubscriptionStatus, useActivateTrial } from '@/hooks/useSubscription';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';

const CARD_CLASS = 'overflow-hidden rounded-2xl';
const CARD_STYLE = { background: 'var(--bg-card)', border: '1px solid var(--border)' } as const;
const CONTENT_PADDING = '14px 16px';

function formatDate(value: string | null): string {
  if (!value) return '';
  return new Date(value).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

export function PremiumCard() {
  const { data, isLoading } = useSubscriptionStatus();
  const activateTrial = useActivateTrial();

  if (isLoading) {
    return (
      <section className={CARD_CLASS} style={CARD_STYLE}>
        <div style={{ padding: CONTENT_PADDING, display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div className="h-4 w-36 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
          <div className="h-3 w-full animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
          <div className="h-7 w-24 animate-pulse rounded-full" style={{ background: 'var(--bg-subtle)' }} />
        </div>
      </section>
    );
  }

  // Active premium / trial state — show status instead of the CTA.
  if (data?.isPremium) {
    return (
      <section className={CARD_CLASS} style={CARD_STYLE}>
        <div style={{ padding: CONTENT_PADDING }}>
          <div className="flex items-center gap-1.5">
            <VerifiedBadge size={16} />
            <h2 className="text-[18px] font-extrabold" style={{ color: 'var(--text)' }}>
              Premium active
            </h2>
          </div>
          <p className="text-[13px]" style={{ marginTop: 2, color: 'var(--text-tertiary)' }}>
            {data.isTrial ? 'Free trial' : 'Subscription'} active
            {data.endDate ? ` until ${formatDate(data.endDate)}` : ''}.
          </p>
        </div>
      </section>
    );
  }

  return (
    <section className={CARD_CLASS} style={CARD_STYLE}>
      <div style={{ padding: CONTENT_PADDING }}>
        <div className="flex flex-wrap items-center gap-1.5">
          <h2 className="text-[18px] font-extrabold" style={{ color: 'var(--text)' }}>
            Subscribe to Premium
          </h2>
          <span
            className="rounded-full text-[12px] font-bold"
            style={{ padding: '2px 8px', background: 'var(--green-soft)', color: 'var(--green)' }}
          >
            1 month free
          </span>
        </div>

        <p className="text-[15px] leading-tight" style={{ marginTop: 4, color: 'var(--text-secondary)' }}>
          Unlock access to more features.
        </p>

        <button
          type="button"
          onClick={() => activateTrial.mutate()}
          disabled={activateTrial.isPending}
          className="rounded-full text-[15px] font-bold transition-opacity hover:opacity-90 disabled:opacity-60"
          style={{ marginTop: 12, padding: '8px 18px', background: 'var(--accent)', color: '#fff' }}
        >
          {activateTrial.isPending ? 'Activating…' : 'Start free trial'}
        </button>

        {activateTrial.isError && (
          <p className="text-[13px]" style={{ marginTop: 8, color: 'var(--text-tertiary)' }}>
            {(activateTrial.error as Error).message}
          </p>
        )}
      </div>
    </section>
  );
}
