'use client';

import { useState } from 'react';
import { useParams } from 'next/navigation';
import { ProfilePageHeader } from '@/components/profile/ProfilePageHeader';
import { ProfileBanner } from '@/components/profile/ProfileBanner';
import { ProfileInfo } from '@/components/profile/ProfileInfo';
import { ProfileTabs, type ProfileTab } from '@/components/profile/ProfileTabs';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import { FollowListModal } from '@/components/profile/FollowListModal';
import { ProfileSkeleton } from '@/components/profile/skeletons/ProfileSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { useProfile } from '@/hooks/useProfile';
import { useUserThreads } from '@/hooks/useUserThreads';
import { useUserReplies } from '@/hooks/useUserReplies';
import { useFollowUser } from '@/hooks/useFollowUser';
import { useThreadActions } from '@/hooks/useThreadActions';
import { useAuthStore } from '@/stores/authStore';
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

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function UserProfilePage() {
  const { userId } = useParams<{ userId: string }>();
  const currentUserId = useAuthStore((s) => s.user?.id);

  const [activeTab, setActiveTab] = useState<ProfileTab>('threads');
  const [followModal, setFollowModal] = useState<'followers' | 'following' | null>(null);

  const { data: profile, isLoading } = useProfile(userId);
  const threadsQuery = useUserThreads(userId);
  const repliesQuery = useUserReplies(userId);
  const { follow, unfollow } = useFollowUser(userId);

  if (isLoading || !profile) return <ProfileSkeleton />;

  const isOwn = currentUserId === profile.id;
  const threads = threadsQuery.data?.data ?? [];
  const replies = repliesQuery.data?.data ?? [];
  const mediaThreads = threads.filter((t) => t.media.length > 0);

  const joinedMonth = profile.created_at
    ? new Date(profile.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
    : undefined;

  function handleFollowToggle() {
    if (profile?.is_following) {
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
        isFollowing={!!profile.is_following}
        onFollowToggle={handleFollowToggle}
        onFollowersClick={() => setFollowModal('followers')}
        onFollowingClick={() => setFollowModal('following')}
      />
      <ProfileTabs activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'threads' && (
        threads.length === 0
          ? <EmptyState title="No threads yet" subtitle="Posts will appear here." />
          : threads.map((t) => <ConnectedCard key={t.id} thread={t} />)
      )}

      {activeTab === 'replies' && (
        replies.length === 0
          ? <EmptyState title="No replies yet" subtitle="Replies to others will appear here." />
          : replies.map((t) => (
            <ConnectedCard
              key={t.id}
              thread={t}
              replyingTo={t.parent_context?.author.username}
              replyingToId={t.parent_context?.author.id}
            />
          ))
      )}

      {activeTab === 'media' && (
        mediaThreads.length === 0
          ? <EmptyState title="No media yet" subtitle="Photos and videos will appear here." />
          : mediaThreads.map((t) => <ConnectedCard key={t.id} thread={t} />)
      )}

      {activeTab === 'likes' && (
        <EmptyState title="Likes are private" subtitle="Only the account owner can see their likes." />
      )}

      {followModal && (
        <FollowListModal userId={profile.id} type={followModal} onClose={() => setFollowModal(null)} />
      )}
    </div>
  );
}
