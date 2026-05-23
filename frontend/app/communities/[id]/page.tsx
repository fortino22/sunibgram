'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft, Search } from 'lucide-react';
import { CommunityDetailSkeleton } from '@/components/community/CommunityDetailSkeleton';
import { CommunityHeader } from '@/components/community/CommunityHeader';
import { CommunityAboutTab } from '@/components/community/CommunityAboutTab';
import { CommunityMembersPanel } from '@/components/community/CommunityMembersPanel';
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
  usePinnedCommunities,
} from '@/hooks/useCommunities';
import {
  useJoinCommunity,
  useLeaveCommunity,
  useRequestToJoin,
  useCancelJoinRequest,
  useApproveRequest,
  useDeclineRequest,
  useDeleteCommunity,
  usePinCommunity,
  useUnpinCommunity,
  usePinCommunityThread,
  useUnpinCommunityThread,
} from '@/hooks/useCommunityMutations';
import { useAuthStore } from '@/stores/authStore';
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
  const [showMembersPanel, setShowMembersPanel] = useState(false);
  const [hasPendingRequest, setHasPendingRequest] = useState(false);
  const [membershipOverride, setMembershipOverride] = useState<CommunityRole | 'none' | null>(null);
  const [processingRequestId, setProcessingRequestId] = useState<number | null>(null);

  const communityQuery = useCommunity(isValidId ? communityId : undefined);
  const membersQuery = useCommunityMembers(isValidId ? communityId : undefined);
  const requestsQuery = useCommunityRequests(isValidId ? communityId : undefined);
  const pinsQuery = usePinnedCommunities();
  const latestQuery = useCommunityThreadsLatest(isValidId && activeTab === 'latest' ? communityId : undefined);
  const topQuery = useCommunityThreadsTop(isValidId && activeTab === 'top' ? communityId : undefined);
  const mediaQuery = useCommunityThreadsMedia(isValidId && activeTab === 'media' ? communityId : undefined);

  const joinMutation = useJoinCommunity(communityId);
  const leaveMutation = useLeaveCommunity(communityId);
  const requestMutation = useRequestToJoin(communityId);
  const cancelMutation = useCancelJoinRequest(communityId);
  const approveMutation = useApproveRequest(communityId);
  const declineMutation = useDeclineRequest(communityId);
  const deleteMutation = useDeleteCommunity();
  const pinCommunityMutation = usePinCommunity(communityId);
  const unpinCommunityMutation = useUnpinCommunity(communityId);
  const pinThreadMutation = usePinCommunityThread(communityId);
  const unpinThreadMutation = useUnpinCommunityThread(communityId);

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
  const requests = requestsQuery.data?.data ?? [];

  const isPinned = pinsQuery.data?.data.some((p) => p.community_id === communityId) ?? false;

  function handlePinToggle() {
    if (isPinned) {
      unpinCommunityMutation.mutate();
    } else {
      pinCommunityMutation.mutate();
    }
  }

  const isOwnerByCreation = currentUserId != null && community.user_id === currentUserId;
  const listRole: CommunityRole | null = currentUserId
    ? (members.find((m) => m.user_id === currentUserId)?.role ?? null)
    : null;
  const myRole: CommunityRole | null =
    membershipOverride === 'none' ? null :
    membershipOverride !== null ? membershipOverride :
    listRole;
  const isAdmin = myRole === 'ADMIN';

  const threadTabs: { id: Tab; label: string }[] = [
    { id: 'top', label: 'Top' },
    { id: 'latest', label: 'Latest' },
    { id: 'media', label: 'Media' },
    { id: 'about', label: 'About' },
    ...(isAdmin ? [{ id: 'requests' as Tab, label: `Requests${requests.length > 0 ? ` (${requests.length})` : ''}` }] : []),
  ];

  async function handleJoin() {
    try {
      await joinMutation.mutateAsync();
      setMembershipOverride('MEMBER');
    } catch (e) {
      const msg = e instanceof Error ? e.message : '';
      if (msg.includes('already a member')) {
        setMembershipOverride('MEMBER');
      }
    }
  }

  async function handleLeave() {
    if (!window.confirm('Leave this community?')) return;
    try {
      await leaveMutation.mutateAsync();
      setMembershipOverride('none');
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
          style={{ background: 'none', border: 'none', color: 'var(--text)', cursor: 'pointer', display: 'flex', alignItems: 'center', padding: 4, flexShrink: 0 }}
        >
          <ArrowLeft size={22} />
        </button>
        <div style={{ flex: 1, fontWeight: 700, fontSize: 17, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {community.name}
        </div>
        <button
          onClick={() => router.push('/communities/discover')}
          style={{ background: 'none', border: 'none', color: 'var(--text)', cursor: 'pointer', display: 'flex', alignItems: 'center', padding: 4, flexShrink: 0 }}
          title="Search communities"
        >
          <Search size={20} />
        </button>
      </div>

      {/* Community header */}
      <CommunityHeader
        community={community}
        myRole={myRole}
        isOwner={isOwnerByCreation}
        hasPendingRequest={hasPendingRequest}
        isPinned={isPinned}
        pinPending={pinCommunityMutation.isPending || unpinCommunityMutation.isPending}
        joinPending={joinMutation.isPending}
        leavePending={leaveMutation.isPending}
        requestPending={requestMutation.isPending}
        cancelPending={cancelMutation.isPending}
        onPinToggle={handlePinToggle}
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
            <ThreadCard
              key={t.id}
              thread={t}
              communityRole={myRole}
              onPinThread={(id) => pinThreadMutation.mutate(id)}
              onUnpinThread={(id) => unpinThreadMutation.mutate(id)}
              pinningThread={pinThreadMutation.isPending || unpinThreadMutation.isPending}
            />
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
            <ThreadCard
              key={t.id}
              thread={t}
              communityRole={myRole}
              onPinThread={(id) => pinThreadMutation.mutate(id)}
              onUnpinThread={(id) => unpinThreadMutation.mutate(id)}
              pinningThread={pinThreadMutation.isPending || unpinThreadMutation.isPending}
            />
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
        <CommunityAboutTab
          communityId={communityId}
          memberCount={community.member_count}
        />
      )}

      {/* Requests tab (mod+) */}
      {activeTab === 'requests' && isAdmin && (
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
              onApprove={async (userId) => {
                setProcessingRequestId(userId);
                try {
                  await approveMutation.mutateAsync(userId);
                } finally {
                  setProcessingRequestId(null);
                }
              }}
              onDecline={async (userId) => {
                setProcessingRequestId(userId);
                try {
                  await declineMutation.mutateAsync(userId);
                } finally {
                  setProcessingRequestId(null);
                }
              }}
              approvePending={processingRequestId === r.user_id && approveMutation.isPending}
              declinePending={processingRequestId === r.user_id && declineMutation.isPending}
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

      {/* Members management panel */}
      {showMembersPanel && (
        <CommunityMembersPanel
          communityId={communityId}
          myRole={myRole}
          onClose={() => setShowMembersPanel(false)}
        />
      )}
    </div>
  );
}
