'use client';

import { useEffect, useRef, useState, type ReactNode } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { ArrowUp, ArrowDown, Minus, Trophy, Flame } from 'lucide-react';

/* ------------------------------------------------------------------ */
/*  Shared primitives — mirror real LeaderboardRow                     */
/* ------------------------------------------------------------------ */

function DemoShell({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="sg-demo-shell">
      <div className="sg-demo-shell-header">
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
          <Trophy size={14} color="var(--amber)" />
          {title}
        </span>
        <span style={{ fontSize: 12, color: 'var(--text-tertiary)', fontWeight: 500 }}>This week</span>
      </div>
      <div className="sg-demo-shell-body">{children}</div>
    </div>
  );
}

function SgAvatar({ size = 40 }: { size?: number }) {
  return (
    <div style={{ width: size, height: size, borderRadius: '50%', overflow: 'hidden', flexShrink: 0 }}>
      <svg viewBox="0 0 100 100" width={size} height={size}>
        <circle cx="50" cy="50" r="50" fill="var(--bg-hover)" />
        <circle cx="50" cy="38" r="18" fill="var(--text-tertiary)" />
        <ellipse cx="50" cy="95" rx="32" ry="26" fill="var(--text-tertiary)" />
      </svg>
    </div>
  );
}

type Leader = { name: string; username: string; faculty: string; score: number };

function LeaderRow({ leader, rank, delta }: { leader: Leader; rank: number; delta: 'up' | 'down' | 'same' }) {
  const Icon = delta === 'up' ? ArrowUp : delta === 'down' ? ArrowDown : Minus;
  const iconColor =
    delta === 'up'   ? 'var(--green)'  :
    delta === 'down' ? 'var(--coral)'  :
    'var(--text-tertiary)';
  const rankColor = rank <= 3 ? 'var(--text)' : 'var(--text-tertiary)';

  return (
    <motion.div
      layout
      transition={{ type: 'spring', damping: 26, stiffness: 220 }}
      style={{
        display: 'grid',
        gridTemplateColumns: '40px 1fr auto auto',
        alignItems: 'center',
        gap: 12,
        padding: '14px 16px',
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg-card)',
      }}
    >
      <span style={{ fontWeight: 700, fontSize: 14, color: rankColor, textAlign: 'center', fontVariantNumeric: 'tabular-nums' }}>
        {rank}
      </span>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, minWidth: 0 }}>
        <SgAvatar size={40} />
        <div style={{ minWidth: 0 }}>
          <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
            {leader.name}
          </div>
          <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>
            @{leader.username} · {leader.faculty}
          </div>
        </div>
      </div>
      <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)', fontVariantNumeric: 'tabular-nums', padding: '0 4px' }}>
        {leader.score.toFixed(1)}
      </span>
      <motion.span
        key={delta + rank}
        initial={{ opacity: 0, y: delta === 'up' ? 4 : -4 }}
        animate={{ opacity: 1, y: 0 }}
        style={{ color: iconColor, display: 'inline-flex' }}
      >
        <Icon size={14} />
      </motion.span>
    </motion.div>
  );
}

/* ------------------------------------------------------------------ */
/*  1 — Weekly rankings (re-sort on tick)                              */
/* ------------------------------------------------------------------ */

const INITIAL: Leader[] = [
  { name: 'Rina',  username: 'rina',  faculty: 'School of Design',            score: 92.3 },
  { name: 'Kevin', username: 'kevin', faculty: 'School of Computer Science',  score: 88.7 },
  { name: 'Maya',  username: 'maya',  faculty: 'School of Information Systems', score: 84.1 },
  { name: 'Reza',  username: 'reza',  faculty: 'BINUS Business School',       score: 79.5 },
  { name: 'Putri', username: 'putri', faculty: 'School of Design',            score: 76.8 },
];

