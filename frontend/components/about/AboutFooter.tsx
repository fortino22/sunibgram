import Link from 'next/link';
import { Logo } from '@/components/ui/Logo';

type Column = {
  title: string;
  links: { label: string; href: string }[];
};

const COLUMNS: Column[] = [
  {
    title: 'The app',
    links: [
      { label: 'Threads', href: '/' },
      { label: 'Communities', href: '/communities' },
      { label: 'Events', href: '/events' },
      { label: 'Direct messages', href: '/messages' },
      { label: 'Leaderboard', href: '/leaderboard' },
      { label: 'Find people', href: '/connect' },
      { label: 'Bookmarks', href: '/bookmarks' },
      { label: 'Premium', href: '/settings/subscription' },
    ],
  },
  {
    title: 'Join Sunibgram',
    links: [
      { label: 'Sign up', href: '/signup' },
      { label: 'Sign in', href: '/login' },
      { label: 'Forgot password', href: '/forgot-password' },
      { label: 'Explore the feed', href: '/explore' },
    ],
  },
  {
    title: 'For organizers',
    links: [
      { label: 'Start a community', href: '/communities' },
      { label: 'Host an event', href: '/events' },
      { label: 'Become a moderator', href: '/communities' },
      { label: 'Admin tools', href: '/admin' },
    ],
  },
  {
    title: 'The team',
    links: [
      { label: 'About', href: '/about' },
      { label: 'Talk to us', href: '/feedback' },
      { label: 'Report a bug', href: '/feedback' },
      { label: 'Request a feature', href: '/feedback' },
    ],
  },
  {
    title: 'Fine print',
    links: [
      { label: 'Trust & safety', href: '/about' },
      { label: 'Privacy', href: '/about' },
      { label: 'Terms', href: '/about' },
      { label: 'Community rules', href: '/about' },
    ],
  },
];

export function AboutFooter() {
  return (
    <footer style={{ background: 'var(--bg)', color: 'var(--text)' }}>
      <div
        className="about-container"
        style={{ paddingTop: '64px', paddingBottom: '40px' }}
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

        <div
          className="flex items-center justify-between flex-wrap gap-4"
          style={{
            marginTop: '56px',
            paddingTop: '24px',
            borderTop: '1px solid var(--border)',
          }}
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
