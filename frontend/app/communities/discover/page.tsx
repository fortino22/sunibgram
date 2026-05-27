'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Icon } from '@/components/ui/Icon';
import { CommunityDiscoverCard } from '@/components/community/CommunityDiscoverCard';
import { CommunityThreadCard } from '@/components/community/CommunityThreadCard';
import { EmptyState } from '@/components/shared/EmptyState';
import {
  useDiscoverCommunities,
  useCategories,
  useSearchCommunities,
  useSearchCommunityThreadsTop,
  useSearchCommunityThreadsLatest,
} from '@/hooks/useCommunities';

type SearchTab = 'communities' | 'top' | 'latest';

export default function DiscoverCommunitiesPage() {
  const router = useRouter();
  const [inputValue, setInputValue] = useState('');
  const [query, setQuery] = useState('');
  const [searchTab, setSearchTab] = useState<SearchTab>('communities');
  const [categoryId, setCategoryId] = useState<string | undefined>(undefined);
  const inputRef = useRef<HTMLInputElement>(null);

  const categorySliderRef = useRef<HTMLDivElement>(null);
  const categoryDrag = useRef({ active: false, startX: 0, scrollLeft: 0, moved: false });

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  useEffect(() => {
    const t = setTimeout(() => setQuery(inputValue.trim()), 400);
    return () => clearTimeout(t);
  }, [inputValue]);

  const discoverQuery = useDiscoverCommunities(categoryId);
  const { data: categoriesData } = useCategories();
  const searchCommunitiesQuery = useSearchCommunities(query);
  const topThreadsQuery = useSearchCommunityThreadsTop(query);
  const latestThreadsQuery = useSearchCommunityThreadsLatest(query);

  const communities = discoverQuery.data ?? [];
  const categories = categoriesData?.data ?? [];
  const isSearching = query.length > 0;

  const onCategoryMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
    const el = categorySliderRef.current;
    if (!el) return;
    categoryDrag.current = { active: true, startX: e.pageX - el.offsetLeft, scrollLeft: el.scrollLeft, moved: false };
    el.style.cursor = 'grabbing';
    el.style.userSelect = 'none';
  };

  const onCategoryMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!categoryDrag.current.active) return;
    const el = categorySliderRef.current;
    if (!el) return;
    const walk = e.pageX - el.offsetLeft - categoryDrag.current.startX;
    if (Math.abs(walk) > 4) categoryDrag.current.moved = true;
    el.scrollLeft = categoryDrag.current.scrollLeft - walk;
  };

  const onCategoryMouseUp = () => {
    categoryDrag.current.active = false;
    if (categorySliderRef.current) {
      categorySliderRef.current.style.cursor = 'grab';
      categorySliderRef.current.style.userSelect = '';
    }
  };

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)' }}>
      {/* Header */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '20px 14px',
          display: 'flex',
          gap: 10,
          alignItems: 'center',
        }}
      >
        <button
          onClick={() => router.back()}
          className="btn-ghost"
          style={{
            border: 'none',
            cursor: 'pointer',
            padding: '4px 6px',
            color: 'var(--text)',
            fontSize: 20,
            lineHeight: 1,
            flexShrink: 0,
          }}
        >
          ←
        </button>

        <div style={{ flex: 1, position: 'relative' }}>
          <span
            style={{
              position: 'absolute',
              left: 11,
              top: '50%',
              transform: 'translateY(-50%)',
              color: 'var(--text-tertiary)',
              pointerEvents: 'none',
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <Icon name="search" size={15} />
          </span>
          <input
            ref={inputRef}
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            placeholder="Search for Communities and Posts"
            style={{
              width: '100%',
              padding: '9px 36px 9px 36px',
              borderRadius: 999,
              border: '1px solid var(--border)',
              background: 'var(--bg-subtle)',
              color: 'var(--text)',
              fontSize: 14,
              outline: 'none',
              boxSizing: 'border-box',
            }}
          />
          {inputValue && (
            <button
              onClick={() => { setInputValue(''); setQuery(''); inputRef.current?.focus(); }}
              className="hover-dim"
              style={{
                position: 'absolute',
                right: 10,
                top: '50%',
                transform: 'translateY(-50%)',
                background: 'var(--text-tertiary)',
                border: 'none',
                borderRadius: '50%',
                width: 18,
                height: 18,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                color: 'var(--bg)',
                fontSize: 11,
                fontWeight: 700,
                padding: 0,
              }}
            >
              ✕
            </button>
          )}
        </div>
      </div>

      {isSearching ? (
        <>
          {/* Search tabs */}
          <div style={{ display: 'flex', borderBottom: '1px solid var(--border)' }}>
            {(['communities', 'top', 'latest'] as SearchTab[]).map((tab) => (
              <button
                key={tab}
                onClick={() => setSearchTab(tab)}
                className="btn-ghost"
                style={{
                  flex: 1,
                  padding: '12px 0',
                  border: 'none',
                  borderBottom: searchTab === tab ? '2px solid var(--text)' : '2px solid transparent',
                  color: searchTab === tab ? 'var(--text)' : 'var(--text-tertiary)',
                  fontSize: 14,
                  fontWeight: searchTab === tab ? 700 : 400,
                  cursor: 'pointer',
                  textTransform: 'capitalize',
                }}
              >
                {tab.charAt(0).toUpperCase() + tab.slice(1)}
              </button>
            ))}
          </div>

          {searchTab === 'communities' && (
            <>
              {searchCommunitiesQuery.isLoading && (
                <div style={{ padding: 32, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>Loading...</div>
              )}
              {searchCommunitiesQuery.isError && (
                <EmptyState title="Search failed" subtitle="Could not search communities." />
              )}
              {!searchCommunitiesQuery.isLoading && !searchCommunitiesQuery.isError && (
                (searchCommunitiesQuery.data?.data ?? []).length === 0 ? (
                  <EmptyState title="No communities found" subtitle={`No results for "${query}"`} />
                ) : (
                  <div style={{ paddingTop: 4, paddingBottom: 16 }}>
                    {(searchCommunitiesQuery.data?.data ?? []).map((c) => (
                      <CommunityDiscoverCard
                        key={c.community_id}
                        community={c}
                        onClick={() => router.push(`/communities/${c.community_id}`)}
                      />
                    ))}
                  </div>
                )
              )}
            </>
          )}

          {searchTab === 'top' && (
            <>
              {topThreadsQuery.isLoading && (
                <div style={{ padding: 32, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>Loading...</div>
              )}
              {topThreadsQuery.isError && (
                <EmptyState title="Search failed" subtitle="Could not fetch top threads." />
              )}
              {!topThreadsQuery.isLoading && !topThreadsQuery.isError && (
                (topThreadsQuery.data?.data ?? []).length === 0 ? (
                  <EmptyState title="No threads found" subtitle={`No top results for "${query}"`} />
                ) : (
                  (topThreadsQuery.data?.data ?? []).map((thread) => (
                    <CommunityThreadCard key={thread.id} thread={thread} />
                  ))
                )
              )}
            </>
          )}

          {searchTab === 'latest' && (
            <>
              {latestThreadsQuery.isLoading && (
                <div style={{ padding: 32, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>Loading...</div>
              )}
              {latestThreadsQuery.isError && (
                <EmptyState title="Search failed" subtitle="Could not fetch latest threads." />
              )}
              {!latestThreadsQuery.isLoading && !latestThreadsQuery.isError && (
                (latestThreadsQuery.data?.data ?? []).length === 0 ? (
                  <EmptyState title="No threads found" subtitle={`No latest results for "${query}"`} />
                ) : (
                  (latestThreadsQuery.data?.data ?? []).map((thread) => (
                    <CommunityThreadCard key={thread.id} thread={thread} />
                  ))
                )
              )}
            </>
          )}
        </>
      ) : (
        <>
          <div
            style={{
              padding: '18px 16px 10px',
              fontSize: 20,
              fontWeight: 800,
              color: 'var(--text)',
              letterSpacing: '-0.02em',
            }}
          >
            Discover Communities
          </div>

          {/* Category filter strip */}
          <div
            ref={categorySliderRef}
            onMouseDown={onCategoryMouseDown}
            onMouseMove={onCategoryMouseMove}
            onMouseUp={onCategoryMouseUp}
            onMouseLeave={onCategoryMouseUp}
            style={{
              display: 'flex',
              gap: 8,
              padding: '4px 16px 12px',
              overflowX: 'auto',
              cursor: 'grab',
              scrollbarWidth: 'none',
            }}
          >
            <button
              onClick={() => { if (categoryDrag.current.moved) return; setCategoryId(undefined); }}
              className="hover-dim"
              style={{
                padding: '6px 16px',
                borderRadius: 999,
                border: `1px solid ${!categoryId ? 'var(--text)' : 'var(--border)'}`,
                background: !categoryId ? 'var(--text)' : 'transparent',
                color: !categoryId ? 'var(--bg)' : 'var(--text)',
                fontSize: 13,
                fontWeight: 600,
                cursor: 'pointer',
                whiteSpace: 'nowrap',
                flexShrink: 0,
              }}
            >
              All
            </button>
            {categories.map((c) => (
              <button
                key={c.id}
                onClick={() => { if (categoryDrag.current.moved) return; setCategoryId(c.id); }}
                className="hover-dim"
                style={{
                  padding: '6px 16px',
                  borderRadius: 999,
                  border: `1px solid ${categoryId === c.id ? 'var(--text)' : 'var(--border)'}`,
                  background: categoryId === c.id ? 'var(--text)' : 'transparent',
                  color: categoryId === c.id ? 'var(--bg)' : 'var(--text)',
                  fontSize: 13,
                  fontWeight: 600,
                  cursor: 'pointer',
                  whiteSpace: 'nowrap',
                  flexShrink: 0,
                }}
              >
                {c.name}
              </button>
            ))}
          </div>

          {discoverQuery.isLoading && (
            <div style={{ padding: 32, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>Loading...</div>
          )}
          {discoverQuery.isError && (
            <EmptyState title="Failed to load" subtitle="Could not fetch communities." />
          )}
          {!discoverQuery.isLoading && !discoverQuery.isError && (
            communities.length === 0 ? (
              <EmptyState title="No communities yet" subtitle="Be the first to create one." />
            ) : (
              <div style={{ paddingTop: 4, paddingBottom: 16 }}>
                {communities.map((c) => (
                  <CommunityDiscoverCard
                    key={c.community_id}
                    community={c}
                    onClick={() => router.push(`/communities/${c.community_id}`)}
                  />
                ))}
              </div>
            )
          )}
        </>
      )}
    </div>
  );
}
