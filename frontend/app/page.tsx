'use client';

import { useState, useRef, useLayoutEffect, useEffect } from 'react';
import { ChevronDown } from 'lucide-react';
import { CreatePostBox } from '@/components/shared/CreatePostBox';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { NewPostsBanner } from '@/components/thread/NewPostsBanner';
import { useFollowingThreads } from '@/hooks/useFollowingThreads';
import { useThreads } from '@/hooks/useThreads';
import { useFacultyThreads } from '@/hooks/useFacultyThreads';
import { useFeedNewCount } from '@/hooks/useFeedNewCount';
import { useShowNewPosts } from '@/hooks/useShowNewPosts';
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
  const ref = useRef<HTMLButtonElement | null>(null);
  useEffect(() => {
    if (active && ref.current) {
      ref.current.scrollIntoView({ behavior: 'smooth', inline: 'center', block: 'nearest' });
    }
  }, [active]);
  return (
    <button
      ref={ref}
      onClick={onClick}
      className="btn-ghost feed-tab-btn"
      style={{
        flex: 1,
        minWidth: 'max-content',
        padding: '12px 8px',
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
      title={tab.label}
    >
      <span className="feed-tab-label">{tab.label}</span>
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

  // For You "new posts" banner watermark.
  // sinceId is the highest feed-eligible thread id that existed when the feed first
  // loaded, captured ONCE from page 1's meta.max_thread_id (the backend computes the
  // true global max — page-1 ordering is fresh/feed_score, not id, so we cannot derive
  // it from the visible threads). Frozen for the session and only ever advanced, never
  // decreased: the count of thread_id > sinceId is therefore 0 on a fresh load and grows
  // only for posts created afterward. Advancing on banner click (below) prevents the just-
  // prepended posts from re-triggering the banner.
  const forYouThreads = dedupeThreads(threadsQuery.data?.pages.flatMap((p) => p.data) ?? []);
  const sessionMaxIdRef = useRef(0);
  const [sinceId, setSinceId] = useState(0);
  const feedMaxThreadId = threadsQuery.data?.pages[0]?.meta.max_thread_id ?? 0;
  useEffect(() => {
    // Seed the watermark from page 1 once. Guard on ref so a later refetch with a
    // higher max_thread_id (e.g. background poll) does not silently swallow new posts.
    if (feedMaxThreadId > 0 && sessionMaxIdRef.current === 0) {
      sessionMaxIdRef.current = feedMaxThreadId;
      setSinceId(feedMaxThreadId);
    }
  }, [feedMaxThreadId]);

  const newCount = useFeedNewCount(sinceId, isForYou);
  const showNewPosts = useShowNewPosts();

  function handleShowNew() {
    if (showNewPosts.isPending) return;
    showNewPosts.mutate(sinceId, {
      onSuccess: (result) => {
        // Advance the watermark past the just-prepended posts so the next poll counts
        // only threads newer than these. Never decrease.
        if (result.maxId > sessionMaxIdRef.current) {
          sessionMaxIdRef.current = result.maxId;
          setSinceId(result.maxId);
        }
      },
    });
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  const sentinelRef = useRef<HTMLDivElement>(null);

  // Subscribe to the active query's pagination state via primitives so the
  // observer only re-creates when those primitives change, not on every render.
  const { hasNextPage, isFetchingNextPage, fetchNextPage } = activeQuery;

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && hasNextPage && !isFetchingNextPage) {
          void fetchNextPage();
        }
      },
      { rootMargin: '300px' },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

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
              className="btn-ghost"
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                width: '100%',
                padding: '11px 14px',
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
          background: 'var(--bg-blur)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div className="feed-header" style={{ padding: '20px 20px 0' }}>
          <div
            className="feed-header-title"
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
            className="feed-tabs"
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
                    className="btn-ghost feed-tab-btn"
                    title={t.label}
                    style={{
                      flex: 1,
                      minWidth: 'max-content',
                      padding: '12px 8px',
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
                    <span className="feed-tab-label">{t.label}</span>
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
