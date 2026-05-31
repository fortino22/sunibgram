import Link from 'next/link';
import { Crown, Shield, User } from 'lucide-react';
import type { ReactNode } from 'react';

type Role = {
  icon: ReactNode;
  label: string;
  community: string;
  tone: 'owner' | 'mod' | 'member';
};

const ROLES: Role[] = [
  { icon: <Crown size={14} />,  label: 'OWNER',     community: '@bnpc-club',    tone: 'owner'  },
  { icon: <Shield size={14} />, label: 'MODERATOR', community: '@design-binus', tone: 'mod'    },
  { icon: <User size={14} />,   label: 'MEMBER',    community: '@anime-binus',  tone: 'member' },
];

const TONE: Record<Role['tone'], { bg: string; fg: string }> = {
  owner:  { bg: 'var(--amber-soft)', fg: 'var(--amber)' },
  mod:    { bg: 'var(--accent-soft)', fg: 'var(--accent)' },
  member: { bg: 'var(--green-soft)', fg: 'var(--green)' },
};

export function AboutSplitRoles() {
  return (
    <section
      id="organizers"
      className="about-container"
      style={{
        paddingTop: '96px',
        paddingBottom: '96px',
      }}
    >
      <div
        className="grid gap-16 items-center"
        style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))' }}
      >
        {/* Left — copy */}
        <div style={{ maxWidth: '34rem' }}>
          <span className="about-eyebrow">Roles move with you</span>
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
            Owner here. Mod there. Member everywhere.
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
            You don&apos;t pick one identity. Run your own community, help
            moderate another, just lurk in a third — all from the same Binus
            account. No juggling logins.
          </p>
          <div style={{ marginTop: '28px' }}>
            <Link href="/signup" className="about-link-mono">
              Start your first community →
            </Link>
          </div>
        </div>

        {/* Right — avatar branching to three role badges */}
        <div
          style={{
            justifySelf: 'end',
            width: '100%',
            maxWidth: '32rem',
            position: 'relative',
          }}
        >
          <div className="flex items-center gap-8">
            {/* Avatar */}
            <div
              className="about-surface-card"
              style={{
                padding: '14px 16px',
                display: 'inline-flex',
                alignItems: 'center',
                gap: 12,
                boxShadow: '0 8px 22px oklch(0% 0 0 / 0.06)',
              }}
            >
              <span
                style={{
                  width: 36,
                  height: 36,
                  borderRadius: 999,
                  background:
                    'linear-gradient(135deg, var(--accent) 0%, oklch(58% 0.18 220) 100%)',
                }}
              />
              <span style={{ display: 'inline-flex', flexDirection: 'column' }}>
                <span
                  style={{
                    color: 'var(--text)',
                    fontWeight: 600,
                    fontSize: '0.875rem',
                  }}
                >
                  @you
                </span>
                <span
                  style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}
                >
                  School of Design
                </span>
              </span>
            </div>

            {/* Branch + role badges */}
            <div style={{ flex: 1, position: 'relative' }}>
              <svg
                aria-hidden
                viewBox="0 0 120 160"
                preserveAspectRatio="none"
                style={{
                  position: 'absolute',
                  inset: 0,
                  width: '100%',
                  height: '100%',
                  color: 'var(--border-strong)',
                }}
              >
                <path d="M0 80 C 40 80, 50 20, 120 20" stroke="currentColor" strokeWidth="1" fill="none" />
                <path d="M0 80 C 40 80, 60 80, 120 80" stroke="currentColor" strokeWidth="1" fill="none" />
                <path d="M0 80 C 40 80, 50 140, 120 140" stroke="currentColor" strokeWidth="1" fill="none" />
              </svg>

              <div
                style={{
                  position: 'relative',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 16,
                }}
              >
                {ROLES.map((r) => {
                  const t = TONE[r.tone];
                  return (
                    <div
                      key={r.label}
                      className="about-surface-card"
                      style={{
                        padding: '10px 14px',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 10,
                        alignSelf: 'flex-end',
                      }}
                    >
                      <span
                        style={{
                          display: 'inline-flex',
                          alignItems: 'center',
                          gap: 6,
                          fontSize: '0.6875rem',
                          letterSpacing: '0.08em',
                          padding: '3px 8px',
                          borderRadius: 4,
                          background: t.bg,
                          color: t.fg,
                          fontWeight: 600,
                        }}
                      >
                        {r.icon}
                        {r.label}
                      </span>
                      <span
                        style={{
                          fontSize: '0.8125rem',
                          color: 'var(--text-muted)',
                        }}
                      >
                        {r.community}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
