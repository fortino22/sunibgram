'use client';

import { useState, useEffect, useRef } from 'react';
import { ArrowLeft } from 'lucide-react';
import { CommunityMemberCard } from '@/components/community/CommunityMemberCard';
import { useCommunityUsers } from '@/hooks/useCommunities';
import { useKickMember, useUpdateMemberRole } from '@/hooks/useCommunityMutations';
import type { CommunityRole } from '@/lib/services/communities';
import { useAuthStore } from '@/stores/authStore';

interface CommunityMembersPanelProps {
  communityId: number;
  communityName: string;
  myRole: CommunityRole | null;
  onClose: () => void;
}

type RoleFilter = 'all' | 'MEMBER' | 'ADMIN';

const ROLE_TABS: { id: RoleFilter; label: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'MEMBER', label: 'Members' },
  { id: 'ADMIN', label: 'Admins' },
];

export function CommunityMembersPanel({ communityId, communityName, myRole, onClose }: CommunityMembersPanelProps) {
  const currentUserId = useAuthStore((s) => s.user?.id);
  const [roleFilter, setRoleFilter] = useState<RoleFilter>('all');
  const [inputValue, setInputValue] = useState('');
  const [query, setQuery] = useState('');
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const kickMutation = useKickMember(communityId);
  const roleMutation = useUpdateMemberRole(communityId);

  const { data, isLoading } = useCommunityUsers(communityId, {
    role: roleFilter === 'all' ? undefined : roleFilter,
    query: query || undefined,
  });

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => setQuery(inputValue.trim()), 300);
    return () => { if (debounceRef.current) clearTimeout(debounceRef.current); };
  }, [inputValue]);

  // Map CommunityUserData → CommunityMemberData shape that CommunityMemberCard expects
  const members = (data?.data ?? []).map((u) => ({
    community_id: communityId,
    user_id: u.user_id,
    username: u.username,
    avatar_url: u.user_avatar,
    role: u.role,
    joined_at: '',
  }));

  return (
    <div style={{ display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          padding: '14px 16px',
          borderBottom: '1px solid var(--border)',
          position: 'sticky',
          top: 0,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          zIndex: 10,
        }}
      >
        <button
          onClick={onClose}
          className="btn-ghost"
          style={{
            border: 'none',
            cursor: 'pointer',
            color: 'var(--text)',
            display: 'flex',
            alignItems: 'center',
            padding: 4,
            flexShrink: 0,
          }}
        >
          <ArrowLeft size={22} />
        </button>
        <div>
          <div style={{ fontWeight: 700, fontSize: 17, color: 'var(--text)', lineHeight: 1.2 }}>Members</div>
          <div style={{ fontSize: 13, color: 'var(--text-tertiary)', lineHeight: 1.2 }}>{communityName}</div>
        </div>
      </div>

      {/* Search */}
      <div style={{ padding: '12px 16px 0', flexShrink: 0 }}>
        <input
          type="text"
          placeholder="Search by name or username..."
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          style={{
            width: '100%',
            padding: '10px 14px',
            borderRadius: 10,
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
          flexShrink: 0,
          overflowX: 'auto',
          scrollbarWidth: 'none',
        }}
      >
        {ROLE_TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setRoleFilter(tab.id)}
            className="hover-dim"
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
      <div>
        {isLoading && (
          <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
            Loading...
          </div>
        )}
        {!isLoading && members.length === 0 && (
          <div style={{ padding: '40px 20px', textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
            {query ? `No members matching "${query}"` : 'No members found'}
          </div>
        )}
        {members.map((m, i) => (
          <CommunityMemberCard
            key={m.user_id}
            member={m}
            myRole={myRole}
            myUserId={currentUserId ?? 0}
            onKick={(userId) => kickMutation.mutate(userId)}
            onRoleChange={(userId, role) => roleMutation.mutate({ userId, role })}
            kickPending={kickMutation.isPending}
            rolePending={roleMutation.isPending}
            showDivider={i < members.length - 1}
          />
        ))}
      </div>
    </div>
  );
}
