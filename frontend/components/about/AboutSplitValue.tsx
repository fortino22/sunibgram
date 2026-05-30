import Link from 'next/link';
import { MessageSquare, Users, Calendar, Bookmark, Check, Heart } from 'lucide-react';
import type { ReactNode } from 'react';

type ActivityRow = {
  done: boolean;
  icons: ReactNode;
  label: string;
};

const ACTIVITY: ActivityRow[] = [
  {
    done: true,
    icons: <IconStack icons={[<MessageSquare key="m" size={12} />, <Heart key="h" size={12} />]} />,
    label: 'Replied to a thread in @design-binus',
  },
  {
    done: true,
    icons: <IconStack icons={[<Users key="u" size={12} />, <Check key="c" size={12} />]} />,
    label: 'Joined the BNPC community',
  },
  {
    done: false,
    icons: <IconStack icons={[<Calendar key="c" size={12} />, <MessageSquare key="m" size={12} />]} />,
    label: 'RSVP to BINUS Job Fair 2026',
  },
  {
    done: false,
    icons: <IconStack icons={[<Bookmark key="b" size={12} />, <MessageSquare key="m" size={12} />]} />,
    label: 'Save 3 posts to my “Finals prep” collection',
  },
  {
    done: false,
    icons: <IconStack icons={[<Users key="u" size={12} />, <MessageSquare key="m" size={12} />]} />,
    label: 'Follow back the people I matched with',
  },
  {
    done: false,
    icons: <IconStack icons={[<Calendar key="c" size={12} />, <Users key="u" size={12} />]} />,
    label: 'Post the rundown for next club meeting',
  },
];

function IconStack({ icons }: { icons: ReactNode[] }) {
  return (
    <span style={{ display: 'inline-flex', gap: 0 }}>
      {icons.map((ic, i) => (
        <span
          key={i}
          style={{
            width: 22,
            height: 22,
            borderRadius: '50%',
            background: 'var(--bg)',
            border: '1px solid var(--border)',
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--text-muted)',
            marginLeft: i === 0 ? 0 : -6,
          }}
        >
          {ic}
        </span>
      ))}
    </span>
  );
}

export function AboutSplitValue() {
  return (
    <section
      className="about-container"
      style={{ paddingTop: '96px', paddingBottom: '96px' }}
    >
      <div
        className="grid gap-16 items-start"
        style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))' }}
      >
        {/* Left — copy */}
        <div style={{ maxWidth: '34rem' }}>
          <span className="about-eyebrow">One campus, one feed</span>
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
            Everything campus, in one place.
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
            Pin a thread from your faculty. Catch up on a community.
            RSVP to a campus event before it&apos;s full. All in one app,
            all from people you actually share a campus with.
          </p>
          <div style={{ marginTop: '28px' }}>
            <Link href="/signup" className="about-link-mono">
              See what&apos;s happening on campus →
            </Link>
          </div>
        </div>

        {/* Right — campus activity card */}
        <div
          className="about-surface-card"
          style={{
            padding: '20px',
            maxWidth: '34rem',
            justifySelf: 'end',
            width: '100%',
          }}
        >
          <div
            className="flex items-center justify-between"
            style={{ paddingBottom: '14px', borderBottom: '1px solid var(--border)' }}
          >
            <span
              style={{
                fontSize: '0.6875rem',
                textTransform: 'uppercase',
                letterSpacing: '0.08em',
                color: 'var(--text-muted)',
                fontWeight: 600,
              }}
            >
              ☑ Today on campus
            </span>
            <span
              style={{
                fontSize: '0.6875rem',
                color: 'var(--text-tertiary)',
                fontWeight: 600,
              }}
            >
              02 / 06
            </span>
          </div>

          <ul style={{ marginTop: '12px', display: 'flex', flexDirection: 'column' }}>
            {ACTIVITY.map((row, i) => (
              <li
                key={i}
                className="flex items-center gap-3"
                style={{
                  padding: '10px 0',
                  borderBottom: i < ACTIVITY.length - 1 ? '1px solid var(--border)' : 'none',
                }}
              >
                <span
                  style={{
                    width: 18,
                    height: 18,
                    borderRadius: '50%',
                    background: row.done ? 'var(--green)' : 'transparent',
                    border: row.done ? 'none' : '1.5px solid var(--border-strong)',
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}
                >
                  {row.done && <Check size={11} color="#fff" strokeWidth={3} />}
                </span>
                {row.icons}
                <span
                  style={{
                    fontSize: '0.875rem',
                    color: row.done ? 'var(--text-tertiary)' : 'var(--text)',
                    textDecoration: row.done ? 'line-through' : 'none',
                    flex: 1,
                    lineHeight: 1.4,
                  }}
                >
                  {row.label}
                </span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}
