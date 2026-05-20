'use client';

import type { CommunityData, CommunityRole } from '@/lib/services/communities';

const AVATAR_COLORS = [
  '#e57373', '#f06292', '#ba68c8', '#7986cb',
  '#4fc3f7', '#4db6ac', '#81c784', '#ffb74d',
];

function memberColor(userId: string): string {
  let hash = 0;
  for (let i = 0; i < userId.length; i++) hash = userId.charCodeAt(i) + ((hash << 5) - hash);
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
}

interface CommunityHeaderProps {
  community: CommunityData;
  myRole: CommunityRole | null;
  hasPendingRequest: boolean;
  joinPending: boolean;
  leavePending: boolean;
  requestPending: boolean;
  cancelPending: boolean;
  onJoin: () => void;
  onLeave: () => void;
  onRequest: () => void;
  onCancelRequest: () => void;
  onEdit: () => void;
  onDelete: () => void;
}

export function CommunityHeader({
  community,
  myRole,
  hasPendingRequest,
  joinPending,
  leavePending,
  requestPending,
  cancelPending,
  onJoin,
  onLeave,
  onRequest,
  onCancelRequest,
  onEdit,
  onDelete,
}: CommunityHeaderProps) {
  const isOwner = myRole === 'OWNER';
  const isMember = myRole !== null;
  const isPublic = community.visibility === 'PUBLIC';

  function renderActionButton() {
    if (isOwner) {
      return (
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            onClick={onEdit}
            style={{
              padding: '7px 18px',
              borderRadius: 999,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
              fontSize: 14,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Edit
          </button>
          <button
            onClick={onDelete}
            style={{
              padding: '7px 18px',
              borderRadius: 999,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--error, #e53e3e)',
              fontSize: 14,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Delete
          </button>
        </div>
      );
    }

    if (isMember) {
      return (
        <button
          onClick={onLeave}
          disabled={leavePending}
          style={{
            padding: '7px 18px',
            borderRadius: 999,
            border: '1px solid var(--border)',
            background: 'transparent',
            color: 'var(--text)',
            fontSize: 14,
            fontWeight: 600,
            cursor: leavePending ? 'default' : 'pointer',
            opacity: leavePending ? 0.6 : 1,
          }}
        >
          {leavePending ? 'Leaving…' : 'Joined'}
        </button>
      );
    }

    if (!isPublic) {
      if (hasPendingRequest) {
        return (
          <button
            onClick={onCancelRequest}
            disabled={cancelPending}
            style={{
              padding: '7px 18px',
              borderRadius: 999,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text-secondary)',
              fontSize: 14,
              fontWeight: 600,
              cursor: cancelPending ? 'default' : 'pointer',
              opacity: cancelPending ? 0.6 : 1,
            }}
          >
            {cancelPending ? 'Cancelling…' : 'Requested'}
          </button>
        );
      }
      return (
        <button
          onClick={onRequest}
          disabled={requestPending}
          style={{
            padding: '7px 18px',
            borderRadius: 999,
            border: 'none',
            background: 'var(--text)',
            color: 'var(--bg)',
            fontSize: 14,
            fontWeight: 700,
            cursor: requestPending ? 'default' : 'pointer',
            opacity: requestPending ? 0.6 : 1,
          }}
        >
          {requestPending ? 'Requesting…' : 'Request to Join'}
        </button>
      );
    }

    return (
      <button
        onClick={onJoin}
        disabled={joinPending}
        style={{
          padding: '7px 18px',
          borderRadius: 999,
          border: 'none',
          background: 'var(--text)',
          color: 'var(--bg)',
          fontSize: 14,
          fontWeight: 700,
          cursor: joinPending ? 'default' : 'pointer',
          opacity: joinPending ? 0.6 : 1,
        }}
      >
        {joinPending ? 'Joining…' : 'Join'}
      </button>
    );
  }

  return (
    <div>
      {/* Banner */}
      <div
        style={{
          height: 180,
          background: community.banner_url
            ? `url(${community.banner_url}) center/cover no-repeat`
            : 'var(--bg-subtle)',
        }}
      />

      {/* Content */}
      <div style={{ padding: '14px 20px 16px' }}>
        <div style={{ fontSize: 22, fontWeight: 800, color: 'var(--text)', letterSpacing: '-0.02em', marginBottom: 6 }}>
          {community.name}
        </div>

        {community.description && (
          <p style={{ fontSize: 14, color: 'var(--text-secondary)', lineHeight: 1.5, margin: '0 0 12px' }}>
            {community.description}
          </p>
        )}

        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            {/* Stacked member avatars */}
            {(community.top_members ?? []).length > 0 && (
              <div style={{ display: 'flex' }}>
                {(community.top_members ?? []).slice(0, 3).map((m, i) => (
                  <div
                    key={m.user_id}
                    style={{
                      width: 24,
                      height: 24,
                      borderRadius: '50%',
                      border: '2px solid var(--bg)',
                      marginLeft: i === 0 ? 0 : -8,
                      zIndex: (community.top_members ?? []).length - i,
                      position: 'relative',
                      overflow: 'hidden',
                      flexShrink: 0,
                      background: m.avatar_url
                        ? `url(${m.avatar_url}) center/cover no-repeat`
                        : memberColor(m.user_id),
                    }}
                  />
                ))}
              </div>
            )}
            <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>
              <span style={{ fontWeight: 700, color: 'var(--text)' }}>
                {community.member_count.toLocaleString()}
              </span>
              {' Members'}
            </span>
            {community.visibility === 'PRIVATE' && (
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>· 🔒</span>
            )}
            {myRole && myRole !== 'MEMBER' && (
              <span
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: myRole === 'OWNER' || myRole === 'ADMIN' ? 'var(--accent)' : 'var(--text-secondary)',
                }}
              >
                · {myRole.charAt(0) + myRole.slice(1).toLowerCase()}
              </span>
            )}
          </div>
          {renderActionButton()}
        </div>
      </div>
    </div>
  );
}
