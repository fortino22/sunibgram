'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { CommunityDetailSkeleton } from '@/components/community/CommunityDetailSkeleton';
import { CommunityHeader } from '@/components/community/CommunityHeader';
import { CommunityMemberCard } from '@/components/community/CommunityMemberCard';
import { CommunityRequestCard } from '@/components/community/CommunityRequestCard';
import { EditCommunityModal } from '@/components/community/EditCommunityModal';
import { EmptyState } from '@/components/shared/EmptyState';
import { MediaGrid } from '@/components/shared/MediaGrid';
import { ThreadCard } from '@/components/thread/ThreadCard';
import {
  useCommunity,
  useCommunityMembers,
  useCommunityRequests,
  useCommunityThreadsLatest,
  useCommunityThreadsTop,
  useCommunityThreadsMedia,
} from '@/hooks/useCommunities';
import {
  useJoinCommunity,
  useLeaveCommunity,
  useRequestToJoin,
  useCancelJoinRequest,
  useApproveRequest,
  useDeclineRequest,
  useKickMember,
  useUpdateMemberRole,
  useDeleteCommunity,
} from '@/hooks/useCommunityMutations';
import { useAuthStore } from '@/stores/authStore';
import { useHomeStore } from '@/stores/homeStore';
import type { CommunityRole } from '@/lib/services/communities';

type Tab = 'top' | 'latest' | 'media' | 'about' | 'requests';

