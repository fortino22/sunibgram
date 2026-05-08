'use client';

import { useState } from 'react';
import type { QueryKey } from '@tanstack/react-query';
import { ProfilePageHeader } from '@/components/profile/ProfilePageHeader';
import { ProfileBanner } from '@/components/profile/ProfileBanner';
import { ProfileInfo } from '@/components/profile/ProfileInfo';
import { ProfileTabs, type ProfileTab } from '@/components/profile/ProfileTabs';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import { ProfileEditModal } from '@/components/profile/ProfileEditModal';
import { ProfileSkeleton } from '@/components/profile/skeletons/ProfileSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useMyThreads } from '@/hooks/useMyThreads';
import { useMyReplies } from '@/hooks/useMyReplies';
import { useMyLikes } from '@/hooks/useMyLikes';
import { useThreadActions } from '@/hooks/useThreadActions';
import type { ThreadItem } from '@/lib/services/threads';

// ─── Connected card ───────────────────────────────────────────────────────────

interface ConnectedCardProps {
  thread: ThreadItem;
  queryKey: QueryKey;
  replyingTo?: string;
}

function ConnectedCard({ thread, queryKey, replyingTo }: ConnectedCardProps) {
  const { like, repost, bookmark } = useThreadActions(thread.id, queryKey);
  return (
    <ProfileThreadCard
      thread={thread}
      replyingTo={replyingTo}
      onLike={() => like.mutate(!!thread.is_liked)}
      onRepost={() => repost.mutate(!!thread.is_reposted)}
      onBookmark={() => bookmark.mutate(!!thread.is_bookmarked)}
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
      />
      <ProfileTabs activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'threads' && (
        threads.length === 0
          ? <EmptyState title="No threads yet" subtitle="Your posts will appear here." />
          : threads.map((t) => <ConnectedCard key={t.id} thread={t} queryKey={['threads', 'me']} />)
      )}

      {activeTab === 'replies' && (
        replies.length === 0
          ? <EmptyState title="No replies yet" subtitle="Your replies to others will appear here." />
          : replies.map((t) => (
            <ConnectedCard
              key={t.id}
              thread={t}
              queryKey={['threads', 'me', 'replies']}
              replyingTo={t.parent_context?.author.username}
            />
          ))
      )}

      {activeTab === 'media' && (
        mediaThreads.length === 0
          ? <EmptyState title="No media yet" subtitle="Photos and videos from threads will appear here." />
          : mediaThreads.map((t) => <ConnectedCard key={t.id} thread={t} queryKey={['threads', 'me']} />)
      )}

      {activeTab === 'likes' && (
        likes.length === 0
          ? <EmptyState title="No likes yet" subtitle="Posts you've liked will appear here." />
          : likes.map((t) => <ConnectedCard key={t.id} thread={t} queryKey={['likes', 'me']} />)
      )}

      {showEditModal && (
        <ProfileEditModal profile={profile} onClose={() => setShowEditModal(false)} />
      )}
    </div>
  );
}
