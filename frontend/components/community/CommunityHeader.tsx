'use client';

import type { CommunityData, CommunityRole } from '@/lib/services/communities';

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
        <div style={{ fontSize: 22, fontWeight: 800, color: 'var(--text)', letterSpacing: '-0.02em', marginBottom: 8 }}>
          {community.name}
        </div>

        {community.category_name && (
          <div style={{ marginBottom: 10 }}>
            <span
              style={{
                display: 'inline-block',
                padding: '3px 10px',
                borderRadius: 999,
                border: '1px solid var(--border)',
                fontSize: 12,
                fontWeight: 600,
                color: 'var(--text-secondary)',
              }}
            >
              {community.category_name}
            </span>
          </div>
        )}

        {community.description && (
          <p style={{ fontSize: 14, color: 'var(--text)', lineHeight: 1.5, margin: '0 0 14px' }}>
            {community.description}
          </p>
        )}

        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)' }}>
              {community.member_count.toLocaleString()}
            </span>
            <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>Members</span>
            {community.visibility === 'PRIVATE' && (
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>· 🔒</span>
            )}
            {myRole && myRole !== 'MEMBER' && (
              <span
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: myRole === 'OWNER' || myRole === 'ADMIN' ? 'var(--accent)' : 'var(--text-secondary)',
                  marginLeft: 4,
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
