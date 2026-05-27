'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import type { CommunityMemberData, CommunityRole } from '@/lib/services/communities';

const ROLE_COLORS: Record<CommunityRole, string> = {
  ADMIN: 'var(--accent)',
  MEMBER: 'var(--text-tertiary)',
};

interface CommunityMemberCardProps {
  member: CommunityMemberData;
  myRole: CommunityRole | null;
  myUserId: number;
  onKick: (userId: number) => void;
  onRoleChange: (userId: number, role: CommunityRole) => void;
  kickPending: boolean;
  rolePending: boolean;
  showDivider?: boolean;
}

export function CommunityMemberCard({
  member,
  myRole,
  myUserId,
  onKick,
  onRoleChange,
  kickPending,
  rolePending,
  showDivider = true,
}: CommunityMemberCardProps) {
  const router = useRouter();
  const [showActions, setShowActions] = useState(false);
  const isSelf = member.user_id === myUserId;

  const canKick = !isSelf && myRole === 'ADMIN';

  const canChangeRole = !isSelf && myRole === 'ADMIN';

  const showMenu = canKick || canChangeRole;

  const promotableRoles: CommunityRole[] = (['MEMBER', 'ADMIN'] as CommunityRole[]).filter((r) => r !== member.role);

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '12px 20px',
        borderBottom: showDivider ? '1px solid var(--border)' : 'none',
        position: 'relative',
      }}
    >
      <button
        onClick={() => router.push(`/profile/${member.user_id}`)}
        className="btn-ghost"
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          flex: 1,
          border: 'none',
          cursor: 'pointer',
          textAlign: 'left',
        }}
      >
        <Avatar username={member.username} avatarUrl={member.avatar_url || null} size="md" />
        <div>
          <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)' }}>
            @{member.username}
          </div>
          <div
            style={{
              fontSize: 12,
              fontWeight: 600,
              color: ROLE_COLORS[member.role] ?? 'var(--text-tertiary)',
            }}
          >
            {member.role.charAt(0) + member.role.slice(1).toLowerCase()}
          </div>
        </div>
      </button>

      {showMenu && (
        <div style={{ position: 'relative' }}>
          <button
            onClick={() => setShowActions((v) => !v)}
            onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-hover)'; e.currentTarget.style.color = 'var(--text)'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--text-secondary)'; }}
            style={{
              width: 28,
              height: 28,
              borderRadius: '50%',
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text-secondary)',
              cursor: 'pointer',
              fontSize: 16,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'background 0.12s ease, color 0.12s ease',
            }}
          >
            ⋯
          </button>

          {showActions && (
            <div
              style={{
                position: 'absolute',
                right: 0,
                top: 34,
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 12,
                minWidth: 160,
                boxShadow: '0 4px 20px rgba(0,0,0,0.3)',
                zIndex: 20,
                overflow: 'hidden',
              }}
            >
              {canChangeRole &&
                promotableRoles.map((role) => (
                  <button
                    key={role}
                    disabled={rolePending}
                    onClick={() => {
                      onRoleChange(member.user_id, role);
                      setShowActions(false);
                    }}
                    style={{
                      display: 'block',
                      width: '100%',
                      padding: '10px 16px',
                      background: 'transparent',
                      border: 'none',
                      textAlign: 'left',
                      fontSize: 14,
                      color: 'var(--text)',
                      cursor: rolePending ? 'default' : 'pointer',
                    }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
                  >
                    Set as {role.charAt(0) + role.slice(1).toLowerCase()}
                  </button>
                ))}
              {canKick && (
                <button
                  disabled={kickPending}
                  onClick={() => {
                    onKick(member.user_id);
                    setShowActions(false);
                  }}
                  style={{
                    display: 'block',
                    width: '100%',
                    padding: '10px 16px',
                    background: 'transparent',
                    border: 'none',
                    textAlign: 'left',
                    fontSize: 14,
                    color: 'var(--error, #e53e3e)',
                    cursor: kickPending ? 'default' : 'pointer',
                  }}
                  onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
                  onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
                >
                  Remove member
                </button>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
