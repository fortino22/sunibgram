'use client';

import { useState } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { ArrowLeft, Search } from 'lucide-react';
import { CommunityDetailSkeleton } from '@/components/community/CommunityDetailSkeleton';
import { CommunityHeader } from '@/components/community/CommunityHeader';
import { CommunityAboutTab } from '@/components/community/CommunityAboutTab';
import { CommunityMembersPanel } from '@/components/community/CommunityMembersPanel';
import { CommunityRequestCard } from '@/components/community/CommunityRequestCard';
import { CommunityTabBar } from '@/components/community/CommunityTabBar';
import { CommunityEventsTab } from '@/components/community-events/CommunityEventsTab';
import { CreateCommunityEventDialog } from '@/components/community-events/CreateCommunityEventDialog';
import { EditCommunityModal } from '@/components/community/EditCommunityModal';
import { DangerConfirmModal } from '@/components/shared/DangerConfirmModal';
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

type Tab = 'top' | 'latest' | 'events' | 'media' | 'about' | 'requests';

const TAB_VALUES: Tab[] = ['top', 'latest', 'events', 'media', 'about', 'requests'];
function parseTab(v: string | null): Tab {
  return v && (TAB_VALUES as string[]).includes(v) ? (v as Tab) : 'latest';
}

export default function CommunityDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const searchParams = useSearchParams();
  const communityId = Number(params.id);
  const isValidId = Number.isInteger(communityId) && communityId > 0;

  const currentUserId = useAuthStore((s) => s.user?.id);
  // URL (?tab=) is the single source of truth so the tab survives navigating to
  // an event and back, plus browser back/forward — no mirrored state to desync.
  const activeTab = parseTab(searchParams.get('tab'));
  function selectTab(id: Tab) {
    const sp = new URLSearchParams(searchParams.toString());
    sp.set('tab', id);
    router.replace(`/communities/${communityId}?${sp.toString()}`, { scroll: false });
  }
  const [showEdit, setShowEdit] = useState(false);
  const [createEventOpen, setCreateEventOpen] = useState(false);
  const [showMembersPanel, setShowMembersPanel] = useState(false);
  const [hasPendingRequest, setHasPendingRequest] = useState(false);
  const [membershipOverride, setMembershipOverride] = useState<CommunityRole | 'none' | null>(null);
  const [processingRequestId, setProcessingRequestId] = useState<number | null>(null);
  const [confirmLeaveOpen, setConfirmLeaveOpen] = useState(false);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);

  const communityQuery = useCommunity(isValidId ? communityId : undefined);
  const membersQuery = useCommunityMembers(isValidId ? communityId : undefined);
  const isAdminForRequests =
    currentUserId != null &&
    (membersQuery.data?.data ?? []).some(
      (m) => m.user_id === currentUserId && m.role === 'ADMIN',
    );
  const requestsQuery = useCommunityRequests(
    isValidId ? communityId : undefined,
    isAdminForRequests,
  );
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
  const effectiveCommunityRole: CommunityRole | null = myRole;
  void isOwnerByCreation;

  const threadTabs: { id: Tab; label: string }[] = [
    { id: 'top', label: 'Top' },
    { id: 'latest', label: 'Latest' },
    { id: 'events', label: 'Events' },
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

  function handleLeave() {
    setConfirmLeaveOpen(true);
  }

  async function confirmLeave() {
    try {
      await leaveMutation.mutateAsync();
      setMembershipOverride('none');
      setConfirmLeaveOpen(false);
    } catch {
      setConfirmLeaveOpen(false);
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

  function handleDelete() {
    setConfirmDeleteOpen(true);
  }

  async function confirmDelete() {
    try {
      await deleteMutation.mutateAsync(communityId);
      setConfirmDeleteOpen(false);
      router.push('/communities');
    } catch {
      setConfirmDeleteOpen(false);
    }
  }

  if (showMembersPanel) {
    return (
      <div>
        <CommunityMembersPanel
          communityId={communityId}
          communityName={community.name}
          myRole={myRole}
          onClose={() => setShowMembersPanel(false)}
        />
      </div>
    );
  }

  return (
    <div>
      {/* Top bar */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'var(--bg-blur)',
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
          className="btn-ghost"
          style={{ border: 'none', color: 'var(--text)', cursor: 'pointer', display: 'flex', alignItems: 'center', padding: 4, flexShrink: 0 }}
        >
          <ArrowLeft size={22} />
        </button>
        <div style={{ flex: 1, fontWeight: 700, fontSize: 17, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {community.name}
        </div>
        <button
          onClick={() => router.push('/communities/discover')}
          className="btn-ghost"
          style={{ border: 'none', color: 'var(--text)', cursor: 'pointer', display: 'flex', alignItems: 'center', padding: 4, flexShrink: 0 }}
          title="Search communities"
        >
          <Search size={20} />
        </button>
      </div>

      {community.is_active === false && (
        <div
          style={{
            padding: '12px 16px',
            background: 'var(--coral-soft, rgba(239,68,68,0.08))',
            color: 'var(--coral, #e53e3e)',
            fontSize: 13,
            lineHeight: 1.5,
            borderBottom: '1px solid var(--border)',
          }}
        >
          {isOwnerByCreation
            ? 'Your premium has expired — this community is locked. Resubscribe to manage it (edit, post, pin, approve requests, change roles, kick members, or delete).'
            : isAdmin
              ? 'This community is locked because the owner\'s premium has expired. All admin actions are disabled until the owner resubscribes.'
              : 'This community is currently inactive. New posts are paused until the owner resubscribes.'}
        </div>
      )}

      {/* Community header */}
      <CommunityHeader
        community={community}
        myRole={myRole}
        isAdmin={isAdmin}
        isOwner={isOwnerByCreation}
        hasPendingRequest={hasPendingRequest}
        isPinned={isPinned}
        pinPending={pinCommunityMutation.isPending || unpinCommunityMutation.isPending}
        joinPending={joinMutation.isPending}
        leavePending={leaveMutation.isPending}
        requestPending={requestMutation.isPending}
        cancelPending={cancelMutation.isPending}
        deletePending={deleteMutation.isPending}
        onPinToggle={handlePinToggle}
        onJoin={handleJoin}
        onLeave={handleLeave}
        onRequest={handleRequest}
        onCancelRequest={handleCancelRequest}
        onEdit={() => setShowEdit(true)}
        onDelete={handleDelete}
        onCreateEvent={() => { selectTab('events'); setCreateEventOpen(true); }}
        onShowMembers={() => setShowMembersPanel(true)}
      />

      {/* Tab bar — max 5 chips, rest collapse into a "More" dropdown */}
      <CommunityTabBar
        tabs={threadTabs}
        activeId={activeTab}
        onSelect={(id) => selectTab(id as Tab)}
        maxVisible={5}
      />

      {/* Community events */}
      {activeTab === 'events' && (
        <CommunityEventsTab
          communityId={communityId}
          isAdmin={isAdmin}
        />
      )}

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
          {(latestQuery.data?.data ?? []).slice().sort((a, b) => (b.pinned_by_admin ? 1 : 0) - (a.pinned_by_admin ? 1 : 0)).map((t) => (
            <ThreadCard
              key={t.id}
              thread={t}
              communityRole={effectiveCommunityRole}
              onPinThread={(id) => pinThreadMutation.mutate(id)}
              onUnpinThread={(id) => unpinThreadMutation.mutate(id)}
              pinningThread={pinThreadMutation.isPending || unpinThreadMutation.isPending}
              inCommunityContext
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
          {(topQuery.data?.data ?? []).slice().sort((a, b) => (b.pinned_by_admin ? 1 : 0) - (a.pinned_by_admin ? 1 : 0)).map((t) => (
            <ThreadCard
              key={t.id}
              thread={t}
              communityRole={effectiveCommunityRole}
              onPinThread={(id) => pinThreadMutation.mutate(id)}
              onUnpinThread={(id) => unpinThreadMutation.mutate(id)}
              pinningThread={pinThreadMutation.isPending || unpinThreadMutation.isPending}
              inCommunityContext
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

      {isAdmin && (
        <CreateCommunityEventDialog
          open={createEventOpen}
          onOpenChange={setCreateEventOpen}
          communityId={communityId}
        />
      )}

      <DangerConfirmModal
        open={confirmLeaveOpen}
        title="Leave community?"
        description={`You will lose access to posts and discussions in ${community.name}. You can rejoin later.`}
        confirmText={community.name}
        confirmLabel="Leave community"
        busy={leaveMutation.isPending}
        onConfirm={confirmLeave}
        onClose={() => setConfirmLeaveOpen(false)}
      />

      <DangerConfirmModal
        open={confirmDeleteOpen}
        title={`Delete ${community.name}?`}
        description="This permanently removes the community, its posts, and all member data. This action cannot be undone."
        confirmText={community.name}
        confirmLabel="Delete community"
        busy={deleteMutation.isPending}
        onConfirm={confirmDelete}
        onClose={() => setConfirmDeleteOpen(false)}
      />
    </div>
  );
}
