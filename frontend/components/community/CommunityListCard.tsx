'use client';

import { useRouter } from 'next/navigation';
import type { CommunityData } from '@/lib/services/communities';

interface CommunityListCardProps {
  community: CommunityData;
  showDivider?: boolean;
  badge?: string;
}

export function CommunityListCard({ community, showDivider = true, badge }: CommunityListCardProps) {
  const router = useRouter();

  return (
    <button
      onClick={() => router.push(`/communities/${community.community_id}`)}
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
      onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
      onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
    >
      <div
        style={{
          width: 56,
          height: 56,
          borderRadius: 10,
          background: community.banner_url
            ? `url(${community.banner_url}) center/cover no-repeat`
            : 'var(--bg-subtle)',
          border: '1px solid var(--border)',
          flexShrink: 0,
        }}
      />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
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
            {community.name}
          </span>
          {community.visibility === 'PRIVATE' && (
            <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>🔒</span>
          )}
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-tertiary)', marginTop: 2 }}>
          {community.member_count.toLocaleString()} members
          {badge && (
            <span style={{ marginLeft: 8, fontSize: 11, fontWeight: 600, color: 'var(--accent)' }}>
              {badge}
            </span>
          )}
        </div>
      </div>
    </button>
  );
}
