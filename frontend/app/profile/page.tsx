'use client';

import { useState, useEffect, useRef } from 'react';
import { ProfilePageHeader } from '@/components/profile/ProfilePageHeader';
import { ProfileBanner } from '@/components/profile/ProfileBanner';
import { FollowListTab } from '@/components/profile/FollowListTab';
import { ProfileInfo } from '@/components/profile/ProfileInfo';
import { ProfileTabs, type ProfileTab } from '@/components/profile/ProfileTabs';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import { ProfileEditModal } from '@/components/profile/ProfileEditModal';
import { ProfileSkeleton } from '@/components/profile/skeletons/ProfileSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { MediaGrid } from '@/components/shared/MediaGrid';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useMyThreads } from '@/hooks/useMyThreads';
import { useMyReplies } from '@/hooks/useMyReplies';
import { useMyLikes } from '@/hooks/useMyLikes';
import { useProfile } from '@/hooks/useProfile';
import { useThreadActions } from '@/hooks/useThreadActions';
import type { ThreadItem } from '@/lib/services/threads';

// ─── Connected card ───────────────────────────────────────────────────────────

interface ConnectedCardProps {
  thread: ThreadItem;
  replyingTo?: string;
  replyingToId?: number;
}

function ConnectedCard({ thread, replyingTo, replyingToId }: ConnectedCardProps) {
  const { like, repost, bookmark } = useThreadActions(thread);

  return (
    <ProfileThreadCard
      thread={thread}
      replyingTo={replyingTo}
      replyingToId={replyingToId}
      onLike={() => {
        if (!like.isPending) {
          like.mutate(thread.is_liked === true);
        }
      }}
      onRepost={() => {
        if (!repost.isPending) {
          repost.mutate(thread.is_reposted === true);
        }
      }}
      onBookmark={() => {
        if (!bookmark.isPending) {
          bookmark.mutate(thread.is_bookmarked === true);
        }
      }}
    />
  );
}

// ─── Like card — resolves avatar via profile API since likes endpoint omits it ─

function LikeCard({ thread }: { thread: ThreadItem }) {
  const { data: authorProfile } = useProfile(thread.author.id);
  const { like, repost, bookmark } = useThreadActions(thread);
  const patchedThread: ThreadItem = authorProfile?.avatar
    ? { ...thread, author: { ...thread.author, avatar_url: authorProfile.avatar } }
    : thread;
  return (
    <ProfileThreadCard
      thread={patchedThread}
      onLike={() => { if (!like.isPending) like.mutate(thread.is_liked === true); }}
      onRepost={() => { if (!repost.isPending) repost.mutate(thread.is_reposted === true); }}
      onBookmark={() => { if (!bookmark.isPending) bookmark.mutate(thread.is_bookmarked === true); }}
    />
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function ProfilePage() {
  const [activeTab, setActiveTab] = useState<ProfileTab>('threads');
  const [showEditModal, setShowEditModal] = useState(false);

  const { data: profile, isLoading } = useMyProfile();
  const threadsQuery = useMyThreads();
  const repliesQuery = useMyReplies();
  const likesQuery = useMyLikes();

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

  if (isLoading || !profile) return <ProfileSkeleton />;

  const threads = threadsQuery.data?.pages.flatMap((p) => p.data) ?? [];
  const replies = repliesQuery.data?.pages.flatMap((p) => p.data) ?? [];
  const likes = likesQuery.data?.data ?? [];
  const mediaThreads = threads.filter((t) => t.media.length > 0);

  const joinedMonth = profile.created_at
    ? new Date(profile.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
    : undefined;

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
        isOwn
        isFollowing={false}
        onFollowToggle={() => {}}
        followPending={false}
        onEditClick={() => setShowEditModal(true)}
        onFollowersClick={() => setActiveTab('followers')}
        onFollowingClick={() => setActiveTab('following')}
      />
      <ProfileTabs activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'threads' && (
        threads.length === 0
          ? <EmptyState title="No threads yet" subtitle="Your posts will appear here." />
          : <>
              {threads.map((t) => <ConnectedCard key={t.id} thread={t} />)}
              <div ref={threadsSentinelRef} style={{ height: 1 }} />
            </>
      )}

      {activeTab === 'replies' && (
        replies.length === 0
          ? <EmptyState title="No replies yet" subtitle="Your replies to others will appear here." />
          : <>
              {replies.map((t) => (
                <ConnectedCard
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
          ? <EmptyState title="No media yet" subtitle="Photos and videos from threads will appear here." />
          : <MediaGrid threads={mediaThreads} />
      )}

      {activeTab === 'likes' && (
        likes.length === 0
          ? <EmptyState title="No likes yet" subtitle="Posts you've liked will appear here." />
          : likes.map((t) => <LikeCard key={t.id} thread={t} />)
      )}

      {activeTab === 'followers' && <FollowListTab userId={profile.id} type="followers" />}

      {activeTab === 'following' && <FollowListTab userId={profile.id} type="following" />}

      {showEditModal && (
        <ProfileEditModal profile={profile} onClose={() => setShowEditModal(false)} />
      )}
    </div>
  );
}
