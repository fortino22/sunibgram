'use client';

import { useState } from 'react';
import { EventsFeedPanel } from '@/components/events/EventsFeedPanel';
import { CreatePostBox } from '@/components/shared/CreatePostBox';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { Icon } from '@/components/ui/Icon';
import { useFollowingThreads } from '@/hooks/useFollowingThreads';
import { useThreads } from '@/hooks/useThreads';
import type { ThreadItem } from '@/lib/services/threads';

type FeedTabId = 'for-you' | 'following' | 'cs' | 'events';

interface Tab {
  id: FeedTabId;
  label: string;
}

const TABS: Tab[] = [
  { id: 'for-you',   label: 'For You' },
  { id: 'following', label: 'Following' },
  { id: 'cs',        label: 'Comp Sci' },
  { id: 'events',    label: 'Events' },
];

const EMPTY_MESSAGES: Record<FeedTabId, { title: string; subtitle: string }> = {
  'for-you':   { title: 'Nothing here yet',                subtitle: 'Posts from campus will appear here.' },
  'following': { title: 'No posts from people you follow', subtitle: 'Follow more people to see their posts here.' },
  'cs':        { title: 'No Comp Sci posts yet',           subtitle: 'Posts from Computer Science students will appear here.' },
  'events':    { title: 'No events or announcements yet',  subtitle: 'Campus events and announcements will appear here.' },
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
        padding: '12px 0',
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
      }}
    >
      {tab.label}
    </button>
  );
}

function EmptyFeed({ tab }: { tab: FeedTabId }) {
  const { title, subtitle } = EMPTY_MESSAGES[tab];
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
      <div
        style={{
          fontWeight: 600,
          fontSize: 18,
          color: 'var(--text)',
          letterSpacing: '-0.02em',
        }}
      >
        {title}
      </div>
      <div
        style={{
          fontSize: 14,
          color: 'var(--text-tertiary)',
          maxWidth: 280,
          lineHeight: 1.6,
        }}
      >
        {subtitle}
      </div>
    </div>
  );
}

export default function HomePage() {
  const [activeTab, setActiveTab] = useState<FeedTabId>('for-you');
  const threadsQuery = useThreads();
  const followingThreadsQuery = useFollowingThreads(1, 10, activeTab === 'following');
  const isForYouTab = activeTab === 'for-you';
  const isFollowingTab = activeTab === 'following';
  const isEventsTab = activeTab === 'events';
  const activeQuery = isFollowingTab ? followingThreadsQuery : threadsQuery;
  const activeThreads = activeQuery.data?.data ?? [];

  return (
    <div>
      {/* Sticky header with tabs */}
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
          <div style={{ display: 'flex' }}>
            {TABS.map((t) => (
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

      {!isEventsTab && <CreatePostBox />}

      {isEventsTab ? (
        <EventsFeedPanel active={isEventsTab} />
      ) : (isForYouTab || isFollowingTab) && activeQuery.isLoading ? (
        <div>
          {Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)}
        </div>
      ) : (isForYouTab || isFollowingTab) && activeThreads.length > 0 ? (
        <div>
          {activeThreads.map((thread: ThreadItem) => (
            <ThreadCard key={thread.id} thread={thread} />
          ))}
        </div>
      ) : (
        <EmptyFeed tab={activeTab} />
      )}
    </div>
  );
}
