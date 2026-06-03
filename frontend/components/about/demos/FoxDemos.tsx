'use client';

import { useEffect, useState, type ReactNode } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import {
  CheckCircle2,
  Calendar,
  MapPin,
  Bug,
  Sparkles,
  TrendingUp,
  Clock,
  AlertTriangle,
  Trophy,
  Hash,
} from 'lucide-react';

/* ------------------------------------------------------------------ */
/*  Shared demo shell                                                  */
/* ------------------------------------------------------------------ */

function DemoShell({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="sg-demo-shell">
      <div className="sg-demo-shell-header">
        <span>{title}</span>
        <span style={{ fontSize: 12, color: 'var(--text-tertiary)', fontWeight: 500 }}>
          Fox · live demo
        </span>
      </div>
      <div className="sg-demo-shell-body">{children}</div>
    </div>
  );
}

function FoxBadge() {
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        padding: '4px 10px',
        borderRadius: 999,
        background: 'var(--accent-soft)',
        color: 'var(--accent)',
        fontSize: 11,
        fontWeight: 700,
        letterSpacing: '0.06em',
        textTransform: 'uppercase',
      }}
    >
      <Sparkles size={12} strokeWidth={2.5} />
      Fox
    </span>
  );
}

/* ------------------------------------------------------------------ */
/*  1 — Event review                                                   */
/* ------------------------------------------------------------------ */

type Phase = 'submitted' | 'reviewing' | 'verdict';

