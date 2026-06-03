'use client';

import { useEffect, useState, type ReactNode } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Repeat2,
  Bookmark,
  MessageCircle,
  Share2,
  Heart,
  Clock,
  RefreshCw,
  ArrowUp,
  Flame,
  Layers,
} from 'lucide-react';

/* ------------------------------------------------------------------ */
/*  Shared demo shell                                                  */
/* ------------------------------------------------------------------ */

function DemoShell({
  title,
  subtitle = 'For You · live demo',
  children,
}: {
  title: string;
  subtitle?: string;
  children: ReactNode;
}) {
  return (
    <div className="sg-demo-shell">
      <div className="sg-demo-shell-header">
        <span>{title}</span>
        <span style={{ fontSize: 12, color: 'var(--text-tertiary)', fontWeight: 500 }}>
          {subtitle}
        </span>
      </div>
      <div className="sg-demo-shell-body">{children}</div>
    </div>
  );
}

/* A tiny post card used across several demos */
function MiniPost({
  initial,
  name,
  text,
  tint,
}: {
  initial: string;
  name: string;
  text: string;
  tint?: string;
}) {
  return (
    <div style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
      <div
        style={{
          width: 28,
          height: 28,
          borderRadius: 999,
          flexShrink: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: 12,
          fontWeight: 700,
          color: 'var(--bg)',
          background: tint ?? 'var(--text-tertiary)',
        }}
      >
        {initial}
      </div>
      <div style={{ minWidth: 0 }}>
        <div style={{ fontSize: 12.5, fontWeight: 700, color: 'var(--text)' }}>{name}</div>
        <div style={{ fontSize: 12.5, color: 'var(--text-muted)', lineHeight: 1.5 }}>{text}</div>
      </div>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  1 — What shapes the interest score                                 */
/* ------------------------------------------------------------------ */

const SCORE_FACTORS = [
  { label: 'Reactions', detail: 'likes · comments · reposts', target: 82, icon: <Heart size={13} /> },
  { label: 'Freshness', detail: 'how recently it was posted', target: 68, icon: <Clock size={13} /> },
  { label: 'Source', detail: 'people, communities, trending', target: 74, icon: <Layers size={13} /> },
];

export function AlgoDemoInterestScore() {
  const [pcts, setPcts] = useState<number[]>([0, 0, 0]);
  const [score, setScore] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setPcts([0, 0, 0]);
      setScore(0);

      SCORE_FACTORS.forEach((f, i) => {
        timeouts.push(
          setTimeout(() => {
            setPcts((prev) => {
              const next = [...prev];
              next[i] = f.target;
              return next;
            });
          }, 500 + i * 260),
        );
      });

      const steps = 22;
      for (let s = 1; s <= steps; s++) {
        timeouts.push(
          setTimeout(() => setScore(Math.round((86 * s) / steps)), 1400 + s * 30),
        );
      }

      timeouts.push(setTimeout(cycle, 6500));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  return (
    <DemoShell title="Interest score">
      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div
          style={{
            border: '1px solid var(--border)',
            borderRadius: 12,
            background: 'var(--bg-card)',
            padding: '12px 14px',
          }}
        >
          <MiniPost
            initial="R"
            name="Rina"
            text="Found a quiet study spot on the 5th floor of Anggrek 👀"
            tint="var(--accent)"
          />
        </div>

        {/* Factor bars */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {SCORE_FACTORS.map((f, i) => (
            <div key={f.label} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 6,
                  width: 96,
                  fontSize: 12.5,
                  fontWeight: 600,
                  color: 'var(--text)',
                }}
              >
                <span style={{ color: 'var(--text-tertiary)' }}>{f.icon}</span>
                {f.label}
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
                  transition={{ duration: 0.8, ease: 'easeOut' }}
                  style={{ height: '100%', background: 'var(--accent)' }}
                />
              </div>
            </div>
          ))}
        </div>

        {/* Composite score */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            borderTop: '1px solid var(--border)',
            paddingTop: 12,
          }}
        >
          <span style={{ fontSize: 12.5, color: 'var(--text-muted)' }}>
            Blended into one quiet score
          </span>
          <span
            style={{
              marginLeft: 'auto',
              fontSize: 24,
              fontWeight: 700,
              color: 'var(--text)',
              fontVariantNumeric: 'tabular-nums',
              letterSpacing: '-0.02em',
            }}
          >
            {score}
          </span>
        </div>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  2 — Not all reactions count the same                               */
/* ------------------------------------------------------------------ */

const REACTIONS = [
  { label: 'Repost', note: 'Counts the most', width: 100, icon: <Repeat2 size={14} /> },
  { label: 'Bookmark', note: 'Counts a lot', width: 84, icon: <Bookmark size={14} /> },
  { label: 'Comment', note: 'Counts a lot', width: 80, icon: <MessageCircle size={14} /> },
  { label: 'Share', note: 'A bit more than a like', width: 62, icon: <Share2 size={14} /> },
  { label: 'Like', note: 'The baseline', width: 42, icon: <Heart size={14} /> },
];

export function AlgoDemoReactionWeights() {
  const [widths, setWidths] = useState<number[]>(REACTIONS.map(() => 0));

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setWidths(REACTIONS.map(() => 0));
      REACTIONS.forEach((r, i) => {
        timeouts.push(
          setTimeout(() => {
            setWidths((prev) => {
              const next = [...prev];
              next[i] = r.width;
              return next;
            });
          }, 400 + i * 200),
        );
      });
      timeouts.push(setTimeout(cycle, 6500));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  return (
    <DemoShell title="How much a reaction says">
      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
        {REACTIONS.map((r, i) => (
          <div key={r.label} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 6,
                width: 96,
                fontSize: 12.5,
                fontWeight: 700,
                color: 'var(--text)',
              }}
            >
              <span style={{ color: 'var(--text-tertiary)' }}>{r.icon}</span>
              {r.label}
            </span>
            <div
              style={{
                flex: 1,
                height: 10,
                borderRadius: 999,
                background: 'var(--bg-subtle)',
                overflow: 'hidden',
              }}
            >
              <motion.div
                initial={{ width: 0 }}
                animate={{ width: `${widths[i]}%` }}
                transition={{ duration: 0.7, ease: 'easeOut' }}
                style={{
                  height: '100%',
                  background: i === 0 ? 'var(--accent)' : 'var(--text-tertiary)',
                }}
              />
            </div>
            <span
              style={{
                width: 116,
                textAlign: 'right',
                fontSize: 11.5,
                color: 'var(--text-tertiary)',
              }}
            >
              {r.note}
            </span>
          </div>
        ))}

        <p
          style={{
            marginTop: 2,
            fontSize: 12,
            fontStyle: 'italic',
            color: 'var(--text-tertiary)',
            lineHeight: 1.55,
            borderTop: '1px solid var(--border)',
            paddingTop: 10,
          }}
        >
          Replies show up too, but are nudged down a little — the feed leans
          toward standalone posts over back-and-forth replies.
        </p>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  3 — Newer posts rise, older ones fade                              */
/* ------------------------------------------------------------------ */

const DECAY_STOPS = [
  { label: 'Just posted', strength: 100 },
  { label: '+1 hour', strength: 78 },
  { label: '+3 hours', strength: 52 },
  { label: '+6 hours', strength: 31 },
  { label: '+1 day', strength: 14 },
];

export function AlgoDemoTimeDecay() {
  const [active, setActive] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      DECAY_STOPS.forEach((_, i) => {
        timeouts.push(setTimeout(() => setActive(i), 700 + i * 900));
      });
      timeouts.push(setTimeout(cycle, 700 + DECAY_STOPS.length * 900 + 1400));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  const current = DECAY_STOPS[active];

  return (
    <DemoShell title="A post's strength over time">
      <div style={{ padding: 18, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ display: 'flex', alignItems: 'flex-end', gap: 8, height: 120 }}>
          {DECAY_STOPS.map((stop, i) => (
            <div
              key={stop.label}
              style={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: 8,
                height: '100%',
                justifyContent: 'flex-end',
              }}
            >
              <div
                style={{
                  width: '100%',
                  flex: 1,
                  display: 'flex',
                  alignItems: 'flex-end',
                }}
              >
                <motion.div
                  animate={{
                    height: `${stop.strength}%`,
                    opacity: i <= active ? 1 : 0.25,
                  }}
                  transition={{ duration: 0.6, ease: 'easeOut' }}
                  style={{
                    width: '100%',
                    borderRadius: '6px 6px 0 0',
                    background:
                      i === active ? 'var(--accent)' : 'var(--text-tertiary)',
                  }}
                />
              </div>
            </div>
          ))}
        </div>

        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            borderTop: '1px solid var(--border)',
            paddingTop: 12,
          }}
        >
          <Clock size={15} color="var(--text-tertiary)" />
          <span style={{ fontSize: 13, color: 'var(--text)', fontWeight: 600 }}>
            {current.label}
          </span>
          <span
            style={{
              marginLeft: 'auto',
              fontSize: 13,
              color: 'var(--text-muted)',
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            strength {current.strength}%
          </span>
        </div>

        <p style={{ fontSize: 12, color: 'var(--text-tertiary)', lineHeight: 1.55 }}>
          Posts older than about a week leave For You entirely — the feed keeps
          itself recent.
        </p>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  4 — Fresh from the last few minutes                                */
/* ------------------------------------------------------------------ */

const RANKED_POSTS = [
  { initial: 'K', name: 'Kevin', text: 'BNPC results just dropped 🏆', tint: 'var(--text-tertiary)' },
  { initial: 'M', name: 'Maya', text: 'Anyone selling a calculus textbook?', tint: 'var(--text-tertiary)' },
];

const FRESH_DROPS = [
  { initial: 'A', name: 'Andra', text: 'Free pizza at the Kemanggisan lobby right now!' },
  { initial: 'L', name: 'Lia', text: 'Career fair booths are open — line is short' },
];

export function AlgoDemoFreshSlots() {
  const [freshIdx, setFreshIdx] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setFreshIdx(0);
      timeouts.push(setTimeout(() => setFreshIdx(1), 2600));
      timeouts.push(setTimeout(cycle, 6000));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  const fresh = FRESH_DROPS[freshIdx];

  return (
    <DemoShell title="Reserved spots for brand-new posts">
      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
        {/* Fresh slot */}
        <div
          style={{
            border: '1px solid var(--accent)',
            borderRadius: 12,
            background: 'var(--accent-soft)',
            padding: '12px 14px',
          }}
        >
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 5,
              fontSize: 10.5,
              fontWeight: 700,
              letterSpacing: '0.08em',
              color: 'var(--accent)',
              marginBottom: 8,
            }}
          >
            <Flame size={12} />
            FRESH · JUST NOW
          </div>
          <AnimatePresence mode="wait">
            <motion.div
              key={freshIdx}
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: 8 }}
              transition={{ duration: 0.35 }}
            >
              <MiniPost
                initial={fresh.initial}
                name={fresh.name}
                text={fresh.text}
                tint="var(--accent)"
              />
            </motion.div>
          </AnimatePresence>
        </div>

        <div
          style={{
            fontSize: 10.5,
            fontWeight: 700,
            letterSpacing: '0.08em',
            color: 'var(--text-tertiary)',
            padding: '2px 2px',
          }}
        >
          RANKED BY INTEREST
        </div>

        {RANKED_POSTS.map((p) => (
          <div
            key={p.name}
            style={{
              border: '1px solid var(--border)',
              borderRadius: 12,
              background: 'var(--bg-card)',
              padding: '12px 14px',
            }}
          >
            <MiniPost initial={p.initial} name={p.name} text={p.text} tint={p.tint} />
          </div>
        ))}

        <p style={{ fontSize: 12, color: 'var(--text-tertiary)', lineHeight: 1.55 }}>
          Only a couple of fresh posts sit up top at a time, newest first — so a
          just-dropped post is never buried before it can collect reactions.
        </p>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  5 — A feed that feels new each visit                               */
/* ------------------------------------------------------------------ */

const SHUFFLE_BASE = [
  { id: 'a', initial: 'R', name: 'Rina', tint: 'var(--accent)' },
  { id: 'b', initial: 'K', name: 'Kevin', tint: 'var(--text-tertiary)' },
  { id: 'c', initial: 'M', name: 'Maya', tint: 'var(--green, #16a34a)' },
  { id: 'd', initial: 'A', name: 'Andra', tint: 'var(--text-tertiary)' },
];

const SHUFFLE_ORDERS = [
  ['a', 'b', 'c', 'd'],
  ['b', 'a', 'd', 'c'],
  ['a', 'c', 'b', 'd'],
];

export function AlgoDemoReshuffle() {
  const [orderIdx, setOrderIdx] = useState(0);
  const [spinning, setSpinning] = useState(false);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setSpinning(true);
      timeouts.push(setTimeout(() => setSpinning(false), 700));
      timeouts.push(
        setTimeout(() => setOrderIdx((p) => (p + 1) % SHUFFLE_ORDERS.length), 350),
      );
      timeouts.push(setTimeout(cycle, 3200));
    }
    timeouts.push(setTimeout(cycle, 1600));
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  const order = SHUFFLE_ORDERS[orderIdx];
  const items = order.map((id) => SHUFFLE_BASE.find((b) => b.id === id)!);

  return (
    <DemoShell title="Refresh reshuffles similar posts">
      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>
            All roughly equal in interest
          </span>
          <span
            style={{
              marginLeft: 'auto',
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              fontSize: 12,
              fontWeight: 700,
              color: 'var(--accent)',
              padding: '5px 12px',
              borderRadius: 999,
              border: '1px solid var(--accent)',
              background: 'var(--accent-soft)',
            }}
          >
            <motion.span
              animate={{ rotate: spinning ? 360 : 0 }}
              transition={{ duration: 0.7, ease: 'easeInOut' }}
              style={{ display: 'inline-flex' }}
            >
              <RefreshCw size={13} />
            </motion.span>
            Refresh
          </span>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {items.map((item) => (
            <motion.div
              key={item.id}
              layout
              transition={{ type: 'spring', stiffness: 320, damping: 30 }}
              style={{
                border: '1px solid var(--border)',
                borderRadius: 12,
                background: 'var(--bg-card)',
                padding: '10px 14px',
              }}
            >
              <MiniPost
                initial={item.initial}
                name={item.name}
                text="Same interest score — order varies each visit."
                tint={item.tint}
              />
            </motion.div>
          ))}
        </div>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  6 — Fair to everyone (per-author cap)                              */
/* ------------------------------------------------------------------ */

export function AlgoDemoAuthorCap() {
  const [reveal, setReveal] = useState(false);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setReveal(false);
      timeouts.push(setTimeout(() => setReveal(true), 1600));
      timeouts.push(setTimeout(cycle, 5200));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  return (
    <DemoShell title="No single account takes over">
      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
        {[0, 1].map((n) => (
          <div
            key={n}
            style={{
              border: '1px solid var(--border)',
              borderRadius: 12,
              background: 'var(--bg-card)',
              padding: '12px 14px',
            }}
          >
            <MiniPost
              initial="M"
              name="Maya"
              text={
                n === 0
                  ? 'Posting all day about the career fair (1/7)'
                  : 'Booth C has the best swag (2/7)'
              }
              tint="var(--accent)"
            />
          </div>
        ))}

        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            padding: '8px 12px',
            borderRadius: 10,
            border: '1px dashed var(--border)',
            fontSize: 12,
            color: 'var(--text-tertiary)',
          }}
        >
          <Layers size={13} />
          Maya&apos;s other 5 posts are held back so the feed stays varied
        </div>

        <AnimatePresence>
          {reveal &&
            [
              { initial: 'K', name: 'Kevin', text: 'Quiet floor in Anggrek is unreal today' },
              { initial: 'L', name: 'Lia', text: 'Anyone forming a study group for finals?' },
            ].map((p) => (
              <motion.div
                key={p.name}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3 }}
                style={{
                  border: '1px solid var(--border)',
                  borderRadius: 12,
                  background: 'var(--bg-card)',
                  padding: '12px 14px',
                }}
              >
                <MiniPost
                  initial={p.initial}
                  name={p.name}
                  text={p.text}
                  tint="var(--text-tertiary)"
                />
              </motion.div>
            ))}
        </AnimatePresence>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  7 — The "New posts" banner                                         */
