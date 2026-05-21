'use client';

import type { CommunityData, CommunityRole } from '@/lib/services/communities';
import { Avatar } from '@/components/shared/Avatar';

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

        {community.visibility === 'PRIVATE' && (
          <span style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 4,
            fontSize: 12,
            fontWeight: 600,
            color: '#ef4444',
            border: '1.5px solid #ef4444',
            borderRadius: 999,
            padding: '2px 10px',
            marginBottom: 8,
          }}>
            Private
          </span>
        )}

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
                  <Avatar
                    key={m.user_id}
                    username={m.user_id}
                    avatarUrl={m.avatar_url}
                    size={24}
                    style={{
                      border: '2px solid var(--bg)',
                      marginLeft: i === 0 ? 0 : -8,
                      zIndex: (community.top_members ?? []).length - i,
                      position: 'relative',
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
