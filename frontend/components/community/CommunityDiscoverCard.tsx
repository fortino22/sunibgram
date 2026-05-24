'use client';

import type { DiscoverCommunityData } from '@/lib/services/communities';
import { Avatar } from '@/components/shared/Avatar';

interface CommunityDiscoverCardProps {
  community: DiscoverCommunityData;
  onClick: () => void;
}

function formatCount(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

const MONOGRAM_COLORS = [
  '#e57373', '#f06292', '#ba68c8', '#7986cb',
  '#4fc3f7', '#4db6ac', '#81c784', '#ffb74d',
];

function monogramColor(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash);
  return MONOGRAM_COLORS[Math.abs(hash) % MONOGRAM_COLORS.length];
}

export function CommunityDiscoverCard({ community, onClick }: CommunityDiscoverCardProps) {
  const color = monogramColor(community.name);
  const initial = community.name.charAt(0).toUpperCase();
  const thumbUrl = community.banner_url ?? null;

  return (
    <div
      onClick={onClick}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 14,
        padding: '12px 16px',
        cursor: 'pointer',
        transition: 'background 0.12s',
      }}
      onMouseEnter={(e) => { (e.currentTarget as HTMLDivElement).style.background = 'var(--bg-subtle)'; }}
      onMouseLeave={(e) => { (e.currentTarget as HTMLDivElement).style.background = 'transparent'; }}
    >
      {/* Thumbnail */}
      <div
        style={{
          width: 108,
          height: 108,
          borderRadius: 20,
          flexShrink: 0,
          overflow: 'hidden',
          background: thumbUrl ? `url(${thumbUrl}) center/cover no-repeat` : color,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: 40,
          fontWeight: 800,
          color: '#fff',
          letterSpacing: '-0.02em',
          userSelect: 'none',
        }}
      >
        {!thumbUrl && initial}
      </div>

      {/* Info */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            fontSize: 17,
            fontWeight: 700,
            color: 'var(--text)',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            marginBottom: 2,
          }}
        >
          {community.name}
        </div>

        <div style={{ fontSize: 14, color: 'var(--text-secondary)', marginBottom: 2 }}>
          <span style={{ fontWeight: 600, color: 'var(--text)' }}>
            {formatCount(community.member_count)}
          </span>
          {' Members'}
        </div>

        {community.category_name && (
          <div style={{ fontSize: 13, color: 'var(--text-tertiary)', marginBottom: 6 }}>
            {community.category_name}
          </div>
        )}

        {community.top_members.length > 0 && (
          <div style={{ display: 'flex', alignItems: 'center' }}>
            {community.top_members.slice(0, 5).map((m, i) => (
              <Avatar
                key={i}
                username={m.user_id}
                avatarUrl={m.avatar_url || null}
                size={29}
                style={{
                  border: '2px solid var(--bg-card)',
                  marginLeft: i > 0 ? -9 : 0,
                  zIndex: 5 - i,
                  position: 'relative',
                }}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
