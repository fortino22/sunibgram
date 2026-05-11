'use client';

import { useEffect, useState } from 'react';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { SearchBar } from '@/components/explore/SearchBar';
import { TrendingSection } from '@/components/explore/TrendingSection';
import { SuggestedPeopleSection } from '@/components/explore/SuggestedPeopleSection';
import { PersonCard } from '@/components/explore/PersonCard';
import { CommunityCard } from '@/components/explore/CommunityCard';
import { ThreadSnippet } from '@/components/explore/ThreadSnippet';
import {
  useTrendingHashtags,
  useSuggestedPeople,
  useExploreSearch,
  useExploreLatest,
  useExplorePeople,
  useExploreMedia,
  useExploreCommunities,
} from '@/hooks/useExplore';
import type { ThreadItem } from '@/lib/services/threads';

type ExploreTab = 'top' | 'latest' | 'people' | 'media' | 'communities';

const TABS: { id: ExploreTab; label: string }[] = [
  { id: 'top', label: 'Top' },
  { id: 'latest', label: 'Latest' },
  { id: 'people', label: 'People' },
  { id: 'media', label: 'Media' },
  { id: 'communities', label: 'Communities' },
];

function TabBar({
  active,
  onChange,
}: {
  active: ExploreTab;
  onChange: (tab: ExploreTab) => void;
}) {
  return (
    <div
      style={{
        display: 'flex',
        borderBottom: '1px solid var(--border)',
      }}
    >
      {TABS.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onChange(tab.id)}
          style={{
            flex: 1,
            padding: '12px 0',
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            fontWeight: active === tab.id ? 600 : 400,
            fontSize: 14,
            color: active === tab.id ? 'var(--text)' : 'var(--text-tertiary)',
            borderBottom: active === tab.id ? '2px solid var(--accent)' : '2px solid transparent',
            marginBottom: -1,
            transition: 'all 0.15s',
            whiteSpace: 'nowrap',
            fontFamily: 'inherit',
          }}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}

function SectionHeader({ title }: { title: string }) {
  return (
    <div
      style={{
        padding: '12px 20px 4px',
        fontSize: 13,
        fontWeight: 600,
        color: 'var(--text-tertiary)',
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
      }}
    >
      {title}
    </div>
  );
}

function TabContent({ tab, query }: { tab: ExploreTab; query: string }) {
  // Top tab uses ExploreSearch — shows mixed users + thread snippets
  const searchQuery = useExploreSearch(tab === 'top' ? query : '');
  const latestQuery = useExploreLatest(tab === 'latest' ? query : '');
  const peopleQuery = useExplorePeople(tab === 'people' ? query : '');
  const mediaQuery = useExploreMedia(tab === 'media' ? query : '');
  const communitiesQuery = useExploreCommunities(tab === 'communities' ? query : '');

  if (tab === 'top') {
    if (searchQuery.isLoading) {
      return <>{Array.from({ length: 4 }, (_, i) => <ThreadCardSkeleton key={i} />)}</>;
    }
    const users = searchQuery.data?.users ?? [];
    const threads = searchQuery.data?.threads ?? [];
    if (users.length === 0 && threads.length === 0) {
      return <EmptyState title="No results" subtitle={`Nothing found for "${query}"`} />;
    }
    return (
      <>
        {users.length > 0 && (
          <>
            <SectionHeader title="People" />
            {users.map((u) => (
              <PersonCard
                key={u.userId}
                person={{ userId: u.userId, userProfilePic: u.profile, username: u.username, name: u.name, isPremium: u.isPremium, isFollowing: u.isFollowing }}
              />
            ))}
          </>
        )}
        {threads.length > 0 && (
          <>
            <SectionHeader title="Posts" />
            {threads.map((t, idx) => (
              <ThreadSnippet key={idx} content={t.content} />
            ))}
          </>
        )}
      </>
    );
  }

  if (tab === 'latest') {
    if (latestQuery.isLoading) return <>{Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)}</>;
    if (!latestQuery.data?.data.length) return <EmptyState title="No results" subtitle={`No latest posts for "${query}"`} />;
    return <>{latestQuery.data.data.map((t: ThreadItem) => <ThreadCard key={t.id} thread={t} />)}</>;
  }

  if (tab === 'people') {
    if (peopleQuery.isLoading) return <>{Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)}</>;
    if (!peopleQuery.data?.length) return <EmptyState title="No people found" subtitle={`No accounts match "${query}"`} />;
    return <>{peopleQuery.data.map((p) => <PersonCard key={p.userId} person={p} />)}</>;
  }

  if (tab === 'media') {
    if (mediaQuery.isLoading) return <>{Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)}</>;
    if (!mediaQuery.data?.data.length) return <EmptyState title="No media" subtitle={`No media posts for "${query}"`} />;
    return <>{mediaQuery.data.data.map((t: ThreadItem) => <ThreadCard key={t.id} thread={t} />)}</>;
  }

  if (tab === 'communities') {
    if (communitiesQuery.isLoading) return <>{Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)}</>;
    if (!communitiesQuery.data?.length) return <EmptyState title="No communities" subtitle={`No communities match "${query}"`} />;
    return <>{communitiesQuery.data.map((c) => <CommunityCard key={c.communityId} community={c} />)}</>;
  }

  return null;
}

export default function ExplorePage() {
  const [inputValue, setInputValue] = useState('');
  const [query, setQuery] = useState('');
  const [activeTab, setActiveTab] = useState<ExploreTab>('top');

  const trendingQuery = useTrendingHashtags(10);
  const suggestedQuery = useSuggestedPeople(8);

  // Debounce search input by 400ms
  useEffect(() => {
    const timer = setTimeout(() => setQuery(inputValue.trim()), 400);
    return () => clearTimeout(timer);
  }, [inputValue]);

  // Reset to Top tab whenever a new search begins
  useEffect(() => {
    if (query) setActiveTab('top');
  }, [query]);

  const hasQuery = query.length > 0;

  return (
    <div>
      {/* Sticky header */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: hasQuery ? 'none' : '1px solid var(--border)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 16px' }}>
          <SearchBar value={inputValue} onChange={setInputValue} />
        </div>
        {hasQuery && <TabBar active={activeTab} onChange={setActiveTab} />}
      </div>

      {/* Content */}
      {hasQuery ? (
        <TabContent tab={activeTab} query={query} />
      ) : (
        <>
          {trendingQuery.isLoading ? (
            <div style={{ padding: '20px', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading trends…
            </div>
          ) : (
            <TrendingSection hashtags={trendingQuery.data ?? []} />
          )}

          {!suggestedQuery.isLoading && (suggestedQuery.data?.length ?? 0) > 0 && (
            <>
              <div
                style={{
                  height: 8,
                  background: 'var(--bg-subtle)',
                  borderTop: '1px solid var(--border)',
                  borderBottom: '1px solid var(--border)',
                }}
              />
              <SuggestedPeopleSection people={suggestedQuery.data ?? []} />
            </>
          )}
        </>
      )}
    </div>
  );
}
