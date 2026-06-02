'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { SearchDropdown } from './SearchDropdown';
import { PremiumCard } from './PremiumCard';
import { TrendingHashtags } from './TrendingHashtags';
import { TrendingEvents } from './TrendingEvents';
import { WhoToFollow } from './WhoToFollow';

const FOOTER_LINKS = [
  { label: 'Terms', href: '/about/terms' },
  { label: 'Privacy', href: '/about/privacy' },
  { label: 'Cookies', href: '/about/privacy' }, // no dedicated cookies page
  { label: 'About', href: '/about' },
  { label: 'Help', href: '/about/help' },
];

export function RightPanel() {
  const pathname = usePathname();
  const isExplorePage = pathname === '/explore';

  if (pathname.startsWith('/messages') || pathname.startsWith('/admin') || pathname === '/events') return null;

  return (
    <aside className="right-panel">
      <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 16 }}>

        {/* Search dropdown — hidden on the explore page (it has its own search bar) */}
        {!isExplorePage && <SearchDropdown />}

        {/* Premium trial card */}
        <PremiumCard />

        {/* Today's News */}
        <TrendingEvents />

        {/* What's happening — trending hashtags */}
        <TrendingHashtags title="What's happening" limit={4} />

        {/* Who to follow */}
        <WhoToFollow />

        {/* Footer links */}
        <div style={{ fontSize: 11, color: 'var(--text-tertiary)', lineHeight: 1.8 }}>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0 10px' }}>
            {FOOTER_LINKS.map((l) => (
              <Link
                key={l.label}
                href={l.href}
                style={{ cursor: 'pointer', color: 'var(--text-tertiary)' }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.color = 'var(--text-secondary)'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.color = 'var(--text-tertiary)'; }}
              >
                {l.label}
              </Link>
            ))}
          </div>
          <div style={{ marginTop: 4 }}>© 2026 Binusgram</div>
        </div>
      </div>
    </aside>
  );
}