export function LeaderboardDemoWeekly() {
  const [leaders, setLeaders] = useState<Leader[]>(INITIAL);
  const prevRanks = useRef<Record<string, number>>(
    Object.fromEntries(INITIAL.map((l, i) => [l.username, i + 1]))
  );

  useEffect(() => {
    const id = setInterval(() => {
      setLeaders((current) => {
        const next = current.map((l) => ({
          ...l,
          score: Math.max(0, +(l.score + (Math.random() * 4 - 1.4)).toFixed(1)),
        }));
        next.sort((a, b) => b.score - a.score);
        return next;
      });
    }, 2400);
    return () => clearInterval(id);
  }, []);

  return (
    <DemoShell title="Campus leaderboard">
      <AnimatePresence>
        {leaders.map((l, i) => {
          const rank = i + 1;
          const prev = prevRanks.current[l.username] ?? rank;
          const delta: 'up' | 'down' | 'same' =
            rank < prev ? 'up' : rank > prev ? 'down' : 'same';
          prevRanks.current[l.username] = rank;
          return <LeaderRow key={l.username} leader={l} rank={rank} delta={delta} />;
        })}
      </AnimatePresence>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  2 — Engagement scores (your stats card)                            */
/* ------------------------------------------------------------------ */

const METRICS = [
  { label: 'Replies received', target: 48, color: 'var(--accent)' },
  { label: 'Reposts earned',   target: 32, color: 'var(--green)' },
  { label: 'Likes from peers', target: 88, color: 'var(--coral)' },
];

export function LeaderboardDemoScores() {
  const [values, setValues] = useState<number[]>([0, 0, 0]);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setValues([0, 0, 0]);
      timeouts.push(setTimeout(() => setValues(METRICS.map((m) => m.target)), 600));
      timeouts.push(setTimeout(cycle, 4400));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  const total = (values[0] * 1.5 + values[1] * 2 + values[2] * 0.6).toFixed(1);

  return (
    <DemoShell title="Your engagement">
      <div style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 14, background: 'var(--bg-card)' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--text-tertiary)', letterSpacing: '0.05em' }}>
            THIS WEEK
          </span>
          <motion.span
            key={total}
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '4px 12px',
              borderRadius: 999,
              background: 'var(--accent-soft)',
              color: 'var(--accent)',
              fontSize: 13,
              fontWeight: 700,
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            <Flame size={13} />
            Score {total}
          </motion.span>
        </div>
        {METRICS.map((m, i) => (
          <div key={m.label} style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>{m.label}</span>
              <motion.span
                key={values[i]}
                initial={{ y: -4, opacity: 0 }}
                animate={{ y: 0, opacity: 1 }}
                style={{ fontSize: 13, fontWeight: 700, color: 'var(--text)', fontVariantNumeric: 'tabular-nums' }}
              >
                {values[i]}
              </motion.span>
            </div>
            <div
              style={{
                height: 6,
                borderRadius: 999,
                background: 'var(--bg-subtle)',
                overflow: 'hidden',
              }}
            >
              <motion.div
                initial={{ width: 0 }}
                animate={{ width: `${(values[i] / 100) * 100}%` }}
                transition={{ duration: 1.1, ease: 'easeOut' }}
                style={{ height: '100%', background: m.color }}
              />
            </div>
          </div>
        ))}
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  3 — Faculty leaderboards                                           */
/* ------------------------------------------------------------------ */

const FACULTIES = ['Design', 'Computer Science', 'Business'];
const FACULTY_TOP: Record<string, Leader[]> = {
  Design: [
    { name: 'Rina',  username: 'rina',  faculty: 'School of Design', score: 92.3 },
    { name: 'Putri', username: 'putri', faculty: 'School of Design', score: 88.1 },
    { name: 'Arif',  username: 'arif',  faculty: 'School of Design', score: 81.6 },
  ],
  'Computer Science': [
    { name: 'Kevin', username: 'kevin', faculty: 'School of Computer Science', score: 90.4 },
    { name: 'Maya',  username: 'maya',  faculty: 'School of Computer Science', score: 86.2 },
    { name: 'Reza',  username: 'reza',  faculty: 'School of Computer Science', score: 82.1 },
  ],
  Business: [
    { name: 'Nadya', username: 'nadya', faculty: 'BINUS Business School', score: 87.7 },
    { name: 'Bagas', username: 'bagas', faculty: 'BINUS Business School', score: 83.4 },
    { name: 'Sari',  username: 'sari',  faculty: 'BINUS Business School', score: 79.0 },
  ],
};

export function LeaderboardDemoFaculty() {
  const [idx, setIdx] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setIdx((i) => (i + 1) % FACULTIES.length), 2600);
    return () => clearInterval(id);
  }, []);

  const active = FACULTIES[idx];
  const top = FACULTY_TOP[active];

  return (
    <DemoShell title={`Top in ${active}`}>
      <div
        style={{
          display: 'flex',
          gap: 8,
          padding: '12px 20px',
          borderBottom: '1px solid var(--border)',
          background: 'var(--bg-card)',
          overflowX: 'auto',
        }}
      >
        {FACULTIES.map((f) => (
          <motion.span
            key={f}
            animate={{
              background: f === active ? 'var(--text)' : 'var(--bg-subtle)',
              color: f === active ? 'var(--bg)' : 'var(--text)',
            }}
            style={{
              padding: '6px 14px',
              borderRadius: 999,
              fontSize: 13,
              fontWeight: 600,
              whiteSpace: 'nowrap',
            }}
          >
            {f}
          </motion.span>
        ))}
      </div>
      <AnimatePresence mode="wait">
        <motion.div
          key={active}
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -8 }}
          transition={{ duration: 0.3 }}
        >
          {top.map((l, i) => (
            <LeaderRow key={l.username} leader={l} rank={i + 1} delta="same" />
          ))}
        </motion.div>
      </AnimatePresence>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  4 — Your stats (week-over-week trend)                              */
