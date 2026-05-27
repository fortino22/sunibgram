'use client';

import { useState, useRef, useLayoutEffect, useEffect } from 'react';
import { ChevronDown } from 'lucide-react';
import { CreatePostBox } from '@/components/shared/CreatePostBox';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { NewPostsBanner } from '@/components/thread/NewPostsBanner';
import { Icon } from '@/components/ui/Icon';
import { useFollowingThreads } from '@/hooks/useFollowingThreads';
import { useThreads } from '@/hooks/useThreads';
import { useFacultyThreads } from '@/hooks/useFacultyThreads';
import { useFeedNewCount } from '@/hooks/useFeedNewCount';
import { useRefreshFeed } from '@/hooks/useRefreshFeed';
import { usePinnedCommunities, useCommunityThreadsBySort } from '@/hooks/useCommunities';
import { useMyProfile } from '@/hooks/useMyProfile';
import type { ThreadItem } from '@/lib/services/threads';

type CommunitySort = 'latest' | 'top' | 'likes';

// Offset-paginated feeds can return the same thread in two pages when the
// underlying list shifts between page fetches (seed reshuffle on refresh,
// a deleted thread, periodic feed_score recompute). Duplicate thread.ids
// become duplicate React keys, which leaves the duped card unclickable.
// Dedupe by id, keeping first occurrence, before rendering.
function dedupeThreads(threads: ThreadItem[]): ThreadItem[] {
  const seen = new Set<number>();
  const out: ThreadItem[] = [];
  for (const t of threads) {
    if (!seen.has(t.id)) {
      seen.add(t.id);
      out.push(t);
    }
  }
  return out;
}

const SORT_OPTIONS: { value: CommunitySort; label: string }[] = [
  { value: 'latest', label: 'Latest' },
  { value: 'top', label: 'Top' },
  { value: 'likes', label: 'Likes' },
];

interface Tab {
  id: string;
  label: string;
  isCommunity?: boolean;
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

function ActiveCommunityFeed({ communityId, sort }: { communityId: number; sort: CommunitySort }) {
  const query = useCommunityThreadsBySort(communityId, sort);
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
        <ThreadCard key={thread.id} thread={thread} inCommunityContext />
      ))}
    </div>
  );
}

