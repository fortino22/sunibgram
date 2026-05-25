'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft, ChevronDown } from 'lucide-react';
import Link from 'next/link';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useCommunity, useCommunityUsers } from '@/hooks/useCommunities';
import { useUpdateMemberRole } from '@/hooks/useCommunityMutations';
import { useAuthStore } from '@/stores/authStore';
import { EmptyState } from '@/components/shared/EmptyState';
import type { CommunityRole } from '@/lib/services/communities';

type RoleFilter = 'all' | 'MEMBER' | 'ADMIN';

const ROLE_TABS: { id: RoleFilter; label: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'MEMBER', label: 'Members' },
  { id: 'ADMIN', label: 'Admins' },
];

const ROLE_LABEL: Record<CommunityRole, string> = {
  MEMBER: 'Member',
  ADMIN: 'Admin',
};

const ROLE_COLOR: Record<CommunityRole, string> = {
  MEMBER: 'var(--text-tertiary)',
  ADMIN: 'var(--coral)',
};

export default function CommunityMembersPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const communityId = Number(params.id);
  const isValidId = Number.isInteger(communityId) && communityId > 0;

  const currentUserId = useAuthStore((s) => s.user?.id);
  const [roleFilter, setRoleFilter] = useState<RoleFilter>('all');
  const [inputValue, setInputValue] = useState('');
  const [query, setQuery] = useState('');
  const [openMenuUserId, setOpenMenuUserId] = useState<number | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const communityQuery = useCommunity(isValidId ? communityId : undefined);
  const { data, isLoading } = useCommunityUsers(
    isValidId ? communityId : undefined,
    { role: roleFilter === 'all' ? undefined : roleFilter, query: query || undefined }
  );
  const updateRole = useUpdateMemberRole(communityId);

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => setQuery(inputValue.trim()), 300);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [inputValue]);

  if (!isValidId) {
    return <EmptyState title="Community not found" subtitle="Invalid community ID." />;
  }

  const community = communityQuery.data;
  const members = data?.data ?? [];

  const isOwner = currentUserId != null && community?.user_id === currentUserId;
  const myRole: CommunityRole | null = members.find((m) => m.user_id === currentUserId)?.role ?? null;
  const canManageRoles = myRole === 'ADMIN';

  function canManage(memberRole: CommunityRole, memberId: number): boolean {
    if (memberId === currentUserId) return false;
    // Creator can manage all including other admins; other admins can only manage members
    if (isOwner) return true;
    return myRole === 'ADMIN' && memberRole !== 'ADMIN';
  }

  function getTargetRoles(currentRole: CommunityRole): CommunityRole[] {
    const options: CommunityRole[] = ['MEMBER', 'ADMIN'];
    return options.filter((r) => r !== currentRole);
  }

  function handleRoleChange(userId: number, role: CommunityRole) {
    updateRole.mutate({ userId, role });
    setOpenMenuUserId(null);
  }

  const isUpdating = updateRole.isPending;

  return (
    <div
      style={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}
      onClick={() => setOpenMenuUserId(null)}
    >
      {/* Top bar */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '14px 16px',
          display: 'flex',
          alignItems: 'center',
          gap: 12,
        }}
      >
        <button
          onClick={() => router.back()}
          style={{
            background: 'none',
            border: 'none',
            color: 'var(--text)',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            padding: 4,
            flexShrink: 0,
          }}
        >
          <ArrowLeft size={22} />
        </button>
        <div>
          <div
            style={{
              fontWeight: 700,
              fontSize: 17,
              color: 'var(--text)',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            Members
          </div>
          {community && (
            <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
              {community.name}
            </div>
          )}
        </div>
      </div>

      {/* Search */}
      <div style={{ padding: '12px 16px 0' }}>
        <input
          type="text"
          placeholder="Search members..."
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          style={{
            width: '100%',
            padding: '10px 14px',
            borderRadius: 999,
            border: '1px solid var(--border)',
            background: 'var(--bg-subtle)',
            color: 'var(--text)',
            fontSize: 14,
            outline: 'none',
            boxSizing: 'border-box',
          }}
        />
      </div>

      {/* Role filter tabs */}
      <div
        style={{
          display: 'flex',
          gap: 8,
          padding: '10px 16px',
          borderBottom: '1px solid var(--border)',
          overflowX: 'auto',
          scrollbarWidth: 'none',
        }}
      >
        {ROLE_TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setRoleFilter(tab.id)}
            style={{
              padding: '5px 14px',
              borderRadius: 999,
              border: `1px solid ${roleFilter === tab.id ? 'var(--text)' : 'var(--border)'}`,
              background: roleFilter === tab.id ? 'var(--text)' : 'transparent',
              color: roleFilter === tab.id ? 'var(--bg)' : 'var(--text)',
              fontSize: 13,
              fontWeight: 600,
              cursor: 'pointer',
              whiteSpace: 'nowrap',
              flexShrink: 0,
            }}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Member list */}
      <div style={{ flex: 1 }}>
        {isLoading && (
          <div
            style={{
              padding: '40px 0',
              textAlign: 'center',
              color: 'var(--text-tertiary)',
              fontSize: 14,
            }}
          >
            Loading…
          </div>
        )}

        {!isLoading && members.length === 0 && (
          <EmptyState
            title={query ? `No results for "${query}"` : 'No members found'}
            subtitle={query ? 'Try a different search term.' : ''}
          />
        )}

        {members.map((member, i) => {
          const isMenuOpen = openMenuUserId === member.user_id;
          const isThisMemberUpdating = isUpdating && updateRole.variables?.userId === member.user_id;

          return (
            <div key={member.user_id}>
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <Link
                  href={`/profile/${member.user_id}`}
                  style={{
                    display: 'flex',
                    gap: 12,
                    alignItems: 'center',
                    textDecoration: 'none',
                    padding: '12px 16px',
                    flex: 1,
                    minWidth: 0,
                  }}
                >
                  <Avatar
                    username={member.username}
                    avatarUrl={member.user_avatar || null}
                    size={44}
                  />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 5, flexWrap: 'wrap' }}>
                      <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)' }}>
                        {member.name || member.username}
                      </span>
                      {member.is_premium && <VerifiedBadge size={15} />}
                    </div>
                    <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
                      @{member.username}
                    </div>
                  </div>
                </Link>

                {/* Role badge + manage button */}
                <div
                  style={{ display: 'flex', alignItems: 'center', paddingRight: 16, flexShrink: 0 }}
                  onClick={(e) => e.stopPropagation()}
                >
                  {canManageRoles && canManage(member.role, member.user_id) ? (
                    <div style={{ position: 'relative' }}>
                      <button
                        disabled={isThisMemberUpdating}
                        onClick={() =>
                          setOpenMenuUserId(isMenuOpen ? null : member.user_id)
                        }
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 4,
                          padding: '4px 10px',
                          borderRadius: 999,
                          border: `1px solid ${ROLE_COLOR[member.role]}`,
                          background: 'transparent',
                          color: ROLE_COLOR[member.role],
                          fontSize: 12,
                          fontWeight: 600,
                          cursor: isThisMemberUpdating ? 'not-allowed' : 'pointer',
                          opacity: isThisMemberUpdating ? 0.5 : 1,
                          transition: 'opacity 0.15s',
                        }}
                      >
                        {isThisMemberUpdating ? 'Saving…' : ROLE_LABEL[member.role]}
                        {!isThisMemberUpdating && <ChevronDown size={12} />}
                      </button>

                      {isMenuOpen && (
                        <div
                          style={{
                            position: 'absolute',
                            right: 0,
                            top: 'calc(100% + 6px)',
                            background: 'var(--bg-card)',
                            border: '1px solid var(--border)',
                            borderRadius: 10,
                            overflow: 'hidden',
                            zIndex: 50,
                            minWidth: 140,
                            boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
                          }}
                        >
                          {getTargetRoles(member.role).map((role) => (
                            <button
                              key={role}
                              onClick={() => handleRoleChange(member.user_id, role)}
                              style={{
                                display: 'block',
                                width: '100%',
                                padding: '10px 14px',
                                textAlign: 'left',
                                background: 'none',
                                border: 'none',
                                color: ROLE_COLOR[role],
                                fontSize: 13,
                                fontWeight: 600,
                                cursor: 'pointer',
                              }}
                              onMouseEnter={(e) => {
                                (e.currentTarget as HTMLButtonElement).style.background = 'var(--bg-subtle)';
                              }}
                              onMouseLeave={(e) => {
                                (e.currentTarget as HTMLButtonElement).style.background = 'none';
                              }}
                            >
                              Make {ROLE_LABEL[role]}
                            </button>
                          ))}
                        </div>
                      )}
                    </div>
                  ) : member.role !== 'MEMBER' ? (
                    <span
                      style={{
                        fontSize: 12,
                        fontWeight: 600,
                        color: ROLE_COLOR[member.role],
                        padding: '4px 10px',
                        borderRadius: 999,
                        border: `1px solid ${ROLE_COLOR[member.role]}`,
                      }}
                    >
                      {ROLE_LABEL[member.role]}
                    </span>
                  ) : null}
                </div>
              </div>
              {i < members.length - 1 && (
                <div style={{ height: 1, background: 'var(--border)', margin: '0 16px' }} />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
