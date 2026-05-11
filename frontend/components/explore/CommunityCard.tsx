import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
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
      <Avatar
        username={community.communityName}
        avatarUrl={community.communityProfile || null}
        size="md"
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
