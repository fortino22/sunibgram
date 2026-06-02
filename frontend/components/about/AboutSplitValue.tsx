'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { MessageSquare, Users, Calendar, Bookmark, Check, Heart } from 'lucide-react';
import type { ReactNode } from 'react';

type ActivityRow = {
  icons: ReactNode;
  label: string;
};

const ACTIVITY: ActivityRow[] = [
  {
    icons: <IconStack icons={[<MessageSquare key="m" size={12} />, <Heart key="h" size={12} />]} />,
    label: 'Replied to a thread in @design-binus',
  },
  {
    icons: <IconStack icons={[<Users key="u" size={12} />, <Check key="c" size={12} />]} />,
    label: 'Joined the BNPC community',
  },
  {
    icons: <IconStack icons={[<Calendar key="c" size={12} />, <MessageSquare key="m" size={12} />]} />,
    label: 'RSVP to BINUS Job Fair 2026',
  },
  {
    icons: <IconStack icons={[<Bookmark key="b" size={12} />, <MessageSquare key="m" size={12} />]} />,
    label: 'Saved 3 posts to my "Finals prep" collection',
  },
  {
    icons: <IconStack icons={[<Users key="u" size={12} />, <MessageSquare key="m" size={12} />]} />,
    label: 'Followed back the people I matched with',
  },
  {
    icons: <IconStack icons={[<Calendar key="c" size={12} />, <Users key="u" size={12} />]} />,
    label: 'Posted the rundown for next club meeting',
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
  /* Cycle through checking off items one by one, pause, then reset. */
  const [checked, setChecked] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setChecked(0);
      for (let i = 1; i <= ACTIVITY.length; i++) {
        timeouts.push(setTimeout(() => setChecked(i), 800 + i * 700));
      }
      timeouts.push(setTimeout(cycle, 800 + (ACTIVITY.length + 1) * 700 + 1800));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

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
              fontWeight: 500,
              fontSize: 'clamp(1.75rem, 3.5vw, 2.75rem)',
              lineHeight: 1.12,
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
            <Link href="/signup" className="about-btn-dark">
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
            <motion.span
              key={checked}
              initial={{ opacity: 0, y: -4 }}
              animate={{ opacity: 1, y: 0 }}
              style={{
                fontSize: '0.6875rem',
                color: 'var(--text-tertiary)',
                fontWeight: 600,
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              {String(checked).padStart(2, '0')} / {String(ACTIVITY.length).padStart(2, '0')}
            </motion.span>
          </div>

          <ul style={{ marginTop: '12px', display: 'flex', flexDirection: 'column' }}>
            {ACTIVITY.map((row, i) => {
              const done = i < checked;
              return (
                <li
                  key={i}
                  className="flex items-center gap-3"
                  style={{
                    padding: '10px 0',
                    borderBottom: i < ACTIVITY.length - 1 ? '1px solid var(--border)' : 'none',
                  }}
                >
                  <motion.span
                    animate={{
                      background: done ? 'var(--green)' : 'rgba(0,0,0,0)',
                      borderColor: done ? 'rgba(0,0,0,0)' : 'var(--border-strong)',
                    }}
                    transition={{ duration: 0.3 }}
                    style={{
                      width: 18,
                      height: 18,
                      borderRadius: '50%',
                      border: '1.5px solid var(--border-strong)',
                      display: 'inline-flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      flexShrink: 0,
                    }}
                  >
                    <AnimatePresence>
                      {done && (
                        <motion.span
                          key="check"
                          initial={{ scale: 0, rotate: -90 }}
                          animate={{ scale: 1, rotate: 0 }}
                          exit={{ scale: 0, opacity: 0 }}
                          transition={{ type: 'spring', stiffness: 400, damping: 22 }}
                          style={{ display: 'inline-flex' }}
                        >
                          <Check size={11} color="#fff" strokeWidth={3} />
                        </motion.span>
                      )}
                    </AnimatePresence>
                  </motion.span>
                  {row.icons}
                  <motion.span
                    animate={{
                      color: done ? 'var(--text-tertiary)' : 'var(--text)',
                    }}
                    transition={{ duration: 0.3 }}
                    style={{
                      fontSize: '0.875rem',
                      textDecoration: done ? 'line-through' : 'none',
                      flex: 1,
                      lineHeight: 1.4,
                    }}
                  >
                    {row.label}
                  </motion.span>
                </li>
              );
            })}
          </ul>
        </div>
      </div>
    </section>
  );
}
