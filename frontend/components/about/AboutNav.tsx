import Link from 'next/link';
import { Logo } from '@/components/ui/Logo';

const NAV_LINKS = [
  { label: 'What it is', href: '#what' },
  { label: 'Use cases', href: '#use-cases' },
  { label: 'For organizers', href: '#organizers' },
  { label: 'The team', href: '#team' },
] as const;

export function AboutNav() {
  return (
    <nav
      style={{
        background: 'var(--bg)',
        borderBottom: '1px solid var(--border)',
        position: 'sticky',
        top: 0,
        zIndex: 50,
      }}
    >
      <div className="about-container h-16 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link href="/" className="flex items-center gap-2.5 hover:opacity-80 transition-opacity">
            <Logo size={26} />
            <span
              style={{
                color: 'var(--text)',
                fontWeight: 700,
                fontSize: '1rem',
                letterSpacing: '-0.01em',
              }}
            >
              Sunibgram
            </span>
          </Link>
          <span
            className="mono"
            style={{
              fontSize: '0.6875rem',
              padding: '3px 8px',
              border: '1px solid var(--border)',
              borderRadius: '4px',
              color: 'var(--accent)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              fontWeight: 600,
            }}
          >
            Binus only
          </span>
        </div>

        <div className="hidden md:flex items-center gap-7">
          {NAV_LINKS.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className="mono hover:opacity-70 transition-opacity"
              style={{
                color: 'var(--text)',
                fontSize: '0.8125rem',
                textTransform: 'uppercase',
                letterSpacing: '0.06em',
                fontWeight: 600,
              }}
            >
              {link.label}
            </Link>
          ))}
          <Link
            href="/login"
            className="mono hover:opacity-70 transition-opacity"
            style={{
              color: 'var(--text)',
              fontSize: '0.8125rem',
              textTransform: 'uppercase',
              letterSpacing: '0.06em',
              fontWeight: 600,
            }}
          >
            Sign in
          </Link>
        </div>

        <Link href="/signup" className="about-btn-primary">
          Join Sunibgram
        </Link>
      </div>
    </nav>
  );
}
