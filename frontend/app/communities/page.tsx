'use client';

import { useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { CommunityThreadCard } from '@/components/community/CommunityThreadCard';
import { CreateCommunityModal } from '@/components/community/CreateCommunityModal';
import { EmptyState } from '@/components/shared/EmptyState';
import { Icon } from '@/components/ui/Icon';
import {
  useJoinedCommunities,
  useCategories,
  useMyCommunitiesThreadsFeed,
  useExploreCommunitiesThreads,
} from '@/hooks/useCommunities';

type Tab = 'home' | 'explore';

export default function CommunitiesPage() {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<Tab>('home');
  const [categoryId, setCategoryId] = useState<string | undefined>(undefined);
  const [showCreate, setShowCreate] = useState(false);

  const sliderRef = useRef<HTMLDivElement>(null);
  const drag = useRef({ active: false, startX: 0, scrollLeft: 0, moved: false });
  const categorySliderRef = useRef<HTMLDivElement>(null);
  const categoryDrag = useRef({ active: false, startX: 0, scrollLeft: 0, moved: false });

  const onSliderMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
    const el = sliderRef.current;
    if (!el) return;
    drag.current = { active: true, startX: e.pageX - el.offsetLeft, scrollLeft: el.scrollLeft, moved: false };
    el.style.cursor = 'grabbing';
    el.style.userSelect = 'none';
  };

  const onSliderMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!drag.current.active) return;
    const el = sliderRef.current;
    if (!el) return;
    const walk = e.pageX - el.offsetLeft - drag.current.startX;
    if (Math.abs(walk) > 4) drag.current.moved = true;
    el.scrollLeft = drag.current.scrollLeft - walk;
  };

  const onSliderMouseUp = () => {
    drag.current.active = false;
    if (sliderRef.current) {
      sliderRef.current.style.cursor = 'grab';
      sliderRef.current.style.userSelect = '';
    }
  };

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

  const joinedQuery = useJoinedCommunities();
  const homeFeedQuery = useMyCommunitiesThreadsFeed();
  const exploreFeedQuery = useExploreCommunitiesThreads(categoryId);
  const { data: categoriesData } = useCategories();

  const communities = joinedQuery.data?.data ?? [];
  const categories = categoriesData?.data ?? [];
  const homeThreads = homeFeedQuery.data?.data ?? [];
  const exploreThreads = exploreFeedQuery.data?.data ?? [];

  return (
    <div>
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '14px 20px 0',
          }}
        >
          <div style={{ fontWeight: 700, fontSize: 20, color: 'var(--text)', letterSpacing: '-0.02em' }}>
            Communities
          </div>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <button
              onClick={() => router.push('/communities/discover')}
              style={{
                width: 34,
                height: 34,
                borderRadius: '50%',
                border: '1px solid var(--border)',
                background: 'transparent',
                color: 'var(--text)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Icon name="search" size={16} />
            </button>
            <button
              onClick={() => setShowCreate(true)}
              style={{
                padding: '7px 14px',
                borderRadius: 999,
                border: 'none',
                background: 'var(--text)',
                color: 'var(--bg)',
                fontSize: 13,
                fontWeight: 700,
                cursor: 'pointer',
              }}
            >
              + Create
            </button>
          </div>
        </div>

        <div style={{ display: 'flex', padding: '0 20px' }}>
          {(['home', 'explore'] as Tab[]).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              style={{
                flex: 1,
                padding: '12px 0',
                background: 'none',
                border: 'none',
                borderBottom: activeTab === tab ? '2px solid var(--text)' : '2px solid transparent',
                color: activeTab === tab ? 'var(--text)' : 'var(--text-tertiary)',
                fontSize: 14,
                fontWeight: activeTab === tab ? 700 : 400,
                cursor: 'pointer',
                textTransform: 'capitalize',
              }}
            >
              {tab}
            </button>
          ))}
        </div>
      </div>

      {activeTab === 'home' && (
        <>
          {communities.length > 0 && (
            <div
              ref={sliderRef}
              onMouseDown={onSliderMouseDown}
              onMouseMove={onSliderMouseMove}
              onMouseUp={onSliderMouseUp}
              onMouseLeave={onSliderMouseUp}
              style={{
                display: 'flex',
                gap: 12,
                padding: '16px 20px 12px',
                overflowX: 'auto',
                borderBottom: '1px solid var(--border)',
                scrollbarWidth: 'thin',
                scrollbarColor: 'var(--border) transparent',
                paddingBottom: 12,
                cursor: 'grab',
              }}
            >
              {communities.map((c) => (
                <button
                  key={c.community_id}
                  onClick={() => { if (drag.current.moved) return; router.push(`/communities/${c.community_id}`); }}
                  style={{
                    position: 'relative',
                    flexShrink: 0,
                    width: 140,
                    height: 90,
                    borderRadius: 12,
                    overflow: 'hidden',
                    border: '1px solid var(--border)',
                    background: c.banner_url
                      ? `url(${c.banner_url}) center/cover no-repeat`
                      : 'var(--bg-subtle)',
                    cursor: 'pointer',
                    padding: 0,
                  }}
                >
                  <div
                    style={{
                      position: 'absolute',
                      inset: 0,
                      background: 'linear-gradient(to bottom, transparent 35%, rgba(0,0,0,0.72) 100%)',
                    }}
                  />
                  <span
                    style={{
                      position: 'absolute',
                      bottom: 7,
                      left: 8,
                      right: 8,
                      fontSize: 12,
                      fontWeight: 700,
                      color: '#fff',
                      textAlign: 'left',
                      overflow: 'hidden',
                      display: '-webkit-box',
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: 'vertical' as const,
                      lineHeight: '1.3',
                      textShadow: '0 1px 3px rgba(0,0,0,0.6)',
                    }}
                  >
                    {c.name}
                  </span>
                </button>
              ))}
            </div>
          )}

          {homeFeedQuery.isLoading && (
            <div style={{ padding: 32, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading...
            </div>
          )}
          {homeFeedQuery.isError && (
            <EmptyState title="Failed to load" subtitle="Could not fetch community threads." />
          )}
          {!homeFeedQuery.isLoading && !homeFeedQuery.isError && (
            homeThreads.length === 0 ? (
              <EmptyState title="No posts yet" subtitle="Join communities to see their posts here." />
            ) : (
              homeThreads.map((thread) => (
                <CommunityThreadCard key={thread.id} thread={thread} />
              ))
            )
          )}
        </>
      )}

      {activeTab === 'explore' && (
        <>
          <div
            ref={categorySliderRef}
            onMouseDown={onCategoryMouseDown}
            onMouseMove={onCategoryMouseMove}
            onMouseUp={onCategoryMouseUp}
            onMouseLeave={onCategoryMouseUp}
            style={{
              display: 'flex',
              gap: 8,
              padding: '12px 20px',
              overflowX: 'auto',
              borderBottom: '1px solid var(--border)',
              cursor: 'grab',
              scrollbarWidth: 'thin',
              scrollbarColor: 'var(--border) transparent',
            }}
          >
            <button
              onClick={() => { if (categoryDrag.current.moved) return; setCategoryId(undefined); }}
              style={{
                padding: '5px 14px',
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
                style={{
                  padding: '5px 14px',
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

          {exploreFeedQuery.isLoading && (
            <div style={{ padding: 32, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
              Loading...
            </div>
          )}
          {exploreFeedQuery.isError && (
            <EmptyState title="Failed to load" subtitle="Could not fetch community threads." />
          )}
          {!exploreFeedQuery.isLoading && !exploreFeedQuery.isError && (
            exploreThreads.length === 0 ? (
              <EmptyState title="No posts found" subtitle="Try a different category or check back later." />
            ) : (
              exploreThreads.map((thread) => (
                <CommunityThreadCard key={thread.id} thread={thread} />
              ))
            )
          )}
        </>
      )}

      {showCreate && (
        <CreateCommunityModal
          onClose={() => setShowCreate(false)}
          onSuccess={(id) => {
            setShowCreate(false);
            router.push(`/communities/${id}`);
          }}
        />
      )}
    </div>
  );
}
