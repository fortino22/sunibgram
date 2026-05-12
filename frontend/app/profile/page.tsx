'use client';

import { useState } from 'react';
import { ProfilePageHeader } from '@/components/profile/ProfilePageHeader';
import { ProfileBanner } from '@/components/profile/ProfileBanner';
import { ProfileInfo } from '@/components/profile/ProfileInfo';
import { ProfileTabs, type ProfileTab } from '@/components/profile/ProfileTabs';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import { ProfileEditModal } from '@/components/profile/ProfileEditModal';
import { FollowListModal } from '@/components/profile/FollowListModal';
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
  const [followModal, setFollowModal] = useState<'followers' | 'following' | null>(null);

  const { data: profile, isLoading } = useMyProfile();
  const threadsQuery = useMyThreads();
  const repliesQuery = useMyReplies();
  const likesQuery = useMyLikes();

  if (isLoading || !profile) return <ProfileSkeleton />;

  const threads = threadsQuery.data?.data ?? [];
  const replies = repliesQuery.data?.data ?? [];
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
        onEditClick={() => setShowEditModal(true)}
        onFollowersClick={() => setFollowModal('followers')}
        onFollowingClick={() => setFollowModal('following')}
      />
      <ProfileTabs activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'threads' && (
        threads.length === 0
          ? <EmptyState title="No threads yet" subtitle="Your posts will appear here." />
          : threads.map((t) => <ConnectedCard key={t.id} thread={t} />)
      )}

      {activeTab === 'replies' && (
        replies.length === 0
          ? <EmptyState title="No replies yet" subtitle="Your replies to others will appear here." />
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
          ? <EmptyState title="No media yet" subtitle="Photos and videos from threads will appear here." />
          : <MediaGrid threads={mediaThreads} />
      )}

      {activeTab === 'likes' && (
        likes.length === 0
          ? <EmptyState title="No likes yet" subtitle="Posts you've liked will appear here." />
          : likes.map((t) => <LikeCard key={t.id} thread={t} />)
      )}

      {showEditModal && (
        <ProfileEditModal profile={profile} onClose={() => setShowEditModal(false)} />
      )}
      {followModal && (
        <FollowListModal userId={profile.id} type={followModal} onClose={() => setFollowModal(null)} />
      )}
    </div>
  );
}
