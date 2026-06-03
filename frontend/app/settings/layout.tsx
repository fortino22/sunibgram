'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';
import { Icon } from '@/components/ui/Icon';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';

interface SettingsSection {
  href: string;
  icon: string;
  label: string;
  description: string;
}

const SECTIONS: SettingsSection[] = [
  { href: '/settings/account',      icon: 'userCog',  label: 'Your account',          description: 'Display name, username, email' },
  { href: '/settings/privacy',      icon: 'lock',     label: 'Privacy and safety',    description: 'Control who sees your posts' },
  { href: '/settings/appearance',   icon: 'palette',  label: 'Appearance',            description: 'Light or dark theme' },
  { href: '/settings/password',     icon: 'key',      label: 'Password',              description: 'Change account password' },
  { href: '/settings/blocked-users', icon: 'userX',   label: 'Blocked accounts',      description: 'People you’ve blocked' },
  { href: '/settings/muted-users',  icon: 'volumeX',  label: 'Muted accounts',        description: 'People you’ve muted' },
  { href: '/settings/subscription', icon: 'premium',  label: 'Premium',               description: 'Manage your subscription' },
];

export default function SettingsLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const isIndex = pathname === '/settings';

  return (
    <div className={`settings-shell${isIndex ? ' is-index' : ' is-detail'}`} style={{ display: 'flex', minHeight: '100vh' }}>
      {/* Section nav — sticky, scrolls only on overflow */}
      <aside
        className="settings-aside"
        style={{
          width: 360,
          flexShrink: 0,
          borderRight: '1px solid var(--border)',
          position: 'sticky',
          top: 0,
          alignSelf: 'flex-start',
          height: '100vh',
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        <div
          style={{
            padding: '14px 20px',
            borderBottom: '1px solid var(--border)',
            background: 'var(--bg)',
          }}
        >
          <h1 style={{ fontWeight: 700, fontSize: 20, color: 'var(--text)', letterSpacing: '-0.02em', margin: 0 }}>
            Settings
          </h1>
        </div>
        <nav style={{ padding: '8px 0' }}>
          {SECTIONS.map((s) => {
            const active = pathname === s.href || pathname.startsWith(s.href + '/');
            return <SettingsNavItem key={s.href} section={s} active={active} />;
          })}
        </nav>
      </aside>

      {/* Detail pane */}
      <section className="settings-detail" style={{ flex: 1, minWidth: 0, borderLeft: '1px solid var(--border)' }}>
        {isIndex ? <SettingsIndexHint /> : children}
      </section>
    </div>
  );
}

function SettingsNavItem({ section, active }: { section: SettingsSection; active: boolean }) {
  return (
    <Link
      href={section.href}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 14,
        padding: '14px 20px',
        textDecoration: 'none',
        background: active ? 'var(--bg-hover)' : 'transparent',
        borderLeft: active ? '3px solid var(--accent)' : '3px solid transparent',
        transition: 'background 0.12s ease',
      }}
      onMouseEnter={(e) => { if (!active) e.currentTarget.style.background = 'var(--bg-hover)'; }}
      onMouseLeave={(e) => { if (!active) e.currentTarget.style.background = 'transparent'; }}
    >
      {section.icon === 'premium' ? (
        <VerifiedBadge size={20} variant="mono" />
      ) : (
        <Icon name={section.icon} size={20} color="var(--text)" strokeWidth={1.8} />
      )}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 15, fontWeight: 600, color: 'var(--text)', letterSpacing: '-0.01em' }}>
          {section.label}
        </div>
        <div style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 2 }}>
          {section.description}
        </div>
      </div>
      <Icon name="chevronRight" size={16} color="var(--text-tertiary)" />
    </Link>
  );
}

function SettingsIndexHint() {
  return (
    <div
      style={{
        height: '100%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 40,
        color: 'var(--text-tertiary)',
        fontSize: 14,
      }}
    >
      Select a setting on the left.
    </div>
  );
}
