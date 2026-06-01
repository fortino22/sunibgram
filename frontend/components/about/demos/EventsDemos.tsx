'use client';

import { useEffect, useState, type ReactNode } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import {
  CalendarDays,
  MapPin,
  Users,
  Heart,
  MessageCircle,
  MoreHorizontal,
  Send,
  Repeat2,
  Bookmark,
} from 'lucide-react';

/* ------------------------------------------------------------------ */
/*  Shared primitives — mirror real EventCard                          */
/* ------------------------------------------------------------------ */

function DemoShell({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="sg-demo-shell">
      <div className="sg-demo-shell-header">
        <span>{title}</span>
        <span style={{ fontSize: 12, color: 'var(--text-tertiary)', fontWeight: 500 }}>
          Sunibgram preview
        </span>
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

function OfficialBadge({ official }: { official: boolean }) {
  return (
    <span
      style={{
        fontSize: 11,
        fontWeight: 600,
        color: official ? 'var(--accent)' : 'var(--text-tertiary)',
        background: official ? 'var(--accent-soft)' : 'transparent',
        border: `1px solid ${official ? 'var(--accent-soft)' : 'var(--border)'}`,
        borderRadius: 999,
        padding: '1px 8px',
        whiteSpace: 'nowrap',
      }}
    >
      {official ? 'Official' : 'Unofficial'}
    </span>
  );
}

function EventRow({
  title,
  date,
  location,
  official,
  live = false,
  thumbTint = 'color-mix(in oklab, var(--accent) 14%, var(--bg-subtle))',
}: {
  title: string;
  date: string;
  location: string;
  official: boolean;
  live?: boolean;
  thumbTint?: string;
}) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '14px 20px',
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg-card)',
      }}
    >
      <div
        style={{
          width: 56,
          height: 56,
          borderRadius: 10,
          background: thumbTint,
          border: '1px solid var(--border)',
          flexShrink: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <CalendarDays size={22} color="var(--text-tertiary)" />
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4, flexWrap: 'wrap' }}>
          <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.005em' }}>
            {title}
          </span>
          <OfficialBadge official={official} />
          {live && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                fontSize: 11,
                fontWeight: 700,
                color: 'var(--green)',
              }}
            >
              <span
                style={{
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  background: 'var(--green)',
                  animation: 'sg-pulse 1.6s ease-in-out infinite',
                }}
              />
              Live
            </span>
          )}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, color: 'var(--text-tertiary)' }}>
          <CalendarDays size={13} />
          {date}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, color: 'var(--text-tertiary)', marginTop: 2 }}>
          <MapPin size={13} />
          {location}
        </div>
      </div>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  1 — Official events (approval flow)                                */
/* ------------------------------------------------------------------ */

type Status = 'pending' | 'revision' | 'approved';

export function EventsDemoOfficial() {
  const [status, setStatus] = useState<Status>('pending');

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setStatus('pending');
      timeouts.push(setTimeout(() => setStatus('revision'), 1700));
      timeouts.push(setTimeout(() => setStatus('approved'), 3400));
      timeouts.push(setTimeout(cycle, 6000));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  const statusTone =
    status === 'approved' ? { color: 'var(--green)',  label: 'Approved · Official' } :
    status === 'revision' ? { color: 'var(--amber)',  label: 'Revision requested' } :
    { color: 'var(--text-tertiary)', label: 'Pending admin review' };

  return (
    <DemoShell title="Submit an event">
      <EventRow
        title="BNPC Final 2026"
        date="Sat, Jun 14 · 09:00 – 17:00"
        location="Binus Anggrek"
        official={status === 'approved'}
      />
      <div style={{ padding: '14px 20px', display: 'flex', flexDirection: 'column', gap: 10, background: 'var(--bg-card)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--text-tertiary)', letterSpacing: '0.05em' }}>
            STATUS
          </span>
          <motion.span
            key={status}
            initial={{ opacity: 0, x: -6 }}
            animate={{ opacity: 1, x: 0 }}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '3px 10px',
              borderRadius: 999,
              border: `1px solid ${statusTone.color}`,
              color: statusTone.color,
              fontSize: 12,
              fontWeight: 700,
            }}
          >
            <span style={{ width: 6, height: 6, borderRadius: 999, background: statusTone.color }} />
            {statusTone.label}
          </motion.span>
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-muted)', lineHeight: 1.5 }}>
          Submissions go through faculty review. Once approved, the event lands on the campus feed with a verified Official badge.
        </div>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  2 — Student-run meetups (form fill + publish)                       */
