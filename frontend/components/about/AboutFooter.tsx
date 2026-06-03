import Link from 'next/link';
import { Logo } from '@/components/ui/Logo';

type Column = {
  title: string;
  links: { label: string; href: string }[];
};

const COLUMNS: Column[] = [
  {
    title: 'Features',
    links: [
      { label: 'Threads', href: '/about/threads' },
      { label: 'Communities', href: '/about/communities' },
      { label: 'Messages', href: '/about/messages' },
      { label: 'Events', href: '/about/events' },
      { label: 'Leaderboard', href: '/about/leaderboard' },
      { label: 'Fox AI', href: '/about/fox' },
    ],
  },
  {
    title: 'Our Story',
    links: [
      { label: 'Mission', href: '/about/mission' },
      { label: 'Story', href: '/about/story' },
      { label: 'Use cases', href: '/about/use-cases' },
      { label: 'News', href: '/about/news' },
    ],
  },
  {
    title: 'Resources',
    links: [
      { label: 'Help center', href: '/about/help' },
      { label: 'Algorithm', href: '/about/algorithm' },
      { label: 'How we operate', href: '/about/hermes' },
    ],
  },
  {
    title: 'Safety',
    links: [
      { label: 'Privacy center', href: '/about/privacy' },
      { label: 'Terms and conditions', href: '/about/terms' },
    ],
  },
  {
    title: 'Get started',
    links: [
      { label: 'Sign up', href: '/signup' },
      { label: 'Sign in', href: '/login' },
      { label: 'Forgot password', href: '/forgot-password' },
      { label: 'Explore the feed', href: '/explore' },
    ],
  },
];

export function AboutFooter() {
  return (
    <footer
      style={{
        background: 'var(--bg)',
        color: 'var(--text)',
        borderTop: '1px solid var(--border)',
      }}
    >
      <div
        className="about-container"
        style={{ paddingTop: '64px', paddingBottom: '56px' }}
      >
        <div className="flex items-center gap-2.5" style={{ marginBottom: '48px' }}>
          <Logo size={35} />
          <span style={{ fontWeight: 700, fontSize: '1.35rem', letterSpacing: '-0.01em' }}>
            Sunibgram
          </span>
        </div>

        <div
          className="grid gap-10"
          style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))' }}
        >
          {COLUMNS.map((col) => (
            <div key={col.title}>
              <div
                style={{
                  fontSize: '0.6875rem',
                  textTransform: 'uppercase',
                  letterSpacing: '0.1em',
                  color: 'var(--text-muted)',
                  marginBottom: '18px',
                  fontWeight: 600,
                }}
              >
                {col.title}
              </div>
              <ul style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {col.links.map((link) => (
                  <li key={link.label}>
                    <Link
                      href={link.href}
                      style={{
                        color: 'var(--text)',
                        fontSize: '0.875rem',
                        opacity: 0.85,
                      }}
                      className="hover:opacity-60 transition-opacity"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>

      <div style={{ borderTop: '1px solid var(--border)' }}>
        <div
          className="about-container flex items-center justify-between flex-wrap gap-4"
          style={{ paddingTop: '24px', paddingBottom: '24px' }}
        >
          <p
            style={{
              color: 'var(--text-muted)',
              fontSize: '0.8125rem',
              maxWidth: '36rem',
              lineHeight: 1.55,
            }}
          >
            Sunibgram is a student-built project — not officially affiliated with
            Binus University. Made in Jakarta with a lot of coffee.
          </p>
          <span
            style={{
              fontSize: '0.75rem',
              letterSpacing: '0.06em',
              color: 'var(--text-muted)',
              textTransform: 'uppercase',
              fontWeight: 600,
            }}
          >
            © Sunibgram {new Date().getFullYear()}
          </span>
        </div>
      </div>
    </footer>
  );
}
