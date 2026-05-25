'use client';

import { usePathname } from 'next/navigation';
import { SearchDropdown } from './SearchDropdown';
import { TrendingHashtags } from './TrendingHashtags';

export function RightPanel() {
  const pathname = usePathname();
  const isExplorePage = pathname === '/explore';

  if (pathname.startsWith('/messages')) return null;

  return (
    <aside className="right-panel">
      <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 20 }}>

        {/* Search dropdown — hidden on the explore page itself */}
        {!isExplorePage && <SearchDropdown />}

        {/* Trending hashtags */}
        <TrendingHashtags title="Trending on campus" initialVisibleCount={5} />

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