export function FoxDemoEventReview() {
  const [phase, setPhase] = useState<Phase>('submitted');
  const [confidence, setConfidence] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setPhase('submitted');
      setConfidence(0);
      timeouts.push(setTimeout(() => setPhase('reviewing'), 900));
      timeouts.push(
        setTimeout(() => {
          setPhase('verdict');
          setConfidence(92);
        }, 2600),
      );
      timeouts.push(setTimeout(cycle, 6800));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  return (
    <DemoShell title="Event review queue">
      <div style={{ padding: 18, display: 'flex', flexDirection: 'column', gap: 14 }}>
        {/* Event proposal card */}
        <div
          style={{
            border: '1px solid var(--border)',
            borderRadius: 12,
            background: 'var(--bg-card)',
            padding: '14px 16px',
          }}
        >
          <div
            style={{
              fontSize: 11,
              fontWeight: 700,
              letterSpacing: '0.08em',
              color: 'var(--text-tertiary)',
            }}
          >
            OFFICIAL · SUBMITTED
          </div>
          <div
            style={{
              marginTop: 6,
              fontWeight: 700,
              fontSize: 15,
              color: 'var(--text)',
              letterSpacing: '-0.01em',
            }}
          >
            Design Critique Night
          </div>
          <div
            style={{
              marginTop: 8,
              display: 'flex',
              gap: 14,
              fontSize: 12.5,
              color: 'var(--text-muted)',
              flexWrap: 'wrap',
            }}
          >
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
              <Calendar size={13} />
              Fri · 7:00 PM
            </span>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
              <MapPin size={13} />
              Anggrek, room 502
            </span>
          </div>
        </div>

        {/* Fox status / verdict */}
        <AnimatePresence mode="wait">
          {phase === 'submitted' && (
            <motion.div
              key="waiting"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              style={{
                padding: 12,
                fontSize: 13,
                color: 'var(--text-tertiary)',
                textAlign: 'center',
              }}
            >
              Waiting in queue…
            </motion.div>
          )}

          {phase === 'reviewing' && (
            <motion.div
              key="reviewing"
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0 }}
              style={{
                border: '1px solid var(--border)',
                borderRadius: 12,
                padding: '14px 16px',
                background: 'var(--bg-subtle)',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
              }}
            >
              <FoxBadge />
              <span style={{ fontSize: 13, color: 'var(--text)' }}>
                Reading proposal
                <motion.span
                  animate={{ opacity: [0.2, 1, 0.2] }}
                  transition={{ duration: 1.2, repeat: Infinity }}
                  style={{ marginLeft: 4 }}
                >
                  …
                </motion.span>
              </span>
            </motion.div>
          )}

          {phase === 'verdict' && (
            <motion.div
              key="verdict"
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.35 }}
              style={{
                border: '1px solid var(--green, #16a34a)',
                borderRadius: 12,
                padding: '14px 16px',
                background: 'var(--bg-card)',
              }}
            >
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  flexWrap: 'wrap',
                }}
              >
                <FoxBadge />
                <span
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: 6,
                    color: 'var(--green, #16a34a)',
                    fontWeight: 700,
                    fontSize: 14,
                    letterSpacing: '0.04em',
                  }}
                >
                  <CheckCircle2 size={16} />
                  APPROVED
                </span>
                <span
                  style={{
                    marginLeft: 'auto',
                    fontSize: 12,
                    color: 'var(--text-tertiary)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  Confidence {confidence}%
                </span>
              </div>

              {/* Confidence bar */}
              <div
                style={{
                  marginTop: 10,
                  height: 6,
                  borderRadius: 999,
                  background: 'var(--bg-subtle)',
                  overflow: 'hidden',
                }}
              >
                <motion.div
                  initial={{ width: 0 }}
                  animate={{ width: `${confidence}%` }}
                  transition={{ duration: 0.9, ease: 'easeOut' }}
                  style={{
                    height: '100%',
                    background: 'var(--green, #16a34a)',
                  }}
                />
              </div>

              {/* Reason */}
              <p
                style={{
                  marginTop: 12,
                  fontSize: 13,
                  lineHeight: 1.55,
                  color: 'var(--text-muted)',
                }}
              >
                Matches event policy: clear organizer, on-campus venue, no
                external sponsors flagged.
              </p>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  2 — Daily report (morning brief)                                   */
/* ------------------------------------------------------------------ */

type Stat = { label: string; target: number; icon: ReactNode };

const DAILY_STATS: Stat[] = [
  { label: 'Events decided', target: 12, icon: <Calendar size={14} /> },
  { label: 'Auto-approved', target: 9, icon: <CheckCircle2 size={14} /> },
  { label: 'Open bugs', target: 4, icon: <Bug size={14} /> },
  { label: 'Feature requests', target: 7, icon: <Sparkles size={14} /> },
];

const DAILY_ROWS = [
  { title: 'Coding Night', verdict: 'Approve', confidence: 94 },
  { title: 'Charity Run', verdict: 'Ask for changes', confidence: 71 },
  { title: 'Off-campus mixer', verdict: 'Decline', confidence: 88 },
];

export function FoxDemoDailyReport() {
  const [counts, setCounts] = useState<number[]>([0, 0, 0, 0]);
  const [visibleRows, setVisibleRows] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setCounts([0, 0, 0, 0]);
      setVisibleRows(0);

      DAILY_STATS.forEach((stat, i) => {
        const steps = 20;
        for (let s = 1; s <= steps; s++) {
          timeouts.push(
            setTimeout(() => {
              setCounts((prev) => {
                const next = [...prev];
                next[i] = Math.round((stat.target * s) / steps);
                return next;
              });
            }, 300 + s * 40),
          );
        }
      });

      timeouts.push(setTimeout(() => setVisibleRows(1), 1600));
      timeouts.push(setTimeout(() => setVisibleRows(2), 2100));
      timeouts.push(setTimeout(() => setVisibleRows(3), 2600));
      timeouts.push(setTimeout(cycle, 6500));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  return (
    <DemoShell title="Daily brief · Mon morning">
      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 14 }}>
        {/* At-a-glance */}
        <div
          style={{
            borderLeft: '3px solid var(--accent)',
            paddingLeft: 12,
            fontSize: 13,
            color: 'var(--text-muted)',
            lineHeight: 1.55,
          }}
        >
          12 event calls overnight, 1 bug spiking in the chat module. Top
          feature ask: dark mode for the leaderboard.
        </div>

        {/* Stat cards */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(2, 1fr)',
            gap: 8,
          }}
        >
          {DAILY_STATS.map((stat, i) => (
            <div
              key={stat.label}
              style={{
                border: '1px solid var(--border)',
                borderRadius: 10,
                background: 'var(--bg-card)',
                padding: '10px 12px',
              }}
            >
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 6,
                  color: 'var(--text-tertiary)',
                  fontSize: 11,
                  fontWeight: 600,
                  letterSpacing: '0.04em',
                }}
              >
                {stat.icon}
                {stat.label}
              </div>
              <div
                style={{
                  marginTop: 4,
                  fontSize: 22,
                  fontWeight: 700,
                  color: 'var(--text)',
                  fontVariantNumeric: 'tabular-nums',
                  letterSpacing: '-0.02em',
                }}
              >
                {counts[i]}
              </div>
            </div>
          ))}
        </div>

        {/* Event decisions table */}
        <div
          style={{
            border: '1px solid var(--border)',
            borderRadius: 10,
            overflow: 'hidden',
            background: 'var(--bg-card)',
          }}
        >
          <div
            style={{
              padding: '8px 12px',
              fontSize: 11,
              fontWeight: 700,
              letterSpacing: '0.08em',
              color: 'var(--text-tertiary)',
              borderBottom: '1px solid var(--border)',
              background: 'var(--bg-subtle)',
            }}
          >
            EVENT DECISIONS
          </div>
          <AnimatePresence>
            {DAILY_ROWS.slice(0, visibleRows).map((row) => {
              const ok = row.verdict === 'Approve';
              const warn = row.verdict === 'Ask for changes';
              const color = ok
                ? 'var(--green, #16a34a)'
                : warn
                ? 'var(--accent)'
                : 'var(--coral, #ef4444)';
              return (
                <motion.div
                  key={row.title}
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.25 }}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                    padding: '10px 12px',
                    borderTop: '1px solid var(--border)',
                    fontSize: 13,
                  }}
                >
                  <span style={{ color: 'var(--text)', fontWeight: 600, flex: 1 }}>
                    {row.title}
                  </span>
                  <span
                    style={{
                      color,
                      fontWeight: 700,
                      fontSize: 11.5,
                      letterSpacing: '0.04em',
                      textTransform: 'uppercase',
                    }}
                  >
                    {row.verdict}
                  </span>
                  <span
                    style={{
                      color: 'var(--text-tertiary)',
                      fontSize: 11.5,
                      fontVariantNumeric: 'tabular-nums',
                      width: 36,
                      textAlign: 'right',
                    }}
                  >
                    {row.confidence}%
                  </span>
                </motion.div>
              );
            })}
          </AnimatePresence>
        </div>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  3 — Weekly report (story of the week)                              */