/* ------------------------------------------------------------------ */

export function AlgoDemoNewPostsBanner() {
  const [stage, setStage] = useState<0 | 1 | 2>(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setStage(0);
      timeouts.push(setTimeout(() => setStage(1), 1600));
      timeouts.push(setTimeout(() => setStage(2), 3400));
      timeouts.push(setTimeout(cycle, 6400));
    }
    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  return (
    <DemoShell title="Quietly checking for new posts">
      <div style={{ padding: 16, position: 'relative', minHeight: 220 }}>
        {/* Banner */}
        <div style={{ height: 44, marginBottom: 10, display: 'flex', justifyContent: 'center' }}>
          <AnimatePresence>
            {stage >= 1 && (
              <motion.button
                key="banner"
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                transition={{ duration: 0.3 }}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 6,
                  padding: '8px 16px',
                  borderRadius: 999,
                  background: 'var(--text)',
                  color: 'var(--bg)',
                  fontSize: 12.5,
                  fontWeight: 700,
                  border: 'none',
                  cursor: 'pointer',
                }}
              >
                <ArrowUp size={13} />
                4 new posts
              </motion.button>
            )}
          </AnimatePresence>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <AnimatePresence>
            {stage >= 2 &&
              [
                { initial: 'A', name: 'Andra', text: 'Just slid in on top — your spot stays put' },
                { initial: 'L', name: 'Lia', text: 'Nothing below gets reshuffled' },
              ].map((p) => (
                <motion.div
                  key={p.name}
                  initial={{ opacity: 0, height: 0, y: -12 }}
                  animate={{ opacity: 1, height: 'auto', y: 0 }}
                  transition={{ duration: 0.35 }}
                  style={{
                    border: '1px solid var(--accent)',
                    borderRadius: 12,
                    background: 'var(--accent-soft)',
                    padding: '12px 14px',
                    overflow: 'hidden',
                  }}
                >
                  <MiniPost
                    initial={p.initial}
                    name={p.name}
                    text={p.text}
                    tint="var(--accent)"
                  />
                </motion.div>
              ))}
          </AnimatePresence>

          {[
            { initial: 'R', name: 'Rina', text: "Where you were scrolling — doesn't move" },
            { initial: 'K', name: 'Kevin', text: 'Your current read stays anchored here' },
          ].map((p) => (
            <div
              key={p.name}
              style={{
                border: '1px solid var(--border)',
                borderRadius: 12,
                background: 'var(--bg-card)',
                padding: '12px 14px',
              }}
            >
              <MiniPost initial={p.initial} name={p.name} text={p.text} tint="var(--text-tertiary)" />
            </div>
          ))}
        </div>
      </div>
    </DemoShell>
  );
}
