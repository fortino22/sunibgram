'use client';

import { useState } from 'react';
import { CreatePostBox } from '@/components/shared/CreatePostBox';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { Icon } from '@/components/ui/Icon';
import { useFollowingThreads } from '@/hooks/useFollowingThreads';
import { useThreads } from '@/hooks/useThreads';
import { useFacultyThreads } from '@/hooks/useFacultyThreads';
import { usePinnedCommunities, useCommunityThreadsLatest } from '@/hooks/useCommunities';
import { useMyProfile } from '@/hooks/useMyProfile';
import type { ThreadItem } from '@/lib/services/threads';

interface Tab {
  id: string;
  label: string;
}

const EMPTY_MESSAGES: Record<string, { title: string; subtitle: string }> = {
  'for-you':   { title: 'Nothing here yet',                subtitle: 'Posts from campus will appear here.' },
  'following': { title: 'No posts from people you follow', subtitle: 'Follow more people to see their posts here.' },
  'faculty':   { title: 'No faculty posts yet',            subtitle: 'Posts from your faculty will appear here.' },
};

interface FeedTabButtonProps {
  tab: Tab;
  active: boolean;
  onClick: () => void;
}

function FeedTabButton({ tab, active, onClick }: FeedTabButtonProps) {
  return (
    <button
      onClick={onClick}
      style={{
        flex: 1,
        minWidth: 'max-content',
        padding: '12px 8px',
        background: 'none',
        border: 'none',
        cursor: 'pointer',
        fontWeight: active ? 600 : 400,
        fontSize: 14,
        color: active ? 'var(--text)' : 'var(--text-tertiary)',
        borderBottom: active ? '2px solid var(--accent)' : '2px solid transparent',
        marginBottom: -1,
        transition: 'all 0.15s ease',
        letterSpacing: '-0.01em',
        fontFamily: 'inherit',
        whiteSpace: 'nowrap',
        textAlign: 'center',
      }}
    >
      {tab.label}
    </button>
  );
}

function EmptyFeed({ emptyKey }: { emptyKey: string }) {
  const msg = EMPTY_MESSAGES[emptyKey] ?? { title: 'Nothing here yet', subtitle: 'Posts will appear here.' };
  return (
    <div
      style={{
        padding: '80px 24px',
        textAlign: 'center',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 12,
      }}
    >
      <div
        style={{
          width: 52,
          height: 52,
          borderRadius: '50%',
          background: 'var(--bg-subtle)',
          border: '1px solid var(--border)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          marginBottom: 4,
        }}
      >
        <Icon name="reply" size={22} color="var(--border)" strokeWidth={1.2} />
      </div>
      <div style={{ fontWeight: 600, fontSize: 18, color: 'var(--text)', letterSpacing: '-0.02em' }}>
        {msg.title}
      </div>
      <div style={{ fontSize: 14, color: 'var(--text-tertiary)', maxWidth: 280, lineHeight: 1.6 }}>
        {msg.subtitle}
      </div>
    </div>
  );
}

function ActiveCommunityFeed({ communityId }: { communityId: number }) {
  const query = useCommunityThreadsLatest(communityId);
  const threads = query.data?.data ?? [];

  if (query.isLoading) {
    return <div>{Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)}</div>;
  }
  if (threads.length === 0) {
    return <EmptyFeed emptyKey="community" />;
  }
  return (
    <div>
      {threads.map((thread: ThreadItem) => (
        <ThreadCard key={thread.id} thread={thread} />
      ))}
    </div>
  );
}

export default function HomePage() {
  const [activeTab, setActiveTab] = useState<string>('for-you');

  const { data: profile } = useMyProfile();
  const facultyId = profile?.faculty?.faculty_id ?? null;
  const facultyName = profile?.faculty?.faculty_name ?? 'Faculty';

  const isForYou = activeTab === 'for-you';
  const isFollowing = activeTab === 'following';
  const isFaculty = activeTab === 'faculty';
  const activeCommunityId = activeTab.startsWith('community-')
    ? Number(activeTab.replace('community-', ''))
    : null;

  const threadsQuery = useThreads();
  const followingQuery = useFollowingThreads(1, 10, isFollowing);
  const facultyQuery = useFacultyThreads(facultyId, 1, 10, isFaculty);
  const pinnedCommunitiesQuery = usePinnedCommunities();

  const activeQuery = isFollowing
    ? followingQuery
    : isFaculty
    ? facultyQuery
    : threadsQuery;

  const activeThreads = activeQuery.data?.data ?? [];
  const pinnedCommunities = pinnedCommunitiesQuery.data?.data ?? [];

  const allTabs: Tab[] = [
    { id: 'for-you', label: 'For You' },
    { id: 'following', label: 'Following' },
    { id: 'faculty', label: facultyName },
    ...pinnedCommunities.map((c) => ({
      id: `community-${c.community_id}`,
      label: c.community_name,
    })),
  ];

  return (
    <div>
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div style={{ padding: '20px 20px 0' }}>
          <div
            style={{
              fontWeight: 600,
              fontSize: 20,
              color: 'var(--text)',
              letterSpacing: '-0.02em',
              marginBottom: 10,
            }}
          >
            Home
          </div>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              overflowX: 'auto',
              scrollbarWidth: 'none',
            }}
          >
            {allTabs.map((t) => (
              <FeedTabButton
                key={t.id}
                tab={t}
                active={activeTab === t.id}
                onClick={() => setActiveTab(t.id)}
              />
            ))}
          </div>
        </div>
      </div>

      {(isForYou || isFollowing) && <CreatePostBox />}

      {activeCommunityId ? (
        <ActiveCommunityFeed communityId={activeCommunityId} />
      ) : activeQuery.isLoading ? (
        <div>
          {Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)}
        </div>
      ) : activeThreads.length > 0 ? (
        <div>
          {activeThreads.map((thread: ThreadItem) => (
            <ThreadCard key={thread.id} thread={thread} />
          ))}
        </div>
      ) : (
        <EmptyFeed emptyKey={activeTab} />
      )}
    </div>
  );
}