/* ------------------------------------------------------------------ */

const HASHTAGS = [
  { tag: '#bnpc', target: 88 },
  { tag: '#anggrekkantin', target: 64 },
  { tag: '#careerfair', target: 51 },
  { tag: '#midterms', target: 39 },
];

const LEADERS = [
  { name: 'Kevin', engagements: 412 },
  { name: 'Rina', engagements: 318 },
  { name: 'Maya', engagements: 244 },
];

export function FoxDemoWeeklyReport() {
  const [pcts, setPcts] = useState<number[]>([0, 0, 0, 0]);
  const [leaders, setLeaders] = useState<number[]>([0, 0, 0]);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setPcts([0, 0, 0, 0]);
      setLeaders([0, 0, 0]);

      HASHTAGS.forEach((_, i) => {
        timeouts.push(
          setTimeout(() => {
            setPcts((prev) => {
              const next = [...prev];
              next[i] = HASHTAGS[i].target;
              return next;
            });
          }, 500 + i * 220),
        );
      });

      LEADERS.forEach((leader, i) => {
        const steps = 18;
        for (let s = 1; s <= steps; s++) {
          timeouts.push(
            setTimeout(() => {
              setLeaders((prev) => {
                const next = [...prev];
                next[i] = Math.round((leader.engagements * s) / steps);
                return next;
              });
            }, 1800 + s * 35 + i * 120),
          );
        }
      });

      timeouts.push(setTimeout(cycle, 6500));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  const maxLeader = Math.max(...LEADERS.map((l) => l.engagements));

  return (
    <DemoShell title="Weekly story · Sat morning">
      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 16 }}>
        {/* Trending hashtags */}
        <div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              fontSize: 11,
              fontWeight: 700,
              letterSpacing: '0.08em',
              color: 'var(--text-tertiary)',
            }}
          >
            <TrendingUp size={12} />
            WHAT&apos;S TRENDING
          </div>
          <div
            style={{
              marginTop: 10,
              display: 'flex',
              flexDirection: 'column',
              gap: 8,
            }}
          >
            {HASHTAGS.map((h, i) => (
              <div
                key={h.tag}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                }}
              >
                <span
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: 4,
                    width: 130,
                    fontSize: 12.5,
                    color: 'var(--text)',
                    fontWeight: 600,
                  }}
                >
                  <Hash size={12} color="var(--text-tertiary)" />
                  {h.tag.replace('#', '')}
                </span>
                <div
                  style={{
                    flex: 1,
                    height: 8,
                    borderRadius: 999,
                    background: 'var(--bg-subtle)',
                    overflow: 'hidden',
                  }}
                >
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{ width: `${pcts[i]}%` }}
                    transition={{ duration: 0.9, ease: 'easeOut' }}
                    style={{
                      height: '100%',
                      background: 'var(--accent)',
                    }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Community leaderboard */}
        <div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              fontSize: 11,
              fontWeight: 700,
              letterSpacing: '0.08em',
              color: 'var(--text-tertiary)',
            }}
          >
            <Trophy size={12} />
            COMMUNITY LEADERBOARD
          </div>
          <div
            style={{
              marginTop: 10,
              display: 'flex',
              flexDirection: 'column',
              gap: 8,
            }}
          >
            {LEADERS.map((leader, i) => (
              <div
                key={leader.name}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                }}
              >
                <span
                  style={{
                    width: 18,
                    fontSize: 12,
                    fontWeight: 700,
                    color: 'var(--text-tertiary)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  #{i + 1}
                </span>
                <span
                  style={{
                    width: 70,
                    fontSize: 13,
                    color: 'var(--text)',
                    fontWeight: 600,
                  }}
                >
                  {leader.name}
                </span>
                <div
                  style={{
                    flex: 1,
                    height: 8,
                    borderRadius: 999,
                    background: 'var(--bg-subtle)',
                    overflow: 'hidden',
                  }}
                >
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{
                      width: `${(leaders[i] / maxLeader) * 100}%`,
                    }}
                    transition={{ duration: 0.4, ease: 'easeOut' }}
                    style={{
                      height: '100%',
                      background: 'var(--green, #16a34a)',
                    }}
                  />
                </div>
                <span
                  style={{
                    width: 46,
                    textAlign: 'right',
                    fontSize: 12,
                    fontVariantNumeric: 'tabular-nums',
                    color: 'var(--text-muted)',
                  }}
                >
                  {leaders[i]}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Honest notes */}
        <div
          style={{
            fontSize: 12,
            fontStyle: 'italic',
            color: 'var(--text-tertiary)',
            lineHeight: 1.55,
            borderTop: '1px solid var(--border)',
            paddingTop: 10,
          }}
        >
          Honest note: chat-module bug skewed engagement on Thursday — treat
          Kevin&apos;s lead with a small grain of salt.
        </div>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  4 — Human-in-the-loop                                              */
/* ------------------------------------------------------------------ */

export function FoxDemoHumanInTheLoop() {
  const [stage, setStage] = useState<0 | 1 | 2>(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setStage(0);
      timeouts.push(setTimeout(() => setStage(1), 1400));
      timeouts.push(setTimeout(() => setStage(2), 3000));
      timeouts.push(setTimeout(cycle, 5800));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  return (
    <DemoShell title="Risky event · queued for human">
      <div style={{ padding: 18, display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div
          style={{
            border: '1px solid var(--border)',
            borderRadius: 12,
            background: 'var(--bg-card)',
            padding: '14px 16px',
          }}
        >
          <div
            style={{
              fontSize: 11,
              fontWeight: 700,
              letterSpacing: '0.08em',
              color: 'var(--text-tertiary)',
            }}
          >
            OFFICIAL · SUBMITTED
          </div>
          <div
            style={{
              marginTop: 6,
              fontWeight: 700,
              fontSize: 14.5,
              color: 'var(--text)',
            }}
          >
            Off-campus club night · 11 PM
          </div>
          <div
            style={{
              marginTop: 8,
              display: 'flex',
              gap: 14,
              fontSize: 12.5,
              color: 'var(--text-muted)',
              flexWrap: 'wrap',
            }}
          >
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
              <Calendar size={13} />
              Sat · 11:00 PM
            </span>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
              <MapPin size={13} />
              Off-campus venue
            </span>
          </div>
        </div>

        <AnimatePresence mode="wait">
          {stage >= 1 && (
            <motion.div
              key="lowconf"
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0 }}
              style={{
                border: '1px solid var(--border)',
                borderRadius: 12,
                padding: '12px 14px',
                background: 'var(--bg-subtle)',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                flexWrap: 'wrap',
              }}
            >
              <FoxBadge />
              <span style={{ fontSize: 13, color: 'var(--text)' }}>
                Confidence too low to auto-approve
              </span>
              <span
                style={{
                  marginLeft: 'auto',
                  fontSize: 12,
                  color: 'var(--text-tertiary)',
                  fontVariantNumeric: 'tabular-nums',
                }}
              >
                <Clock size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                54%
              </span>
            </motion.div>
          )}

          {stage >= 2 && (
            <motion.div
              key="human"
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.35 }}
              style={{
                border: '1px solid var(--accent)',
                borderRadius: 12,
                background: 'var(--bg-card)',
                padding: '12px 14px',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
              }}
            >
              <AlertTriangle size={16} color="var(--accent)" />
              <span style={{ fontSize: 13, color: 'var(--text)', fontWeight: 600 }}>
                Routed to an admin for the final call
              </span>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </DemoShell>
  );
}
