'use client';

import { useState } from 'react';
import { CreatePostBox } from '@/components/shared/CreatePostBox';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { Icon } from '@/components/ui/Icon';
import { useFollowingThreads } from '@/hooks/useFollowingThreads';
import { useThreads } from '@/hooks/useThreads';
import { useFacultyThreads } from '@/hooks/useFacultyThreads';
import { useCommunityThreadsLatest } from '@/hooks/useCommunities';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useHomeStore } from '@/stores/homeStore';
import type { ThreadItem } from '@/lib/services/threads';

type FeedTabId = 'for-you' | 'following' | 'dynamic';

interface Tab {
  id: FeedTabId;
  label: string;
}

const STATIC_TABS: Tab[] = [
  { id: 'for-you', label: 'For You' },
  { id: 'following', label: 'Following' },
];

const EMPTY_MESSAGES: Record<string, { title: string; subtitle: string }> = {
  'for-you':   { title: 'Nothing here yet',                subtitle: 'Posts from campus will appear here.' },
  'following': { title: 'No posts from people you follow', subtitle: 'Follow more people to see their posts here.' },
  'faculty':   { title: 'No faculty posts yet',            subtitle: 'Posts from your faculty will appear here.' },
  'community': { title: 'No community posts yet',          subtitle: 'Posts from this community will appear here.' },
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

function EmptyFeed({ emptyKey }: { emptyKey: string }) {
  const { title, subtitle } = EMPTY_MESSAGES[emptyKey] ?? EMPTY_MESSAGES['for-you'];
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

  const { pinnedCommunity, tab3Mode, setTab3Mode } = useHomeStore();
  const { data: profile } = useMyProfile();
  const facultyId = profile?.faculty?.faculty_id ?? null;
  const facultyName = profile?.faculty?.faculty_name ?? 'Faculty';
  const tab3Label = tab3Mode === 'community' && pinnedCommunity ? pinnedCommunity.name : facultyName;

  const isForYou = activeTab === 'for-you';
  const isFollowing = activeTab === 'following';
  const isDynamic = activeTab === 'dynamic';
  const isFacultyMode = tab3Mode === 'faculty' || !pinnedCommunity;
  const isCommunityMode = tab3Mode === 'community' && !!pinnedCommunity;

  const threadsQuery = useThreads();
  const followingQuery = useFollowingThreads(1, 10, isFollowing);
  const facultyQuery = useFacultyThreads(facultyId, 1, 10, isDynamic && isFacultyMode);
  const communityQuery = useCommunityThreadsLatest(
    isDynamic && isCommunityMode ? pinnedCommunity?.id : undefined
  );

  const activeQuery = isFollowing
    ? followingQuery
    : isDynamic && isCommunityMode
    ? communityQuery
    : isDynamic
    ? facultyQuery
    : threadsQuery;

  const activeThreads = activeQuery.data?.data ?? [];
  const emptyKey = isDynamic ? (isCommunityMode ? 'community' : 'faculty') : activeTab;

  const allTabs: Tab[] = [
    ...STATIC_TABS,
    { id: 'dynamic', label: tab3Label },
  ];

  const showModeSwitcher = isDynamic && !!pinnedCommunity;

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
          <div style={{ display: 'flex' }}>
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

        {showModeSwitcher && (
          <div
            style={{
              display: 'flex',
              gap: 8,
              padding: '8px 20px',
              borderTop: '1px solid var(--border)',
            }}
          >
            <button
              onClick={() => setTab3Mode('faculty')}
              style={{
                padding: '4px 14px',
                borderRadius: 9999,
                border: '1px solid var(--border)',
                background: isFacultyMode ? 'var(--accent)' : 'transparent',
                color: isFacultyMode ? '#fff' : 'var(--text-tertiary)',
                fontSize: 13,
                fontWeight: isFacultyMode ? 600 : 400,
                cursor: 'pointer',
                fontFamily: 'inherit',
                transition: 'all 0.15s ease',
              }}
            >
              {facultyName}
            </button>
            <button
              onClick={() => setTab3Mode('community')}
              style={{
                padding: '4px 14px',
                borderRadius: 9999,
                border: '1px solid var(--border)',
                background: isCommunityMode ? 'var(--accent)' : 'transparent',
                color: isCommunityMode ? '#fff' : 'var(--text-tertiary)',
                fontSize: 13,
                fontWeight: isCommunityMode ? 600 : 400,
                cursor: 'pointer',
                fontFamily: 'inherit',
                transition: 'all 0.15s ease',
              }}
            >
              {pinnedCommunity.name}
            </button>
          </div>
        )}
      </div>

      {(isForYou || isFollowing) && <CreatePostBox />}

      {activeQuery.isLoading ? (
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
        <EmptyFeed emptyKey={emptyKey} />
      )}
    </div>
  );
}
