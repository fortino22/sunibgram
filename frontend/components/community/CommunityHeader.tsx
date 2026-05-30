'use client';

import { useState } from 'react';
import { Pencil, Trash2 } from 'lucide-react';
import type { CommunityData, CommunityRole } from '@/lib/services/communities';
import { Avatar } from '@/components/shared/Avatar';

interface CommunityHeaderProps {
  community: CommunityData;
  myRole: CommunityRole | null;
  isAdmin: boolean;
  isOwner: boolean;
  hasPendingRequest: boolean;
  isPinned: boolean;
  pinPending: boolean;
  joinPending: boolean;
  leavePending: boolean;
  requestPending: boolean;
  cancelPending: boolean;
  deletePending: boolean;
  onPinToggle: () => void;
  onJoin: () => void;
  onLeave: () => void;
  onRequest: () => void;
  onCancelRequest: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onShowMembers: () => void;
}

export function CommunityHeader({
  community,
  myRole,
  isAdmin,
  isOwner,
  hasPendingRequest,
  isPinned,
  pinPending,
  joinPending,
  leavePending,
  requestPending,
  cancelPending,
  deletePending,
  onPinToggle,
  onJoin,
  onLeave,
  onRequest,
  onCancelRequest,
  onEdit,
  onDelete,
  onShowMembers,
}: CommunityHeaderProps) {
  void isOwner;
  const [joinedHovered, setJoinedHovered] = useState(false);
  const isMember = myRole !== null;
  const isPublic = community.visibility === 'PUBLIC';

  const adminIconBtn = (extra: React.CSSProperties = {}): React.CSSProperties => ({
    border: '1px solid var(--border)',
    borderRadius: 999,
    width: 36,
    height: 36,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    background: 'transparent',
    color: 'var(--text-tertiary)',
    cursor: 'pointer',
    flexShrink: 0,
    ...extra,
  });

  function renderActionButton() {
    if (isAdmin) {
      return null;
    }

    if (isMember) {
      return (
        <button
          onClick={onLeave}
          disabled={leavePending}
          onMouseEnter={() => setJoinedHovered(true)}
          onMouseLeave={() => setJoinedHovered(false)}
          style={{
            padding: '7px 18px',
            borderRadius: 999,
            border: `1px solid ${joinedHovered && !leavePending ? 'var(--error, #e53e3e)' : 'var(--border)'}`,
            background: 'transparent',
            color: joinedHovered && !leavePending ? 'var(--error, #e53e3e)' : 'var(--text)',
            fontSize: 14,
            fontWeight: 600,
            cursor: leavePending ? 'default' : 'pointer',
            opacity: leavePending ? 0.6 : 1,
            transition: 'color 0.15s, border-color 0.15s',
          }}
        >
          {leavePending ? 'Leaving…' : joinedHovered ? 'Leave' : 'Joined'}
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
          className="btn-solid"
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
        className="btn-solid"
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
          height: 234,
          background: community.banner_url
            ? `url(${community.banner_url}) center/cover no-repeat`
            : 'var(--bg-subtle)',
        }}
      />

      {/* Content */}
      <div style={{ padding: '14px 20px 16px', background: 'var(--bg)' }}>
        <div style={{ fontSize: 28, fontWeight: 800, color: 'var(--text)', letterSpacing: '-0.02em', marginBottom: 8 }}>
          {community.name}
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8, flexWrap: 'wrap' }}>
          {community.category_name && (
            <span style={{
              display: 'inline-flex',
              alignItems: 'center',
              fontSize: 13,
              fontWeight: 600,
              color: 'var(--text)',
              border: '1.5px solid var(--border)',
              borderRadius: 999,
              padding: '2px 12px',
            }}>
              {community.category_name}
            </span>
          )}
        </div>

        {community.description && (
          <p style={{ fontSize: 14, color: 'var(--text-secondary)', lineHeight: 1.5, margin: '0 0 12px' }}>
            {community.description}
          </p>
        )}

        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <button
            onClick={onShowMembers}
            className="btn-ghost"
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              border: 'none',
              cursor: 'pointer',
              padding: 0,
            }}
          >
            {/* Stacked member avatars */}
            {(community.top_members ?? []).length > 0 && (
              <div style={{ display: 'flex' }}>
                {(community.top_members ?? []).slice(0, 3).map((m, i) => (
                  <Avatar
                    key={m.user_id}
                    username={m.user_id}
                    avatarUrl={m.avatar_url}
                    size={38}
                    style={{
                      border: '2px solid var(--bg)',
                      marginLeft: i === 0 ? 0 : -13,
                      zIndex: (community.top_members ?? []).length - i,
                      position: 'relative',
                    }}
                  />
                ))}
              </div>
            )}
            <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>
              <span style={{ fontWeight: 700, color: 'var(--text)' }}>
                {community.member_count >= 1_000_000
                  ? `${(community.member_count / 1_000_000).toFixed(1)}M`
                  : community.member_count >= 1000
                  ? `${(community.member_count / 1000).toFixed(1)}K`
                  : community.member_count.toLocaleString()}
              </span>
              {' Members'}
            </span>
            {myRole === 'ADMIN' && (
              <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--accent)' }}>
                · Admin
              </span>
            )}
          </button>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <button
              onClick={onPinToggle}
              disabled={pinPending}
              title={isPinned ? 'Unpin community' : 'Pin community'}
              className="btn-ghost"
              style={{
                ...adminIconBtn(),
                color: isPinned ? 'var(--accent)' : 'var(--text-tertiary)',
                cursor: pinPending ? 'not-allowed' : 'pointer',
                opacity: pinPending ? 0.5 : 1,
              }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill={isPinned ? 'currentColor' : 'none'} stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <line x1="12" y1="17" x2="12" y2="22" />
                <path d="M5 17h14v-1.76a2 2 0 0 0-1.11-1.79l-1.78-.9A2 2 0 0 1 15 10.76V6h1a2 2 0 0 0 0-4H8a2 2 0 0 0 0 4h1v4.76a2 2 0 0 1-1.11 1.79l-1.78.9A2 2 0 0 0 5 15.24Z" />
              </svg>
            </button>
            {isAdmin && (
              <>
                <button
                  onClick={onDelete}
                  disabled={deletePending}
                  title="Delete community"
                  className="btn-ghost"
                  style={{
                    ...adminIconBtn(),
                    color: 'var(--coral, #e53e3e)',
                    cursor: deletePending ? 'not-allowed' : 'pointer',
                    opacity: deletePending ? 0.5 : 1,
                  }}
                >
                  <Trash2 size={16} />
                </button>
                <button
                  onClick={onEdit}
                  title="Edit community"
                  className="btn-ghost"
                  style={adminIconBtn({ color: 'var(--text)' })}
                >
                  <Pencil size={16} />
                </button>
              </>
            )}
            {renderActionButton()}
          </div>
        </div>
      </div>
    </div>
  );
}
