'use client';

import { useRouter } from 'next/navigation';
import type { ExploreCommunity } from '@/lib/services/explore';

interface CommunityCardProps {
  community: ExploreCommunity;
  showDivider?: boolean;
}

export function CommunityCard({ community, showDivider = true }: CommunityCardProps) {
  const router = useRouter();

  return (
    <button
      onClick={() => router.push(`/communities/${community.communityId}`)}
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
      <div
        style={{
          width: 48,
          height: 48,
          borderRadius: 10,
          background: community.banner
            ? `url(${community.banner}) center/cover no-repeat`
            : 'var(--bg-subtle)',
          border: '1px solid var(--border)',
          flexShrink: 0,
        }}
      />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            fontSize: 15,
            fontWeight: 700,
            color: 'var(--text)',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {community.communityName}
        </div>
        <div style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>
          {community.memberCount.toLocaleString()} members
          {community.isJoin && (
            <span
              style={{
                marginLeft: 8,
                fontSize: 12,
                fontWeight: 600,
                color: 'var(--accent)',
              }}
            >
              Joined
            </span>
          )}
        </div>
      </div>
    </button>
  );
}
