'use client';

import { Suspense } from 'react';
import { useParams, useSearchParams, useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import { FollowListTab } from '@/components/profile/FollowListTab';
import { useProfile } from '@/hooks/useProfile';

type FollowTab = 'verified' | 'followers' | 'following';

const TABS: { id: FollowTab; label: string }[] = [
  { id: 'verified', label: 'Verified followers' },
  { id: 'followers', label: 'Followers' },
  { id: 'following', label: 'Following' },
];

function FollowersPageInner() {
  const params = useParams<{ userId: string }>();
  const searchParams = useSearchParams();
  const router = useRouter();
  const userId = Number(params.userId);
  const activeTab = (searchParams.get('tab') as FollowTab) ?? 'followers';

  const { data: profile } = useProfile(userId);

  function setTab(tab: FollowTab) {
    const url = new URLSearchParams(searchParams.toString());
    url.set('tab', tab);
    router.replace(`/profile/${userId}/followers?${url.toString()}`);
  }

  return (
    <div>
      {/* Header */}
      <header
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 20,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '12px 16px',
          display: 'flex',
          alignItems: 'center',
          gap: 14,
        }}
      >
        <button
          type="button"
          aria-label="Go back"
          onClick={() => router.back()}
          className="btn-ghost"
          style={{
            border: 'none',
            cursor: 'pointer',
            color: 'var(--text)',
            padding: 6,
            borderRadius: '50%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}
        >
          <ArrowLeft size={20} strokeWidth={1.8} />
        </button>
        <div>
          <div
            style={{
              fontWeight: 700,
              fontSize: 17,
              color: 'var(--text)',
              letterSpacing: '-0.02em',
              lineHeight: 1.2,
            }}
          >
            {profile?.name ?? profile?.username ?? ''}
          </div>
          {profile?.username && (
            <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
              @{profile.username}
            </div>
          )}
        </div>
      </header>

      {/* Tabs */}
      <div
        role="tablist"
        style={{
          display: 'flex',
          borderBottom: '1px solid var(--border)',
          position: 'sticky',
          top: 65,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
        }}
      >
        {TABS.map((tab) => (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={activeTab === tab.id}
            onClick={() => setTab(tab.id)}
            className="btn-ghost"
            style={{
              flex: 1,
              padding: '16px 8px',
              border: 'none',
              cursor: 'pointer',
              fontSize: 14,
              fontWeight: activeTab === tab.id ? 700 : 400,
              color: activeTab === tab.id ? 'var(--text)' : 'var(--text-tertiary)',
              borderBottom: activeTab === tab.id ? '2px solid var(--accent)' : '2px solid transparent',
              fontFamily: 'inherit',
              letterSpacing: '-0.01em',
              whiteSpace: 'nowrap',
              transition: 'color 0.15s ease, border-color 0.15s ease',
              marginBottom: -1,
            }}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Content */}
      {activeTab === 'verified' && (
        <FollowListTab userId={userId} type="followers" filterPremium />
      )}
      {activeTab === 'followers' && (
        <FollowListTab userId={userId} type="followers" />
      )}
      {activeTab === 'following' && (
        <FollowListTab userId={userId} type="following" />
      )}
    </div>
  );
}

export default function FollowersPage() {
  return (
    <Suspense>
      <FollowersPageInner />
    </Suspense>
  );
}
