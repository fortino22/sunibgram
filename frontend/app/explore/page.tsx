'use client';

import { Suspense } from 'react';
import { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { ThreadCardSkeleton } from '@/components/thread/skeletons/ThreadCardSkeleton';
import { EmptyState } from '@/components/shared/EmptyState';
import { MediaGrid } from '@/components/shared/MediaGrid';
import { ExploreSearchBar } from '@/components/explore/ExploreSearchBar';
import { PersonCard } from '@/components/explore/PersonCard';
import { CommunityCard } from '@/components/explore/CommunityCard';
import { ExploreForYou } from '@/components/explore/ExploreForYou';
import { TrendingHashtagsTab } from '@/components/explore/TrendingHashtagsTab';
import {
  useExploreTop,
  useExploreLatest,
  useExplorePeople,
  useExploreMedia,
  useExploreCommunities,
} from '@/hooks/useExplore';
import type { ThreadItem } from '@/lib/services/threads';

type ExploreTab = 'top' | 'latest' | 'people' | 'media' | 'communities';
type HomeTab = 'foryou' | 'trending';

const TABS: { id: ExploreTab; label: string }[] = [
  { id: 'top', label: 'Top' },
  { id: 'latest', label: 'Latest' },
  { id: 'people', label: 'People' },
  { id: 'media', label: 'Media' },
  { id: 'communities', label: 'Communities' },
];

const HOME_TABS: { id: HomeTab; label: string }[] = [
  { id: 'foryou', label: 'For You' },
  { id: 'trending', label: 'Trending' },
];

function TabButton({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      style={{
        flex: 1,
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
        whiteSpace: 'nowrap',
        textAlign: 'center',
        fontFamily: 'inherit',
        letterSpacing: '-0.01em',
      }}
    >
      {label}
    </button>
  );
}

function TabBar({ active, onChange }: { active: ExploreTab; onChange: (tab: ExploreTab) => void }) {
  return (
    <div style={{ display: 'flex', borderBottom: '1px solid var(--border)' }}>
      {TABS.map((tab) => (
        <TabButton key={tab.id} label={tab.label} active={active === tab.id} onClick={() => onChange(tab.id)} />
      ))}
    </div>
  );
}

function HomeTabBar({ active, onChange }: { active: HomeTab; onChange: (tab: HomeTab) => void }) {
  return (
    <div style={{ display: 'flex', borderBottom: '1px solid var(--border)' }}>
      {HOME_TABS.map((tab) => (
        <TabButton key={tab.id} label={tab.label} active={active === tab.id} onClick={() => onChange(tab.id)} />
      ))}
    </div>
  );
}

function SectionLabel({ title }: { title: string }) {
  return (
    <div style={{ padding: '14px 20px 6px', fontSize: 19, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.02em' }}>
      {title}
    </div>
  );
}

function SectionDivider() {
  return (
    <div style={{ height: 8, background: 'var(--bg-subtle)', borderTop: '1px solid var(--border)', borderBottom: '1px solid var(--border)' }} />
  );
}

function ShowMoreButton({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      style={{
        display: 'block',
        width: '100%',
        padding: '14px 20px',
        background: 'none',
        border: 'none',
        borderTop: '1px solid var(--border)',
        cursor: 'pointer',
        textAlign: 'left',
        fontSize: 15,
        color: 'var(--accent)',
        fontFamily: 'inherit',
      }}
      onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
      onMouseLeave={(e) => { e.currentTarget.style.background = 'none'; }}
    >
      {label}
    </button>
  );
}

function TabContent({
  tab,
  query,
  onTabChange,
}: {
  tab: ExploreTab;
  query: string;
  onTabChange: (tab: ExploreTab) => void;
}) {
  const topQuery = useExploreTop(tab === 'top' ? query : '');
  const topPeopleQuery = useExplorePeople(tab === 'top' ? query : '');
  const latestQuery = useExploreLatest(tab === 'latest' ? query : '');
  const peopleQuery = useExplorePeople(tab === 'people' ? query : '');
  const mediaQuery = useExploreMedia(tab === 'media' ? query : '');
  const communitiesQuery = useExploreCommunities(tab === 'communities' ? query : '');

  if (tab === 'top') {
    const isLoading = topQuery.isLoading || topPeopleQuery.isLoading;
    if (isLoading) return <>{Array.from({ length: 4 }, (_, i) => <ThreadCardSkeleton key={i} />)}</>;

    const people = topPeopleQuery.data ?? [];
    const threads = topQuery.data?.data ?? [];

    if (people.length === 0 && threads.length === 0) {
      return <EmptyState title="No results" subtitle={`Nothing found for "${query}"`} />;
    }

    return (
      <>
        {people.length > 0 && (
          <div>
            <SectionLabel title="People" />
            {people.slice(0, 3).map((p) => <PersonCard key={p.userId} person={p} />)}
            {people.length > 3 && (
              <ShowMoreButton label="Show more people" onClick={() => onTabChange('people')} />
            )}
          </div>
        )}
        {people.length > 0 && threads.length > 0 && <SectionDivider />}
        {threads.length > 0 && (
          <div>
            <SectionLabel title="Posts" />
            {threads.map((t: ThreadItem) => <ThreadCard key={t.id} thread={t} />)}
          </div>
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
    const mediaThreads = (mediaQuery.data?.data ?? []).filter((t) => t.media.length > 0);
    if (!mediaThreads.length) return <EmptyState title="No media" subtitle={`No media posts for "${query}"`} />;
    return <MediaGrid threads={mediaThreads} />;
  }

  if (tab === 'communities') {
    if (communitiesQuery.isLoading) return <>{Array.from({ length: 3 }, (_, i) => <ThreadCardSkeleton key={i} />)}</>;
    if (!communitiesQuery.data?.length) return <EmptyState title="No communities" subtitle={`No communities match "${query}"`} />;
    return <>{communitiesQuery.data.map((c) => <CommunityCard key={c.communityId} community={c} />)}</>;
  }

  return null;
}

function ExploreInner() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const initialQ = searchParams.get('q') ?? '';

  const [inputValue, setInputValue] = useState(initialQ);
  const [query, setQuery] = useState(initialQ);
  const [activeTab, setActiveTab] = useState<ExploreTab>('top');
  const [homeTab, setHomeTabState] = useState<HomeTab>(
    searchParams.get('tab') === 'trending' ? 'trending' : 'foryou',
  );

  // Keep tab + committed query in sync with the URL (deep links, back/forward,
  // "Show more" → ?tab=trending). Search is committed only via commitSearch below.
  useEffect(() => {
    setHomeTabState(searchParams.get('tab') === 'trending' ? 'trending' : 'foryou');
    const q = searchParams.get('q') ?? '';
    setQuery(q);
    setInputValue(q);
  }, [searchParams]);

  const hasQuery = query.length > 0;

  // Commit a search: typing alone never queries — only this does.
  function commitSearch(q: string) {
    const v = q.trim();
    setInputValue(v);
    setQuery(v);
    setActiveTab('top');
    router.replace(v ? `/explore?q=${encodeURIComponent(v)}` : '/explore');
  }

  function setHomeTab(tab: HomeTab) {
    setHomeTabState(tab);
    router.replace(tab === 'foryou' ? '/explore' : `/explore?tab=${tab}`);
  }

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
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '20px 16px' }}>
          <ExploreSearchBar value={inputValue} onChange={setInputValue} onSubmit={commitSearch} />
        </div>
        {hasQuery ? (
          <TabBar active={activeTab} onChange={setActiveTab} />
        ) : (
          <HomeTabBar active={homeTab} onChange={setHomeTab} />
        )}
      </div>

      {/* Content */}
      {hasQuery ? (
        <TabContent tab={activeTab} query={query} onTabChange={setActiveTab} />
      ) : homeTab === 'trending' ? (
        <TrendingHashtagsTab />
      ) : (
        <ExploreForYou />
      )}
    </div>
  );
}

export default function ExplorePage() {
  return (
    <Suspense>
      <ExploreInner />
    </Suspense>
  );
}
