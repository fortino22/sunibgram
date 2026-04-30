'use client';

import { useState } from 'react';
import { Icon } from '@/components/ui/Icon';

interface TrendItem {
  id: string;
  topic: string;
  posts: string;
  category: string;
  hot: boolean;
}

const TRENDING: TrendItem[] = [
  { id: 't1', topic: '#BinusHackathon2026', posts: '4.2K posts', category: 'Events',      hot: true  },
  { id: 't2', topic: '#AIElective',         posts: '1.8K posts', category: 'Academic',    hot: true  },
  { id: 't3', topic: '#LibraryReno',        posts: '987 posts',  category: 'Campus Life', hot: false },
  { id: 't4', topic: '#FinalExamSurvival',  posts: '3.1K posts', category: 'Academic',    hot: false },
  { id: 't5', topic: '#BinusStartupWeek',   posts: '2.4K posts', category: 'Events',      hot: false },
];

function TrendCard({ trend, last }: { trend: TrendItem; last: boolean }) {
  const [hovered, setHovered] = useState(false);
  return (
    <div
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        padding: '10px 16px',
        background: hovered ? 'var(--bg-subtle)' : 'transparent',
        borderBottom: last ? 'none' : '1px solid var(--border)',
        cursor: 'pointer',
        transition: 'background 0.1s ease',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
      }}
    >
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 2 }}>
          <span style={{ fontSize: 13, fontWeight: 500, color: 'var(--text)' }}>
            {trend.topic}
          </span>
          {trend.hot && (
            <span
              style={{
                fontSize: 10,
                fontWeight: 500,
                color: 'var(--coral)',
                background: 'var(--coral-soft)',
                padding: '1px 5px',
                borderRadius: 9999,
                letterSpacing: '0.04em',
              }}
            >
              HOT
            </span>
          )}
        </div>
        <div style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>
          {trend.category} · {trend.posts}
        </div>
      </div>
      <Icon name="more" size={14} color="var(--text-tertiary)" />
    </div>
  );
}

export function RightPanel() {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchFocused, setSearchFocused] = useState(false);

  return (
    <aside className="right-panel">
      <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 20 }}>

        {/* Search */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            background: searchFocused ? 'var(--bg-card)' : 'var(--bg-subtle)',
            border: `1.5px solid ${searchFocused ? 'var(--accent)' : 'var(--border)'}`,
            borderRadius: 9999,
            padding: '8px 14px',
            transition: 'all 0.15s ease',
            boxShadow: searchFocused ? '0 0 0 3px oklch(64% 0.22 270 / 0.15)' : 'none',
          }}
        >
          <Icon name="search" size={15} color="var(--text-tertiary)" />
          <input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onFocus={() => setSearchFocused(true)}
            onBlur={() => setSearchFocused(false)}
            placeholder="Search campus…"
            style={{
              border: 'none',
              background: 'transparent',
              outline: 'none',
              fontSize: 14,
              color: 'var(--text)',
              width: '100%',
              fontFamily: 'inherit',
            }}
          />
        </div>

        {/* Trending hashtags */}
        <div
          style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 14,
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              padding: '14px 16px 10px',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              alignItems: 'center',
              gap: 8,
            }}
          >
            <Icon name="trending" size={16} color="var(--text-secondary)" />
            <span
              style={{
                fontWeight: 500,
                fontSize: 15,
                color: 'var(--text)',
                letterSpacing: '-0.02em',
              }}
            >
              Trending on campus
            </span>
          </div>
          {TRENDING.map((t, i) => (
            <TrendCard key={t.id} trend={t} last={i === TRENDING.length - 1} />
          ))}
          <div style={{ padding: '10px 16px', borderTop: '1px solid var(--border)' }}>
            <button
              style={{
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                fontSize: 13,
                fontWeight: 500,
                color: 'var(--accent)',
                display: 'flex',
                alignItems: 'center',
                gap: 4,
                fontFamily: 'inherit',
              }}
            >
              See all trends
              <Icon name="arrowRight" size={13} color="var(--accent)" />
            </button>
          </div>
        </div>

        {/* Footer links */}
        <div style={{ fontSize: 11, color: 'var(--text-tertiary)', lineHeight: 1.8 }}>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0 10px' }}>
            {['Terms', 'Privacy', 'Cookies', 'About', 'Help'].map((l) => (
              <span
                key={l}
                style={{ cursor: 'pointer' }}
                onMouseEnter={(e) => { (e.target as HTMLElement).style.color = 'var(--text-secondary)'; }}
                onMouseLeave={(e) => { (e.target as HTMLElement).style.color = 'var(--text-tertiary)'; }}
              >
                {l}
              </span>
            ))}
          </div>
          <div style={{ marginTop: 4 }}>© 2026 Binusgram</div>
        </div>
      </div>
    </aside>
  );
}