export default function HomePage() {
  const [activeTab, setActiveTab] = useState<string>('for-you');
  const [communitySort, setCommunitySort] = useState<CommunitySort>('latest');
  const [sortDropdownOpen, setSortDropdownOpen] = useState(false);
  const [dropdownPos, setDropdownPos] = useState<{ top: number; centerX: number } | null>(null);

  const activeCommunityBtnRef = useRef<HTMLButtonElement | null>(null);

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
  const followingQuery = useFollowingThreads(isFollowing);
  const facultyQuery = useFacultyThreads(facultyId, isFaculty);
  const pinnedCommunitiesQuery = usePinnedCommunities();

  const activeQuery = isFollowing
    ? followingQuery
    : isFaculty
    ? facultyQuery
    : threadsQuery;

  const activeThreads = dedupeThreads(activeQuery.data?.pages.flatMap((p) => p.data) ?? []);
  const pinnedCommunities = pinnedCommunitiesQuery.data?.data ?? [];

  // For You "new posts" banner — sinceId is the max loaded thread id (PK), not data[0].
  const forYouThreads = dedupeThreads(threadsQuery.data?.pages.flatMap((p) => p.data) ?? []);
  const sinceId = forYouThreads.length ? Math.max(...forYouThreads.map((t) => t.id)) : 0;
  const newCount = useFeedNewCount(sinceId, isForYou);
  const refreshFeed = useRefreshFeed();

  function handleShowNew() {
    refreshFeed.mutate();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && activeQuery.hasNextPage && !activeQuery.isFetchingNextPage) {
          void activeQuery.fetchNextPage();
        }
      },
      { rootMargin: '300px' },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [activeQuery]);

  const allTabs: Tab[] = [
    { id: 'for-you', label: 'For You' },
    { id: 'following', label: 'Following' },
    { id: 'faculty', label: facultyName },
    ...pinnedCommunities.map((c) => ({
      id: `community-${c.community_id}`,
      label: c.community_name,
      isCommunity: true,
    })),
  ];

  // Measure active button position after DOM updates to position the fixed dropdown
  useLayoutEffect(() => {
    if (sortDropdownOpen && activeCommunityBtnRef.current) {
      const rect = activeCommunityBtnRef.current.getBoundingClientRect();
      setDropdownPos({ top: rect.bottom + 6, centerX: rect.left + rect.width / 2 });
    } else {
      setDropdownPos(null);
    }
  }, [sortDropdownOpen, activeTab]);

  function handleTabClick(tabId: string) {
    if (tabId.startsWith('community-')) {
      if (activeTab === tabId) {
        setSortDropdownOpen((prev) => !prev);
      } else {
        setActiveTab(tabId);
        setCommunitySort('latest');
        setSortDropdownOpen(false);
      }
    } else {
      setActiveTab(tabId);
      setSortDropdownOpen(false);
    }
  }

  function handleSortSelect(sort: CommunitySort) {
    setCommunitySort(sort);
    setSortDropdownOpen(false);
  }

  return (
    <div>
      {/* invisible overlay to close dropdown on outside click */}
      {sortDropdownOpen && (
        <div
          style={{ position: 'fixed', inset: 0, zIndex: 40 }}
          onClick={() => setSortDropdownOpen(false)}
        />
      )}

      {/* sort dropdown — fixed so it escapes the tab strip's overflow container */}
      {sortDropdownOpen && dropdownPos && (
        <div
          style={{
            position: 'fixed',
            top: dropdownPos.top,
            left: dropdownPos.centerX,
            transform: 'translateX(-50%)',
            zIndex: 50,
            background: 'var(--bg-elevated, var(--bg-subtle))',
            border: '1px solid var(--border)',
            borderRadius: 10,
            overflow: 'hidden',
            minWidth: 130,
            boxShadow: '0 4px 20px rgba(0,0,0,0.35)',
          }}
        >
          {SORT_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => handleSortSelect(opt.value)}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                width: '100%',
                padding: '11px 14px',
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: communitySort === opt.value ? 600 : 400,
                color: 'var(--text)',
                textAlign: 'left',
                fontFamily: 'inherit',
                gap: 12,
              }}
            >
              {opt.label}
              {communitySort === opt.value && (
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="20 6 9 17 4 12" />
                </svg>
              )}
            </button>
          ))}
        </div>
      )}

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
            {allTabs.map((t) => {
              const active = activeTab === t.id;
              if (t.isCommunity) {
                return (
                  <button
                    key={t.id}
                    ref={(el) => { if (active) activeCommunityBtnRef.current = el; }}
                    onClick={() => handleTabClick(t.id)}
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
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      gap: 3,
                    }}
                  >
                    {t.label}
                    {active && (
                      <ChevronDown
                        size={13}
                        style={{
                          flexShrink: 0,
                          transition: 'transform 0.15s ease',
                          transform: sortDropdownOpen ? 'rotate(180deg)' : 'rotate(0deg)',
                        }}
                      />
                    )}
                  </button>
                );
              }
              return (
                <FeedTabButton
                  key={t.id}
                  tab={t}
                  active={active}
                  onClick={() => handleTabClick(t.id)}
                />
              );
            })}
          </div>
        </div>
      </div>

      {(isForYou || isFollowing) && <CreatePostBox />}

      {isForYou && newCount > 0 && (
        <NewPostsBanner count={newCount} onClick={handleShowNew} />
      )}

      {activeCommunityId ? (
        <ActiveCommunityFeed communityId={activeCommunityId} sort={communitySort} />
      ) : activeQuery.isLoading ? (
        <div>
          {Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)}
        </div>
      ) : activeThreads.length > 0 ? (
        <div>
          {activeThreads.map((thread: ThreadItem) => (
            <ThreadCard key={thread.id} thread={thread} />
          ))}
          <div ref={sentinelRef} style={{ height: 1 }} />
          {activeQuery.isFetchingNextPage && (
            <div>{Array.from({ length: 2 }, (_, i) => <ThreadCardSkeleton key={i} />)}</div>
          )}
        </div>
      ) : (
        <EmptyFeed emptyKey={activeTab} />
      )}
    </div>
  );
}
