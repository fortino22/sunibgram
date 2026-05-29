'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ProfilePageHeader } from '@/components/profile/ProfilePageHeader';
import { ProfileBanner } from '@/components/profile/ProfileBanner';
import { ProfileInfo } from '@/components/profile/ProfileInfo';
import { ProfileTabs, type ProfileTab } from '@/components/profile/ProfileTabs';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import { ProfileSkeleton } from '@/components/profile/skeletons/ProfileSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { MediaGrid } from '@/components/shared/MediaGrid';
import { useProfile } from '@/hooks/useProfile';
import { useUserThreads } from '@/hooks/useUserThreads';
import { useUserReplies } from '@/hooks/useUserReplies';
import { useFollowUser } from '@/hooks/useFollowUser';
import { useAuthStore } from '@/stores/authStore';

export default function UserProfilePage() {
  const params = useParams<{ userId: string }>();
  const router = useRouter();
  const routeUserId = Number(params.userId);
  const hasValidUserId = Number.isInteger(routeUserId) && routeUserId > 0;
  const profileUserId = hasValidUserId ? routeUserId : undefined;
  const currentUserId = useAuthStore((state) => state.user?.id);
  const [activeTab, setActiveTab] = useState<ProfileTab>('threads');

  const { data: profile, isLoading, isError, error } = useProfile(profileUserId);
  const threadsQuery = useUserThreads(profileUserId);
  const repliesQuery = useUserReplies(profileUserId);
  const { follow, unfollow } = useFollowUser(profileUserId, [profileUserId]);

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

  return (
    <div>
      <ProfilePageHeader name={profile.name ?? profile.username} threadCount={threads.length} />
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
        }}
        isOwn={isOwn}
        isFollowing={isFollowing}
        followPending={follow.isPending || unfollow.isPending}
        onFollowToggle={handleFollowToggle}
        onFollowersClick={() => router.push(`/profile/${profile.id}/followers?tab=followers`)}
        onFollowingClick={() => router.push(`/profile/${profile.id}/followers?tab=following`)}
      />
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
    </div>
  );
}