export default function CommunityDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const communityId = Number(params.id);
  const isValidId = Number.isInteger(communityId) && communityId > 0;

  const currentUserId = useAuthStore((s) => s.user?.id);
  const [activeTab, setActiveTab] = useState<Tab>('latest');
  const [showEdit, setShowEdit] = useState(false);
  const [hasPendingRequest, setHasPendingRequest] = useState(false);

  const { pinnedCommunity, setPinnedCommunity } = useHomeStore();
  const isPinned = pinnedCommunity?.id === communityId;

  const communityQuery = useCommunity(isValidId ? communityId : undefined);
  const membersQuery = useCommunityMembers(isValidId ? communityId : undefined);
  const requestsQuery = useCommunityRequests(isValidId ? communityId : undefined);
  const latestQuery = useCommunityThreadsLatest(isValidId && activeTab === 'latest' ? communityId : undefined);
  const topQuery = useCommunityThreadsTop(isValidId && activeTab === 'top' ? communityId : undefined);
  const mediaQuery = useCommunityThreadsMedia(isValidId && activeTab === 'media' ? communityId : undefined);

  const joinMutation = useJoinCommunity(communityId);
  const leaveMutation = useLeaveCommunity(communityId);
  const requestMutation = useRequestToJoin(communityId);
  const cancelMutation = useCancelJoinRequest(communityId);
  const approveMutation = useApproveRequest(communityId);
  const declineMutation = useDeclineRequest(communityId);
  const kickMutation = useKickMember(communityId);
  const roleMutation = useUpdateMemberRole(communityId);
  const deleteMutation = useDeleteCommunity();

  if (!isValidId) {
    return <EmptyState title="Community not found" subtitle="Invalid community ID." />;
  }

  if (communityQuery.isLoading) {
    return <CommunityDetailSkeleton />;
  }

  if (communityQuery.isError || !communityQuery.data) {
    return <EmptyState title="Community not found" subtitle="This community doesn't exist or was removed." />;
  }

  const community = communityQuery.data;
  const members = membersQuery.data?.data ?? [];

  function handlePinToggle() {
    if (isPinned) {
      setPinnedCommunity(null);
    } else {
      setPinnedCommunity({ id: communityId, name: community.name });
    }
  }
  const requests = requestsQuery.data?.data ?? [];

  // Determine current user's role from members list
  const myMember = currentUserId ? members.find((m) => m.user_id === currentUserId) : undefined;
  const myRole: CommunityRole | null = myMember?.role ?? null;
  const isModOrAbove = myRole === 'MODERATOR' || myRole === 'ADMIN' || myRole === 'OWNER';

  const threadTabs: { id: Tab; label: string }[] = [
    { id: 'top', label: 'Top' },
    { id: 'latest', label: 'Latest' },
    { id: 'media', label: 'Media' },
    { id: 'about', label: 'About' },
    ...(isModOrAbove ? [{ id: 'requests' as Tab, label: `Requests${requests.length > 0 ? ` (${requests.length})` : ''}` }] : []),
  ];

  async function handleJoin() {
    try {
      await joinMutation.mutateAsync();
    } catch (e) {
      const msg = e instanceof Error ? e.message : '';
      if (msg.includes('already a member')) {
        // Already a member — refresh members list
        void membersQuery.refetch();
      }
    }
  }

  async function handleLeave() {
    if (!window.confirm('Leave this community?')) return;
    try {
      await leaveMutation.mutateAsync();
    } catch {
      // noop
    }
  }

  async function handleRequest() {
    try {
      await requestMutation.mutateAsync();
      setHasPendingRequest(true);
    } catch {
      // noop
    }
  }

  async function handleCancelRequest() {
    try {
      await cancelMutation.mutateAsync();
      setHasPendingRequest(false);
    } catch {
      // noop
    }
  }

  async function handleDelete() {
    if (!window.confirm(`Delete "${community.name}"? This cannot be undone.`)) return;
    try {
      await deleteMutation.mutateAsync(communityId);
      router.push('/communities');
    } catch {
      // noop
    }
  }

  return (
    <div>
      {/* Back button */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '20px 20px',
          display: 'flex',
          alignItems: 'center',
          gap: 16,
        }}
      >
        <button
          onClick={() => router.back()}
          style={{ background: 'none', border: 'none', color: 'var(--text)', fontSize: 20, cursor: 'pointer', lineHeight: 1 }}
        >
          ←
        </button>
        <div style={{ fontWeight: 700, fontSize: 17, color: 'var(--text)' }}>{community.name}</div>
        <button
          onClick={handlePinToggle}
          title={isPinned ? 'Unpin from home' : 'Pin to home tab'}
          style={{
            marginLeft: 'auto',
            background: 'none',
            border: '1px solid var(--border)',
            borderRadius: 8,
            padding: '6px 12px',
            color: isPinned ? 'var(--accent)' : 'var(--text-tertiary)',
            fontSize: 13,
            cursor: 'pointer',
            fontFamily: 'inherit',
            fontWeight: isPinned ? 600 : 400,
          }}
        >
          {isPinned ? 'Pinned' : 'Pin to home'}
        </button>
      </div>

      {/* Community header */}
      <CommunityHeader
        community={community}
        myRole={myRole}
        hasPendingRequest={hasPendingRequest}
        joinPending={joinMutation.isPending}
        leavePending={leaveMutation.isPending}
        requestPending={requestMutation.isPending}
        cancelPending={cancelMutation.isPending}
        onJoin={handleJoin}
        onLeave={handleLeave}
        onRequest={handleRequest}
        onCancelRequest={handleCancelRequest}
        onEdit={() => setShowEdit(true)}
        onDelete={handleDelete}
      />

      {/* Tab bar */}
      <div
        style={{
          display: 'flex',
          borderBottom: '1px solid var(--border)',
        }}
      >
        {threadTabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            style={{
              flex: 1,
              padding: '12px 0',
              background: 'none',
              border: 'none',
              borderBottom: activeTab === tab.id ? '2px solid var(--text)' : '2px solid transparent',
              color: activeTab === tab.id ? 'var(--text)' : 'var(--text-tertiary)',
              fontSize: 14,
              fontWeight: activeTab === tab.id ? 700 : 400,
              cursor: 'pointer',
              whiteSpace: 'nowrap',
              textAlign: 'center',
            }}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Latest threads */}
      {activeTab === 'latest' && (
        <>
          {latestQuery.isLoading && (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading…
            </div>
          )}
          {!latestQuery.isLoading && (latestQuery.data?.data ?? []).length === 0 && (
            <EmptyState title="No posts yet" subtitle="Be the first to post in this community." />
          )}
          {(latestQuery.data?.data ?? []).map((t) => (
            <ThreadCard key={t.id} thread={t} />
          ))}
        </>
      )}

      {/* Top threads */}
      {activeTab === 'top' && (
        <>
          {topQuery.isLoading && (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading…
            </div>
          )}
          {!topQuery.isLoading && (topQuery.data?.data ?? []).length === 0 && (
            <EmptyState title="No posts yet" subtitle="Top posts will appear here." />
          )}
          {(topQuery.data?.data ?? []).map((t) => (
            <ThreadCard key={t.id} thread={t} />
          ))}
        </>
      )}

      {/* Media threads */}
      {activeTab === 'media' && (
        <>
          {mediaQuery.isLoading && (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading…
            </div>
          )}
          {!mediaQuery.isLoading && (mediaQuery.data?.data ?? []).length === 0 && (
            <EmptyState title="No media yet" subtitle="Photos and videos will appear here." />
          )}
          {(mediaQuery.data?.data ?? []).length > 0 && (
            <MediaGrid threads={mediaQuery.data?.data ?? []} />
          )}
        </>
      )}

      {/* About tab */}
      {activeTab === 'about' && (
        <>
          {community.rules.length > 0 && (
            <div
              style={{
                margin: '16px 20px 0',
                padding: '12px 16px',
                borderRadius: 10,
                background: 'var(--bg-subtle)',
                border: '1px solid var(--border)',
              }}
            >
              <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--text-secondary)', marginBottom: 8 }}>
                RULES
              </div>
              {community.rules.map((rule, i) => (
                <div key={rule.id} style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.6, marginBottom: i < community.rules.length - 1 ? 6 : 0 }}>
                  {i + 1}. {rule.text}
                </div>
              ))}
            </div>
          )}
          <div style={{ padding: '16px 20px 8px', fontWeight: 700, fontSize: 15, color: 'var(--text)' }}>
            Members ({community.member_count.toLocaleString()})
          </div>
          {membersQuery.isLoading && (
            <div style={{ padding: '20px 0', textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading…
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
        </>
      )}

      {/* Requests tab (mod+) */}
      {activeTab === 'requests' && isModOrAbove && (
        <>
          {requestsQuery.isLoading && (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading…
            </div>
          )}
          {!requestsQuery.isLoading && requests.length === 0 && (
            <EmptyState title="No pending requests" subtitle="All join requests have been handled." />
          )}
          {requests.map((r, i) => (
            <CommunityRequestCard
              key={r.user_id}
              request={r}
              onApprove={(userId) => approveMutation.mutate(userId)}
              onDecline={(userId) => declineMutation.mutate(userId)}
              approvePending={approveMutation.isPending}
              declinePending={declineMutation.isPending}
              showDivider={i < requests.length - 1}
            />
          ))}
        </>
      )}

      {/* Edit modal */}
      {showEdit && (
        <EditCommunityModal
          community={community}
          onClose={() => setShowEdit(false)}
          onSuccess={() => setShowEdit(false)}
        />
      )}


    </div>
  );
}
