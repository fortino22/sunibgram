'use client';

import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { FollowButton } from '@/components/shared/FollowButton';
import { UserHoverCard } from '@/components/shared/UserHoverCard';
import { useSuggestedPeoplePreview } from '@/hooks/useExplore';

const SKELETON_COUNT = 3;

interface WhoToFollowProps {
  /** Flat full-width section (no rounded card chrome) — for in-page use like X. */
  bare?: boolean;
}

export function WhoToFollow({ bare = false }: WhoToFollowProps) {
  const router = useRouter();
  const { data, isLoading, isError } = useSuggestedPeoplePreview();

  const sectionClass = bare ? '' : 'overflow-hidden rounded-2xl';
  const sectionStyle = bare
    ? { borderBottom: '1px solid var(--border)' }
    : { background: 'var(--bg-card)', border: '1px solid var(--border)' };
  const headCls = bare ? 'text-[20px]' : 'text-[18px]';
  const headPad = bare ? '16px 16px 6px' : '14px 16px 8px';

  if (isLoading) {
    return (
      <section className={sectionClass} style={sectionStyle}>
        <h2 className={`${headCls} font-extrabold`} style={{ padding: headPad, color: 'var(--text)' }}>Who to follow</h2>
        <div>
          {Array.from({ length: SKELETON_COUNT }, (_, i) => (
            <div key={i} className="flex items-center gap-2.5" style={{ padding: '12px 16px' }}>
              <div className="h-9 w-9 shrink-0 animate-pulse rounded-full" style={{ background: 'var(--bg-subtle)' }} />
              <div className="flex-1" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                <div className="h-3 w-24 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
                <div className="h-2.5 w-16 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
              </div>
              <div className="h-7 w-16 shrink-0 animate-pulse rounded-full" style={{ background: 'var(--bg-subtle)' }} />
            </div>
          ))}
        </div>
      </section>
    );
  }

  if (isError) return null;

  const people = data ?? [];
  if (people.length === 0) return null;

  return (
    <section className={sectionClass} style={sectionStyle}>
      <h2 className={`${headCls} font-extrabold`} style={{ padding: headPad, color: 'var(--text)' }}>Who to follow</h2>

      <div>
        {people.map((person) => (
          <div
            key={person.userId}
            role="link"
            tabIndex={0}
            onClick={() => router.push(`/profile/${person.userId}`)}
            onKeyDown={(e) => { if (e.key === 'Enter') router.push(`/profile/${person.userId}`); }}
            className="flex cursor-pointer items-center gap-2.5 transition-colors"
            style={{ padding: '12px 16px' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-hover)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <UserHoverCard userId={person.userId} style={{ flexShrink: 0 }}>
              <Avatar username={person.username} avatarUrl={person.userProfilePic || null} size={44} />
            </UserHoverCard>

            <UserHoverCard userId={person.userId} style={{ flex: 1, minWidth: 0 }}>
              <div className="min-w-0 w-full">
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
            </UserHoverCard>

            <FollowButton userId={person.userId} initialFollowing={person.isFollowing} />
          </div>
        ))}
      </div>

      <button
        type="button"
        onClick={() => router.push('/connect')}
        className="w-full text-left text-[15px] transition-colors"
        style={{ padding: '12px 16px', color: 'var(--accent)' }}
        onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-hover)'; }}
        onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
      >
        Show more
      </button>
    </section>
  );
}
