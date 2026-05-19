import { Avatar } from '@/components/shared/Avatar';
import type { DiscoverCommunityData } from '@/lib/services/communities';

interface CommunityDiscoverCardProps {
  community: DiscoverCommunityData;
  showDivider?: boolean;
}

export function CommunityDiscoverCard({ community, showDivider = true }: CommunityDiscoverCardProps) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '14px 20px',
        borderBottom: showDivider ? '1px solid var(--border)' : 'none',
      }}
    >
      {/* Top member avatars stacked */}
      <div style={{ position: 'relative', width: 40, height: 40, flexShrink: 0 }}>
        {community.top_members.slice(0, 3).map((m, i) => (
          <div
            key={i}
            style={{
              position: 'absolute',
              width: 24,
              height: 24,
              borderRadius: '50%',
              border: '2px solid var(--bg)',
              overflow: 'hidden',
              top: i === 0 ? 0 : i === 1 ? 8 : 16,
              left: i === 0 ? 0 : i === 1 ? 8 : 16,
              zIndex: 3 - i,
            }}
          >
            <Avatar username="member" avatarUrl={m.avatar_url || null} size={20} />
          </div>
        ))}
        {community.top_members.length === 0 && (
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: '50%',
              background: 'var(--bg-subtle)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 18,
            }}
          >
            👥
          </div>
        )}
      </div>

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
          {community.name}
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-tertiary)', marginTop: 2 }}>
          {community.member_count.toLocaleString()} members
          {community.category_name && (
            <span
              style={{
                marginLeft: 8,
                padding: '1px 6px',
                borderRadius: 999,
                background: 'var(--bg-subtle)',
                fontSize: 11,
                fontWeight: 500,
                color: 'var(--text-secondary)',
              }}
            >
              {community.category_name}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}