/* ------------------------------------------------------------------ */

const FIELDS = [
  { label: 'Title',    value: 'Sunmori Binus Riders' },
  { label: 'When',     value: 'Sun, Jun 8 · 06:00' },
  { label: 'Where',    value: 'Binus Bekasi' },
  { label: 'Audience', value: 'All Binus students' },
];

export function EventsDemoStudentMeetups() {
  const [filled, setFilled] = useState(0);
  const [published, setPublished] = useState(false);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setFilled(0);
      setPublished(false);
      FIELDS.forEach((_, i) =>
        timeouts.push(setTimeout(() => setFilled(i + 1), 600 + i * 700))
      );
      timeouts.push(setTimeout(() => setPublished(true), 4000));
      timeouts.push(setTimeout(cycle, 6400));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  return (
    <DemoShell title="Create a meetup">
      <div style={{ padding: '14px 20px', display: 'flex', flexDirection: 'column', gap: 12, background: 'var(--bg-card)' }}>
        {FIELDS.map((f, i) => (
          <div key={f.label} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--text-tertiary)', letterSpacing: '0.05em' }}>
              {f.label.toUpperCase()}
            </span>
            <div
              style={{
                border: '1px solid var(--border)',
                borderRadius: 10,
                padding: '8px 12px',
                fontSize: 14,
                color: i < filled ? 'var(--text)' : 'var(--text-tertiary)',
                background: 'var(--bg-card)',
                minHeight: 36,
                display: 'flex',
                alignItems: 'center',
              }}
            >
              {i < filled ? (
                <motion.span initial={{ opacity: 0 }} animate={{ opacity: 1 }}>{f.value}</motion.span>
              ) : (
                <span style={{ fontStyle: 'italic' }}>…</span>
              )}
            </div>
          </div>
        ))}
        <motion.button
          animate={{
            background: published ? 'var(--green)' : 'var(--accent)',
            color: '#fff',
          }}
          style={{
            marginTop: 6,
            padding: '10px 14px',
            borderRadius: 999,
            border: 'none',
            fontSize: 14,
            fontWeight: 700,
            cursor: 'default',
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 8,
          }}
        >
          <Send size={14} />
          {published ? 'Published' : 'Publish meetup'}
        </motion.button>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  3 — Event threads                                                  */
/* ------------------------------------------------------------------ */

type EvThread = { name: string; handle: string; time: string; tag: 'Official' | 'Discussion'; body: string };

const EV_THREADS: EvThread[] = [
  {
    name: 'BNPC Official',
    handle: '@bnpc',
    time: '4h',
    tag: 'Official',
    body: 'doors open jam 08:30. plis bawa KTM & laptop yg udh di-set environmentnya 🙏',
  },
  {
    name: 'Maya',
    handle: '@maya',
    time: '1h',
    tag: 'Discussion',
    body: 'ada yg carpool dari alam sutera? bareng yuk, patungan bensin aja',
  },
];

export function EventsDemoEventThreads() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setCount(0);
      timeouts.push(setTimeout(() => setCount(1), 900));
      timeouts.push(setTimeout(() => setCount(2), 2200));
      timeouts.push(setTimeout(cycle, 4800));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  return (
    <DemoShell title="BNPC Final 2026 · Threads">
      <EventRow
        title="BNPC Final 2026"
        date="Sat, Jun 14"
        location="Binus Anggrek"
        official
        live
      />
      <AnimatePresence>
        {EV_THREADS.slice(0, count).map((t, i) => (
          <motion.article
            key={i}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
            style={{
              display: 'flex',
              gap: 12,
              padding: '14px 20px',
              borderBottom: '1px solid var(--border)',
              background: 'var(--bg-card)',
            }}
          >
            <SgAvatar size={40} />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
                <span style={{ fontSize: 14, fontWeight: 700, color: 'var(--text)' }}>{t.name}</span>
                <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>{t.handle}</span>
                <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>·</span>
                <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>{t.time}</span>
                <OfficialBadge official={t.tag === 'Official'} />
                <span style={{ marginLeft: 'auto' }}>
                  <MoreHorizontal size={16} color="var(--text-tertiary)" />
                </span>
              </div>
              <div style={{ marginTop: 6, fontSize: 14, lineHeight: 1.45, color: 'var(--text)' }}>{t.body}</div>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  marginTop: 10,
                  color: 'var(--text-tertiary)',
                  maxWidth: 320,
                }}
              >
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 13 }}>
                  <MessageCircle size={16} /> 3
                </span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 13 }}>
                  <Repeat2 size={16} /> 2
                </span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 13 }}>
                  <Heart size={16} /> {t.tag === 'Official' ? 41 : 12}
                </span>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 13 }}>
                  <Bookmark size={16} />
                </span>
              </div>
            </div>
          </motion.article>
        ))}
      </AnimatePresence>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  4 — RSVPs                                                          */