/* ------------------------------------------------------------------ */

const TREND_POINTS = [22, 31, 28, 44, 52, 67, 78];
const TREND_LABELS = ['W1', 'W2', 'W3', 'W4', 'W5', 'W6', 'W7'];

export function LeaderboardDemoYourStats() {
  const [drawn, setDrawn] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setDrawn(0);
      timeouts.push(setTimeout(() => setDrawn(1), 500));
      timeouts.push(setTimeout(cycle, 4400));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  /* Chart geometry — mirrors recharts margin: { top: 12, right: 16, left: 0, bottom: 4 } */
  const W = 320;
  const H = 160;
  const PAD_LEFT = 30;     /* room for Y-axis tick labels */
  const PAD_RIGHT = 12;
  const PAD_TOP = 12;
  const PAD_BOTTOM = 24;   /* room for X-axis tick labels */
  const innerW = W - PAD_LEFT - PAD_RIGHT;
  const innerH = H - PAD_TOP - PAD_BOTTOM;

  const yMax = 100;
  const yTicks = [0, 25, 50, 75, 100];

  const px = (i: number) => PAD_LEFT + (i / (TREND_POINTS.length - 1)) * innerW;
  const py = (v: number) => PAD_TOP + (1 - v / yMax) * innerH;

  const linePath = TREND_POINTS.map((v, i) => `${px(i)},${py(v)}`).join(' ');

  return (
    <DemoShell title="Your weekly score">
      <div style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 14, background: 'var(--bg-card)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            <span style={{ fontSize: 24, fontWeight: 700, color: 'var(--text)', fontVariantNumeric: 'tabular-nums', letterSpacing: '-0.02em' }}>
              78
            </span>
            <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>this week</span>
          </div>
          <motion.span
            initial={{ opacity: 0 }}
            animate={{ opacity: drawn ? 1 : 0 }}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '4px 10px',
              borderRadius: 999,
              background: 'color-mix(in oklab, var(--green) 18%, transparent)',
              color: 'var(--green)',
              fontSize: 12,
              fontWeight: 700,
            }}
          >
            <ArrowUp size={12} /> +11 vs W6
          </motion.span>
        </div>
        <svg viewBox={`0 0 ${W} ${H}`} style={{ width: '100%', height: 'auto', display: 'block' }}>
          {/* CartesianGrid — dashed horizontal lines, mirrors recharts strokeDasharray="3 3" */}
          {yTicks.map((t) => (
            <line
              key={`grid-${t}`}
              x1={PAD_LEFT}
              x2={W - PAD_RIGHT}
              y1={py(t)}
              y2={py(t)}
              stroke="var(--border)"
              strokeWidth="1"
              strokeDasharray="3 3"
            />
          ))}

          {/* Y-axis tick labels */}
          {yTicks.map((t) => (
            <text
              key={`ytick-${t}`}
              x={PAD_LEFT - 6}
              y={py(t)}
              fontSize="11"
              fill="var(--text-tertiary)"
              textAnchor="end"
              dominantBaseline="middle"
            >
              {t}
            </text>
          ))}

          {/* X-axis tick labels */}
          {TREND_LABELS.map((l, i) => (
            <text
              key={`xtick-${l}`}
              x={px(i)}
              y={H - 6}
              fontSize="11"
              fill="var(--text-tertiary)"
              textAnchor="middle"
            >
              {l}
            </text>
          ))}

          {/* Line — var(--text), width 2 (matches real YourStats) */}
          <motion.polyline
            points={linePath}
            fill="none"
            stroke="var(--text)"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            initial={{ pathLength: 0 }}
            animate={{ pathLength: drawn ? 1 : 0 }}
            transition={{ duration: 1.1, ease: 'easeOut' }}
          />

          {/* Dots — var(--text) filled, r 3.5 (matches real YourStats) */}
          {TREND_POINTS.map((v, i) => (
            <motion.circle
              key={`dot-${i}`}
              cx={px(i)}
              cy={py(v)}
              r="3.5"
              fill="var(--text)"
              initial={{ opacity: 0 }}
              animate={{ opacity: drawn ? 1 : 0 }}
              transition={{ delay: drawn ? 0.7 + i * 0.05 : 0 }}
            />
          ))}
        </svg>
      </div>
    </DemoShell>
  );
}
