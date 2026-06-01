'use client';

import { useEffect, useState, type ReactNode } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { Check, CheckCheck, Paperclip, FileText, ImageIcon, Send } from 'lucide-react';

/* ------------------------------------------------------------------ */
/*  Shared primitives — mirror real MessageBubble                      */
/* ------------------------------------------------------------------ */

function DemoShell({ title, subtitle, children }: { title: string; subtitle?: string; children: ReactNode }) {
  return (
    <div className="sg-demo-shell">
      <div className="sg-demo-shell-header" style={{ flexDirection: 'column', alignItems: 'flex-start', gap: 2 }}>
        <span>{title}</span>
        {subtitle && (
          <span style={{ fontSize: 12, color: 'var(--text-tertiary)', fontWeight: 500 }}>{subtitle}</span>
        )}
      </div>
      <div
        className="sg-demo-shell-body"
        style={{
          padding: '14px 12px',
          gap: 4,
          background: 'var(--bg)',
          minHeight: 240,
        }}
      >
        {children}
      </div>
    </div>
  );
}

function SgAvatar({ size = 28 }: { size?: number }) {
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

type BubbleProps = {
  own?: boolean;
  showAvatar?: boolean;
  children: ReactNode;
  bottomMeta?: ReactNode;
};

function Bubble({ own = false, showAvatar = false, children, bottomMeta }: BubbleProps) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'flex-end',
        justifyContent: own ? 'flex-end' : 'flex-start',
        gap: 8,
        padding: '0 8px',
      }}
    >
      {!own && (
        <div style={{ width: 28, display: 'flex', justifyContent: 'center' }}>
          {showAvatar ? <SgAvatar size={28} /> : null}
        </div>
      )}
      <div style={{ display: 'flex', flexDirection: 'column', maxWidth: '68%', alignItems: own ? 'flex-end' : 'flex-start' }}>
        <div
          style={{
            padding: '9px 13px',
            borderRadius: own ? '18px 18px 4px 18px' : '18px 18px 18px 4px',
            background: own ? 'var(--accent)' : 'var(--bg-subtle)',
            color: own ? '#fff' : 'var(--text)',
            fontSize: 14,
            lineHeight: 1.45,
            wordBreak: 'break-word',
            whiteSpace: 'pre-wrap',
          }}
        >
          {children}
        </div>
        {bottomMeta && <div style={{ marginTop: 4 }}>{bottomMeta}</div>}
      </div>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  1 — One-on-one DMs                                                 */
/* ------------------------------------------------------------------ */

type DM = { own: boolean; text: string };
const DM_SEQUENCE: DM[] = [
  { own: false, text: 'lo udh selesai softeng project?' },
  { own: true,  text: 'blm bgt 😭 lo?' },
  { own: false, text: 'baru setengah. mau ngerjain bareng?' },
  { own: true,  text: 'gas, di anggrek perpus jam 9 ya' },
];

export function MessagesDemoDMs() {
  const [shown, setShown] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setShown(0);
      DM_SEQUENCE.forEach((_, i) =>
        timeouts.push(setTimeout(() => setShown(i + 1), 700 + i * 1100))
      );
      timeouts.push(setTimeout(cycle, 700 + DM_SEQUENCE.length * 1100 + 1600));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  return (
    <DemoShell title="Rina" subtitle="@rina · online">
      <AnimatePresence>
        {DM_SEQUENCE.slice(0, shown).map((m, i) => {
          const next = DM_SEQUENCE[i + 1];
          const isLastInGroup = !next || next.own !== m.own;
          return (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.25 }}
            >
              <Bubble own={m.own} showAvatar={!m.own && isLastInGroup}>
                {m.text}
              </Bubble>
            </motion.div>
          );
        })}
      </AnimatePresence>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  2 — Group rooms                                                    */
/* ------------------------------------------------------------------ */

const TYPERS = ['Kevin', 'Maya', 'Reza'];