/* ------------------------------------------------------------------ */

export function EventsDemoRSVPs() {
  const [count, setCount] = useState(0);
  const [going, setGoing] = useState(false);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setCount(0);
      setGoing(false);
      timeouts.push(setTimeout(() => setCount(38), 700));
      timeouts.push(setTimeout(() => setCount(74), 1700));
      timeouts.push(setTimeout(() => { setGoing(true); setCount(75); }, 2900));
      timeouts.push(setTimeout(() => setCount(112), 3900));
      timeouts.push(setTimeout(cycle, 5800));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  const avatarCount = Math.min(5, Math.ceil(count / 25));

  return (
    <DemoShell title="Event · RSVP">
      <EventRow
        title="Pekan Olahraga Anggrek"
        date="Fri, Jun 13 · 16:00"
        location="Binus Anggrek"
        official={false}
      />
      <div
        style={{
          padding: '14px 20px',
          background: 'var(--bg-card)',
          display: 'flex',
          flexDirection: 'column',
          gap: 12,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <Users size={16} color="var(--text-tertiary)" />
          <motion.span
            key={count}
            initial={{ y: -4, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)', fontVariantNumeric: 'tabular-nums' }}
          >
            {count}
          </motion.span>
          <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>going</span>
          <div style={{ flex: 1 }} />
          <div style={{ display: 'flex' }}>
            {Array.from({ length: avatarCount }).map((_, i) => (
              <motion.div
                key={i}
                initial={{ scale: 0, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ delay: i * 0.05 }}
                style={{ marginLeft: i === 0 ? 0 : -10, border: '2px solid var(--bg-card)', borderRadius: '50%' }}
              >
                <SgAvatar size={26} />
              </motion.div>
            ))}
          </div>
        </div>
        <motion.button
          animate={{
            background: going ? 'var(--accent)' : 'var(--bg-card)',
            color: going ? '#fff' : 'var(--text)',
            borderColor: going ? 'var(--accent)' : 'var(--border)',
          }}
          style={{
            padding: '10px 14px',
            borderRadius: 999,
            border: '1px solid',
            fontSize: 14,
            fontWeight: 700,
            cursor: 'default',
          }}
        >
          {going ? "You're going" : 'RSVP'}
        </motion.button>
      </div>
    </DemoShell>
  );
}
