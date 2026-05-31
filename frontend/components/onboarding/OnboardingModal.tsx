'use client';

import { useMemo, useState } from 'react';
import { Search } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useOnboardingPeople } from '@/hooks/useExplore';
import { useFollowUser } from '@/hooks/useFollowUser';
import { useCompleteOnboarding } from '@/hooks/useCompleteOnboarding';
import type { ExplorePerson } from '@/lib/services/explore';

// Number of people a first-time user must follow before continuing. Mirrors the
// server-side onboardingMinFollows guard — the backend re-validates, so this is
// only the UX gate.
const MIN_FOLLOWS = 5;

interface OnboardingModalProps {
  /** Controlled by the profile gate in MainWrapper. */
  open: boolean;
}

function PersonRow({
  person,
  initialFollowing,
  onToggle,
}: {
  person: ExplorePerson;
  initialFollowing: boolean;
  onToggle: (userId: number, following: boolean) => void;
}) {
  const [following, setFollowing] = useState(initialFollowing);
  const { follow, unfollow } = useFollowUser(person.userId);

  function handleClick() {
    if (following) {
      setFollowing(false);
      onToggle(person.userId, false);
      unfollow.mutate(undefined, {
        onError: () => {
          setFollowing(true);
          onToggle(person.userId, true);
        },
      });
    } else {
      setFollowing(true);
      onToggle(person.userId, true);
      follow.mutate(undefined, {
        onError: () => {
          setFollowing(false);
          onToggle(person.userId, false);
        },
      });
    }
  }

  return (
    <div className="onboarding-row flex items-center gap-3" style={{ padding: '10px 12px' }}>
      <Avatar username={person.username} avatarUrl={person.userProfilePic || null} size={44} />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1">
          <span className="truncate text-[15px] font-bold" style={{ color: 'var(--text)' }}>
            {person.name}
          </span>
          {person.isPremium && <VerifiedBadge size={15} />}
        </div>
        <div className="truncate text-[14px]" style={{ color: 'var(--text-tertiary)' }}>
          @{person.username}
        </div>
      </div>
      <button
        type="button"
        onClick={handleClick}
        className="shrink-0 rounded-full text-[14px] font-bold transition-colors"
        style={{
          padding: '7px 16px',
          border: 'none',
          ...(following
            ? { background: 'var(--bg-subtle)', color: 'var(--text)' }
            : { background: 'var(--text)', color: 'var(--bg)' }),
        }}
      >
        {following ? 'Following' : 'Follow'}
      </button>
    </div>
  );
}

export function OnboardingModal({ open }: OnboardingModalProps) {
  const [query, setQuery] = useState('');
  const [followedIds, setFollowedIds] = useState<Set<number>>(new Set());

  const trimmed = query.trim();
  const activeQuery = useOnboardingPeople(trimmed);
  const complete = useCompleteOnboarding();

  const isSearching = trimmed.length > 0;
  const people = activeQuery.data ?? [];

  const followedCount = followedIds.size;
  const remaining = Math.max(0, MIN_FOLLOWS - followedCount);
  const canContinue = followedCount >= MIN_FOLLOWS;

  function handleToggle(userId: number, following: boolean) {
    setFollowedIds((prev) => {
      const next = new Set(prev);
      if (following) next.add(userId);
      else next.delete(userId);
      return next;
    });
  }

  const rows = useMemo(
    () =>
      people.map((person) => (
        <PersonRow
          key={person.userId}
          person={person}
          initialFollowing={followedIds.has(person.userId) || person.isFollowing}
          onToggle={handleToggle}
        />
      )),
    // followedIds intentionally excluded: rows manage their own state, parent
    // only needs the initial value at mount. Re-keying on it would reset rows.
    [people] // eslint-disable-line react-hooks/exhaustive-deps
  );

  if (!open) return null;

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 60,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
        backdropFilter: 'blur(4px)',
        padding: 16,
      }}
      // No click-outside close — hard block until the gate is cleared.
    >
      <div
        style={{
          background: 'var(--bg-card)',
          borderRadius: 20,
          width: '100%',
          maxWidth: 480,
          maxHeight: '85vh',
          display: 'flex',
          flexDirection: 'column',
          boxShadow: '0 24px 60px -12px rgba(0,0,0,0.5)',
        }}
      >
        <div style={{ padding: '24px 24px 8px', flexShrink: 0 }}>
          <h2 className="text-[22px] font-extrabold" style={{ color: 'var(--text)' }}>
            Find people to follow
          </h2>
          <p className="text-[14px]" style={{ marginTop: 4, color: 'var(--text-tertiary)' }}>
            Follow at least {MIN_FOLLOWS} people to personalize your feed.
          </p>

          {/* Progress — slim borderless bar + count */}
          <div className="flex items-center gap-3" style={{ marginTop: 14 }}>
            <div style={{ flex: 1, height: 4, borderRadius: 999, background: 'var(--bg-subtle)', overflow: 'hidden' }}>
              <div
                style={{
                  height: '100%',
                  borderRadius: 999,
                  background: 'var(--text)',
                  width: `${Math.min(followedCount / MIN_FOLLOWS, 1) * 100}%`,
                  transition: 'width 0.2s ease',
                }}
              />
            </div>
            <span className="text-[13px] font-semibold tabular-nums" style={{ color: 'var(--text-tertiary)' }}>
              {Math.min(followedCount, MIN_FOLLOWS)}/{MIN_FOLLOWS}
            </span>
          </div>

          <div
            className="flex items-center gap-2"
            style={{
              marginTop: 14,
              padding: '10px 14px',
              borderRadius: 9999,
              background: 'var(--bg-subtle)',
            }}
          >
            <Search size={18} style={{ color: 'var(--text-tertiary)' }} />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search people"
              className="w-full bg-transparent text-[15px] outline-none"
              style={{ color: 'var(--text)' }}
            />
          </div>
        </div>

        <div style={{ overflowY: 'auto', flex: 1, padding: '4px 12px' }}>
          {activeQuery.isLoading && (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading…
            </div>
          )}
          {!activeQuery.isLoading && people.length === 0 && (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              {isSearching ? 'No people found' : 'No suggestions right now'}
            </div>
          )}
          {rows}
        </div>

        <div style={{ padding: '12px 24px 24px', flexShrink: 0 }}>
          <button
            type="button"
            disabled={!canContinue || complete.isPending}
            onClick={() => complete.mutate()}
            className="w-full rounded-full text-[15px] font-bold transition-opacity"
            style={{
              padding: '12px 16px',
              background: 'var(--text)',
              color: 'var(--bg)',
              opacity: canContinue && !complete.isPending ? 1 : 0.5,
              cursor: canContinue && !complete.isPending ? 'pointer' : 'not-allowed',
            }}
          >
            {complete.isPending
              ? 'Setting up…'
              : canContinue
                ? 'Continue'
                : `Follow ${remaining} more`}
          </button>
          {complete.isError && (
            <p className="text-[13px]" style={{ marginTop: 8, textAlign: 'center', color: '#f4212e' }}>
              {complete.error instanceof Error ? complete.error.message : 'Something went wrong'}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