export function MessagesDemoGroups() {
  const [typingIdx, setTypingIdx] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setTypingIdx((i) => (i + 1) % TYPERS.length), 1800);
    return () => clearInterval(id);
  }, []);

  return (
    <DemoShell title="BNPC Final 2026" subtitle="4 members · group">
      <Bubble own={false} showAvatar>
        <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--accent)', display: 'block', marginBottom: 2 }}>
          Kevin
        </span>
        udh gw push solution problem D di branch dev, cek dulu ya
      </Bubble>
      <Bubble own={true}>
        oke gw review setelah kelas, jam 4an
      </Bubble>
      <Bubble own={false} showAvatar>
        <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--coral)', display: 'block', marginBottom: 2 }}>
          Maya
        </span>
        gw nemu edge case di problem E, mau di-pair sama siapa?
      </Bubble>

      <div style={{ display: 'flex', gap: 8, padding: '0 8px', alignItems: 'center' }}>
        <div style={{ width: 28 }}>
          <SgAvatar size={28} />
        </div>
        <motion.div
          key={typingIdx}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          style={{
            padding: '8px 13px',
            borderRadius: '18px 18px 18px 4px',
            background: 'var(--bg-subtle)',
            display: 'inline-flex',
            alignItems: 'center',
            gap: 6,
          }}
        >
          <span style={{ fontSize: 12, color: 'var(--text-tertiary)', fontWeight: 600 }}>
            {TYPERS[typingIdx]} is typing
          </span>
          <span style={{ display: 'inline-flex', gap: 3 }}>
            {[0, 1, 2].map((j) => (
              <motion.span
                key={j}
                animate={{ opacity: [0.25, 1, 0.25] }}
                transition={{ duration: 1, repeat: Infinity, delay: j * 0.18 }}
                style={{
                  width: 4,
                  height: 4,
                  borderRadius: 999,
                  background: 'var(--text-tertiary)',
                }}
              />
            ))}
          </span>
        </motion.div>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  3 — Media & files                                                  */
/* ------------------------------------------------------------------ */

const ATTACHMENTS = [
  { icon: FileText,  label: 'softeng-uc-diagram.pdf', size: '2.4 MB' },
  { icon: ImageIcon, label: 'whiteboard.jpg',         size: '1.1 MB' },
  { icon: FileText,  label: 'requirements-v3.pdf',    size: '780 KB' },
];

export function MessagesDemoMedia() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setCount(0);
      timeouts.push(setTimeout(() => setCount(1), 700));
      timeouts.push(setTimeout(() => setCount(2), 1600));
      timeouts.push(setTimeout(() => setCount(3), 2500));
      timeouts.push(setTimeout(cycle, 4800));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  return (
    <DemoShell title="Maya" subtitle="@maya">
      <Bubble own>
        ini bahan softeng yg gw kerjain semalem 👇
      </Bubble>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6, alignItems: 'flex-end', padding: '0 8px' }}>
        <AnimatePresence>
          {ATTACHMENTS.slice(0, count).map((a, i) => {
            const Icon = a.icon;
            return (
              <motion.div
                key={i}
                initial={{ opacity: 0, x: 12 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.25 }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  padding: '8px 12px',
                  borderRadius: '14px 14px 4px 14px',
                  background: 'var(--accent)',
                  color: '#fff',
                  maxWidth: 260,
                }}
              >
                <div
                  style={{
                    width: 32,
                    height: 32,
                    borderRadius: 8,
                    background: 'rgba(255,255,255,0.15)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <Icon size={16} />
                </div>
                <div style={{ minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 700, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {a.label}
                  </div>
                  <div style={{ fontSize: 11, opacity: 0.8 }}>{a.size}</div>
                </div>
                <Paperclip size={12} style={{ opacity: 0.7 }} />
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  4 — Read receipts                                                  */
/* ------------------------------------------------------------------ */

type ReadState = 'sending' | 'sent' | 'delivered' | 'seen';

export function MessagesDemoReadReceipts() {
  const [state, setState] = useState<ReadState>('sending');

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setState('sending');
      timeouts.push(setTimeout(() => setState('sent'), 700));
      timeouts.push(setTimeout(() => setState('delivered'), 1500));
      timeouts.push(setTimeout(() => setState('seen'), 2500));
      timeouts.push(setTimeout(cycle, 4400));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  const label =
    state === 'sending'   ? 'Sending…' :
    state === 'sent'      ? 'Sent'     :
    state === 'delivered' ? 'Delivered' : 'Seen · 21:42';
  const tickColor = state === 'seen' ? 'var(--accent)' : 'var(--text-tertiary)';

  return (
    <DemoShell title="Kevin" subtitle="@kevin">
      <Bubble own>see you at the library, jam 9 ya</Bubble>
      <Bubble own>jgn telat 😤</Bubble>
      <div style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center', gap: 6, padding: '0 8px', marginTop: 2 }}>
        <motion.span
          key={state}
          initial={{ opacity: 0, scale: 0.85 }}
          animate={{ opacity: 1, scale: 1 }}
          style={{ color: tickColor, display: 'inline-flex' }}
        >
          {state === 'sending' ? (
            <span className="sg-spinner" />
          ) : state === 'sent' ? (
            <Check size={13} />
          ) : (
            <CheckCheck size={13} />
          )}
        </motion.span>
        <motion.span key={label} initial={{ opacity: 0 }} animate={{ opacity: 1 }} style={{ fontSize: 11, color: 'var(--text-muted)' }}>
          {label}
        </motion.span>
      </div>
    </DemoShell>
  );
}
