'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ProfilePageHeader } from '@/components/profile/ProfilePageHeader';
import { ProfileBanner } from '@/components/profile/ProfileBanner';
import { ProfileInfo } from '@/components/profile/ProfileInfo';
import { ReportModal } from '@/components/shared/ReportModal';
import { ProfileTabs, type ProfileTab } from '@/components/profile/ProfileTabs';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import { ProfileSkeleton } from '@/components/profile/skeletons/ProfileSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { MediaGrid } from '@/components/shared/MediaGrid';
import { useProfile } from '@/hooks/useProfile';
import { useUserThreads } from '@/hooks/useUserThreads';
import { useUserReplies } from '@/hooks/useUserReplies';
import { useFollowUser } from '@/hooks/useFollowUser';
import { useBlockUser } from '@/hooks/useBlockUser';
import { useMuteUser } from '@/hooks/useMuteUser';
import { useAuthStore } from '@/stores/authStore';

export default function UserProfilePage() {
  const params = useParams<{ userId: string }>();
  const router = useRouter();
  const routeUserId = Number(params.userId);
  const hasValidUserId = Number.isInteger(routeUserId) && routeUserId > 0;
  const profileUserId = hasValidUserId ? routeUserId : undefined;
  const currentUserId = useAuthStore((state) => state.user?.id);
  const [activeTab, setActiveTab] = useState<ProfileTab>('threads');
  const [reportOpen, setReportOpen] = useState(false);

  const { data: profile, isLoading, isError, error } = useProfile(profileUserId);
  const blockedRelationship = !!profile?.is_blocked || !!profile?.is_blocked_by;
  const threadsQuery = useUserThreads(profileUserId, { enabled: !blockedRelationship });
  const repliesQuery = useUserReplies(profileUserId, { enabled: !blockedRelationship });
  const { follow, unfollow } = useFollowUser(profileUserId, [profileUserId]);
  const { block, unblock } = useBlockUser(profileUserId, [profileUserId]);
  const { mute, unmute } = useMuteUser(profileUserId, [profileUserId]);

  const threadsSentinelRef = useRef<HTMLDivElement>(null);
  const repliesSentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const sentinel = threadsSentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && threadsQuery.hasNextPage && !threadsQuery.isFetchingNextPage) {
          void threadsQuery.fetchNextPage();
        }
      },
      { rootMargin: '200px' },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [threadsQuery]);

  useEffect(() => {
    const sentinel = repliesSentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && repliesQuery.hasNextPage && !repliesQuery.isFetchingNextPage) {
          void repliesQuery.fetchNextPage();
        }
      },
      { rootMargin: '200px' },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [repliesQuery]);

  if (!hasValidUserId) {
    return <EmptyState title="Profile unavailable" subtitle="Invalid profile id." />;
  }

  if (isLoading) return <ProfileSkeleton />;

  if (isError) {
    return <EmptyState title="Profile unavailable" subtitle={error.message} />;
  }

  if (!profile) {
    return <EmptyState title="Profile unavailable" subtitle="This profile could not be loaded." />;
  }

  const isOwn = currentUserId === profile.id;
  const isFollowing = !!profile.is_following;
  const isBlocked = !!profile.is_blocked;
  const isBlockedBy = !!profile.is_blocked_by;
  const isMuted = !!profile.is_muted;
  const threads = threadsQuery.data?.pages.flatMap((p) => p.data) ?? [];
  const replies = repliesQuery.data?.pages.flatMap((p) => p.data) ?? [];
  const mediaThreads = threads.filter((t) => t.media.length > 0);

  const joinedMonth = profile.created_at
    ? new Date(profile.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
    : undefined;

  function handleFollowToggle() {
    if (isFollowing) {
      unfollow.mutate();
    } else {
      follow.mutate();
    }
  }

  function handleBlockToggle() {
    if (isBlocked) {
      unblock.mutate();
    } else {
      block.mutate();
    }
  }

  function handleMuteToggle() {
    if (isMuted) {
      unmute.mutate();
    } else {
      mute.mutate();
    }
  }

  return (
    <div>
      <ProfilePageHeader name={profile.name ?? profile.username} threadCount={threads.length} isPremium={profile.is_premium} />
      <ProfileBanner bannerUrl={profile.banner ?? null} />
      <ProfileInfo
        user={{
          name: profile.name ?? profile.username,
          username: profile.username,
          bio: profile.bio,
          faculty: profile.faculty?.faculty_name ?? null,
          joinedMonth,
          following: profile.following_count ?? 0,
          followers: profile.followers_count ?? 0,
          avatarUrl: profile.avatar ?? null,
          isPremium: profile.is_premium,
        }}
        userId={profile.id}
        isOwn={isOwn}
        isFollowing={isFollowing}
        followPending={follow.isPending || unfollow.isPending}
        onFollowToggle={handleFollowToggle}
        onFollowersClick={() => router.push(`/profile/${profile.id}/followers?tab=followers`)}
        onFollowingClick={() => router.push(`/profile/${profile.id}/followers?tab=following`)}
        onReport={isOwn ? undefined : () => setReportOpen(true)}
        isBlocked={isBlocked}
        isBlockedBy={isBlockedBy}
        onBlock={isOwn || isBlockedBy ? undefined : handleBlockToggle}
        blockPending={block.isPending || unblock.isPending}
        isMuted={isMuted}
        onMute={isOwn || isBlockedBy ? undefined : handleMuteToggle}
      />

      {reportOpen && (
        <ReportModal target={{ type: 'user', userId: profile.id }} onClose={() => setReportOpen(false)} />
      )}
      {isBlocked || isBlockedBy ? (
        <>
          <div
            style={{
              padding: '16px 20px',
              borderTop: '1px solid var(--border)',
              borderBottom: '1px solid var(--border)',
              fontSize: 15,
              fontWeight: 700,
              color: 'var(--text)',
            }}
          >
            {isBlocked ? `You blocked @${profile.username}` : `@${profile.username} blocked you`}
          </div>
          <EmptyState
            title={isBlocked ? 'You blocked this account' : 'You’re blocked'}
            subtitle={isBlocked ? 'Unblock to see their posts.' : `You can’t see @${profile.username}’s posts.`}
          />
        </>
      ) : (
        <>
      <ProfileTabs activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'threads' && (
        threads.length === 0
          ? <EmptyState title="No threads yet" subtitle="Posts will appear here." />
          : <>
              {threads.map((t) => <ProfileThreadCard key={t.id} thread={t} />)}
              <div ref={threadsSentinelRef} style={{ height: 1 }} />
            </>
      )}

      {activeTab === 'replies' && (
        replies.length === 0
          ? <EmptyState title="No replies yet" subtitle="Replies to others will appear here." />
          : <>
              {replies.map((t) => (
                <ProfileThreadCard
                  key={t.id}
                  thread={t}
                  replyingTo={t.parent_context?.author.username}
                  replyingToId={t.parent_context?.author.id}
                />
              ))}
              <div ref={repliesSentinelRef} style={{ height: 1 }} />
            </>
      )}

      {activeTab === 'media' && (
        mediaThreads.length === 0
          ? <EmptyState title="No media yet" subtitle="Photos and videos will appear here." />
          : <MediaGrid threads={mediaThreads} />
      )}

      {activeTab === 'likes' && (
        <EmptyState title="Likes are private" subtitle="Only the account owner can see their likes." />
      )}
        </>
      )}
    </div>
  );
}
