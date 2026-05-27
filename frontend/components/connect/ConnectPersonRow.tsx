'use client';

import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { FollowButton } from '@/components/shared/FollowButton';
import type { ExplorePerson } from '@/lib/services/explore';

interface ConnectPersonRowProps {
  person: ExplorePerson;
}

export function ConnectPersonRow({ person }: ConnectPersonRowProps) {
  const router = useRouter();
  const goToProfile = () => router.push(`/profile/${person.userId}`);

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={goToProfile}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); goToProfile(); } }}
      className="flex cursor-pointer items-start gap-3 transition-colors"
      style={{ padding: '12px 16px', borderBottom: '1px solid var(--border)' }}
      onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)'; }}
      onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
    >
      <Avatar username={person.username} avatarUrl={person.userProfilePic || null} size="md" />

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
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

          <FollowButton userId={person.userId} initialFollowing={person.isFollowing} />
        </div>

        {person.bio && (
          <p className="line-clamp-2 text-[14px]" style={{ marginTop: 4, color: 'var(--text)' }}>
            {person.bio}
          </p>
        )}
      </div>
    </div>
  );
}
