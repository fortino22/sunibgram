'use client';

import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { UserHoverCard } from '@/components/shared/UserHoverCard';
import type { ExplorePerson } from '@/lib/services/explore';

interface PersonCardProps {
  person: ExplorePerson;
  showDivider?: boolean;
}

export function PersonCard({ person, showDivider = true }: PersonCardProps) {
  const router = useRouter();

  return (
    <button
      onClick={() => router.push(`/profile/${person.userId}`)}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        width: '100%',
        padding: '14px 20px',
        background: 'transparent',
        border: 'none',
        borderBottom: showDivider ? '1px solid var(--border)' : 'none',
        cursor: 'pointer',
        textAlign: 'left',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = 'var(--bg-subtle)';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = 'transparent';
      }}
    >
      <UserHoverCard userId={person.userId} style={{ flexShrink: 0 }}>
        <Avatar
          username={person.username}
          avatarUrl={person.userProfilePic || null}
          size="md"
        />
      </UserHoverCard>
      <UserHoverCard userId={person.userId} style={{ flex: 1, minWidth: 0 }}>
        <div style={{ width: '100%', minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <span
              style={{
                fontSize: 15,
                fontWeight: 700,
                color: 'var(--text)',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {person.name}
            </span>
            {person.isPremium && <VerifiedBadge size={15} />}
          </div>
          <div
            style={{
              fontSize: 14,
              color: 'var(--text-tertiary)',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            @{person.username}
          </div>
        </div>
      </UserHoverCard>
    </button>
  );
}
