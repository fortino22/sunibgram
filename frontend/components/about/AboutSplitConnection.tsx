import Link from 'next/link';
import {
  MessageSquare,
  MessageCircle,
  Calendar,
  Users,
  Trophy,
  Bookmark,
  User,
  GraduationCap,
} from 'lucide-react';
import type { ReactNode } from 'react';

type Chip = {
  icon: ReactNode;
  label: string;
  top: string;
  left: string;
};

const CHIPS: Chip[] = [
  { icon: <Calendar size={14} />, label: 'Events', top: '4%', left: '38%' },
  { icon: <MessageSquare size={14} />, label: 'Threads', top: '12%', left: '74%' },
  { icon: <User size={14} />, label: 'Profile', top: '22%', left: '6%' },
  { icon: <MessageCircle size={14} />, label: 'DMs', top: '38%', left: '80%' },
  { icon: <Bookmark size={14} />, label: 'Collections', top: '46%', left: '0%' },
  { icon: <Users size={14} />, label: 'Communities', top: '64%', left: '66%' },
  { icon: <Trophy size={14} />, label: 'Leaderboard', top: '78%', left: '18%' },
  { icon: <GraduationCap size={14} />, label: 'Faculty', top: '88%', left: '52%' },
];

export function AboutSplitConnection() {
  return (
    <section
      className="about-container"
      style={{
        paddingTop: '96px',
        paddingBottom: '96px',
        borderTop: '1px solid var(--border)',
      }}
    >
      <div
        className="grid gap-16 items-center"
        style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))' }}
      >
        {/* Left — copy */}
        <div style={{ maxWidth: '34rem' }}>
          <span className="about-eyebrow">Verified · one account</span>
          <h2
            style={{
              color: 'var(--text)',
              fontWeight: 600,
              fontSize: 'clamp(2rem, 4vw, 3rem)',
              lineHeight: 1.1,
              letterSpacing: '-0.02em',
              marginTop: '20px',
            }}
          >
            One sign-in. Everything Binus.
          </h2>
          <p
            style={{
              color: 'var(--text-muted)',
              fontSize: '1rem',
              lineHeight: 1.65,
              marginTop: '20px',
              maxWidth: '32rem',
            }}
          >
            Sign in once with your <span style={{ color: 'var(--text)', fontWeight: 600 }}>@binus.ac.id</span>.
            Your faculty, your profile, your DMs, your communities — they all
            know it&apos;s really you. No second account, no second password.
          </p>
          <div style={{ marginTop: '28px' }}>
            <Link href="/signup" className="about-link-mono">
              Create your account →
            </Link>
          </div>
        </div>

        {/* Right — dotted sphere with floating chips */}
        <div
          style={{
            position: 'relative',
            width: '100%',
            maxWidth: '560px',
            aspectRatio: '1 / 1',
            justifySelf: 'end',
          }}
        >
          <div
            aria-hidden
            style={{
              position: 'absolute',
              inset: '8%',
              borderRadius: '50%',
              backgroundImage:
                'radial-gradient(color-mix(in oklab, var(--border-strong) 70%, transparent) 1px, transparent 1px)',
              backgroundSize: '14px 14px',
              maskImage: 'radial-gradient(circle, black 55%, transparent 100%)',
              WebkitMaskImage: 'radial-gradient(circle, black 55%, transparent 100%)',
              opacity: 0.85,
            }}
          />

          {CHIPS.map((c) => (
            <div
              key={c.label}
              style={{
                position: 'absolute',
                top: c.top,
                left: c.left,
                display: 'inline-flex',
                alignItems: 'center',
                gap: 8,
                padding: '8px 14px',
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 999,
                boxShadow: '0 6px 18px oklch(0% 0 0 / 0.06)',
                fontSize: '0.8125rem',
                color: 'var(--text)',
                fontWeight: 600,
                whiteSpace: 'nowrap',
              }}
            >
              <span style={{ color: 'var(--accent)' }}>{c.icon}</span>
              {c.label}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
