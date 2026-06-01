'use client';

import { useEffect, useState, type ReactNode } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Heart,
  MessageCircle,
  Repeat2,
  Bookmark,
  Send,
  MoreHorizontal,
} from 'lucide-react';

/* ------------------------------------------------------------------ */
/*  Shared primitives — mirror the real Sunibgram components           */
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

/* Avatar silhouette — identical to the real Sunibgram fallback */
function SgAvatar({ size = 40 }: { size?: number }) {
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: '50%',
        overflow: 'hidden',
        flexShrink: 0,
      }}
    >
      <svg viewBox="0 0 100 100" width={size} height={size}>
        <circle cx="50" cy="50" r="50" fill="var(--bg-hover)" />
        <circle cx="50" cy="38" r="18" fill="var(--text-tertiary)" />
        <ellipse cx="50" cy="95" rx="32" ry="26" fill="var(--text-tertiary)" />
      </svg>
    </div>
  );
}

function FacultyChip({ children }: { children: ReactNode }) {
  return (
    <span
      style={{
        fontSize: 11,
        fontWeight: 600,
        padding: '2px 8px',
        borderRadius: 999,
        background: 'var(--accent-soft)',
        color: 'var(--accent)',
        letterSpacing: '0.01em',
        whiteSpace: 'nowrap',
      }}
    >
      {children}
    </span>
  );
}

function ActionBar({
  comments = 0,
  reposts = 0,
  likes = 0,
  bookmarks = 0,
  liked = false,
  reposted = false,
  bookmarked = false,
}: {
  comments?: number;
  reposts?: number;
  likes?: number;
  bookmarks?: number;
  liked?: boolean;
  reposted?: boolean;
  bookmarked?: boolean;
}) {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        marginTop: 10,
        color: 'var(--text-tertiary)',
        maxWidth: 360,
      }}
    >
      <ActionItem icon={<MessageCircle size={18} strokeWidth={2} />} count={comments} />
      <ActionItem
        icon={<Repeat2 size={18} strokeWidth={2} />}
        count={reposts}
        color={reposted ? 'var(--green)' : undefined}
      />
      <motion.span animate={liked ? { scale: [1, 1.3, 1] } : { scale: 1 }} transition={{ duration: 0.35 }}>
        <ActionItem
          icon={<Heart size={18} strokeWidth={2} fill={liked ? 'var(--coral)' : 'none'} />}
          count={likes}
          color={liked ? 'var(--coral)' : undefined}
        />
      </motion.span>
      <ActionItem
        icon={<Bookmark size={18} strokeWidth={2} fill={bookmarked ? 'var(--accent)' : 'none'} />}
        count={bookmarks || undefined}
        color={bookmarked ? 'var(--accent)' : undefined}
      />
      <ActionItem icon={<Send size={18} strokeWidth={2} />} />
    </div>
  );
}

function ActionItem({ icon, count, color }: { icon: ReactNode; count?: number; color?: string }) {
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, color: color ?? 'inherit' }}>
      {icon}
      {typeof count === 'number' && (
        <span style={{ fontSize: 13, fontVariantNumeric: 'tabular-nums' }}>{count}</span>
      )}
    </span>
  );
}

function ThreadRow({
  name,
  handle,
  time,
  faculty,
  body,
  showCaret = false,
  meta,
}: {
  name: string;
  handle: string;
  time: string;
  faculty?: string;
  body: ReactNode;
  showCaret?: boolean;
  meta?: ReactNode;
}) {
  return (
    <article
      style={{
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg-card)',
        padding: '14px 20px',
        display: 'flex',
        gap: 12,
      }}
    >
      <SgAvatar size={40} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4, flexWrap: 'wrap' }}>
          <span style={{ fontWeight: 700, fontSize: 14, color: 'var(--text)', letterSpacing: '-0.01em' }}>
            {name}
          </span>
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>{handle}</span>
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>·</span>
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>{time}</span>
          <span style={{ marginLeft: 'auto' }}>
            <MoreHorizontal size={16} color="var(--text-tertiary)" />
          </span>
        </div>
        {faculty && (
          <div style={{ marginTop: 4 }}>
            <FacultyChip>{faculty}</FacultyChip>
          </div>
        )}
        <div style={{ marginTop: 6, fontSize: 14, lineHeight: 1.45, color: 'var(--text)', whiteSpace: 'pre-wrap' }}>
          {body}
          {showCaret && <span className="sg-caret" />}
        </div>
        {meta}
      </div>
    </article>
  );
}

/* ------------------------------------------------------------------ */
/*  1 — Threads & replies                                              */
/* ------------------------------------------------------------------ */

const THREAD_TEXT = 'ada yg udh selesai project softeng? deadline jam 11 malem 😭';
const REPLY_TEXT = 'blm bgt, masih stuck di sequence diagram. mau ngerjain bareng ga?';

export function ThreadsDemoReplies() {
  const [typed, setTyped] = useState(0);
  const [posted, setPosted] = useState(false);
  const [reply, setReply] = useState(false);

  useEffect(() => {
    let mounted = true;
    let typeId: ReturnType<typeof setInterval> | null = null;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setTyped(0);
      setPosted(false);
      setReply(false);

      timeouts.push(setTimeout(() => {
        typeId = setInterval(() => {
          setTyped((p) => {
            if (p >= THREAD_TEXT.length) {
              if (typeId) clearInterval(typeId);
              return p;
            }
            return p + 1;
          });
        }, 50);
      }, 400));

      timeouts.push(setTimeout(() => setPosted(true), 3200));
      timeouts.push(setTimeout(() => setReply(true), 4600));
      timeouts.push(setTimeout(cycle, 8000));
    }
    cycle();
    return () => {
      mounted = false;
      if (typeId) clearInterval(typeId);
      timeouts.forEach(clearTimeout);
    };
  }, []);

  const typing = typed < THREAD_TEXT.length;

  return (
    <DemoShell title="For you">
      <AnimatePresence mode="wait">
        {!posted ? (
          /* Composer collapsed-row, matches PostComposer placeholder */
          <motion.div
            key="composer"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              padding: '14px 20px',
              borderBottom: '1px solid var(--border)',
              background: 'var(--bg-card)',
            }}
          >
            <SgAvatar size={40} />
            <div style={{ flex: 1, fontSize: 15, color: typing || typed > 0 ? 'var(--text)' : 'var(--text-tertiary)' }}>
              {typed > 0 ? THREAD_TEXT.slice(0, typed) : 'What is happening?!'}
              {typing && <span className="sg-caret" />}
            </div>
            <button
              disabled
              style={{
                background: typed > 0 ? 'var(--accent)' : 'var(--bg-subtle)',
                color: typed > 0 ? '#fff' : 'var(--text-tertiary)',
                border: '1px solid var(--border)',
                borderRadius: 20,
                fontSize: 14,
                fontWeight: 700,
                padding: '7px 18px',
                cursor: 'default',
              }}
            >
              Post
            </button>
          </motion.div>
        ) : (
          <motion.div
            key="posted"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.25 }}
          >
            <ThreadRow
              name="kamu"
              handle="@kamu"
              time="now"
              faculty="School of Computer Science"
              body={THREAD_TEXT}
              meta={<ActionBar comments={reply ? 1 : 0} likes={reply ? 8 : 0} liked={reply} />}
            />
            <AnimatePresence>
              {reply && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3 }}
                >
                  <ThreadRow
                    name="Maya"
                    handle="@maya"
                    time="just now"
                    faculty="School of Computer Science"
                    body={REPLY_TEXT}
                    meta={<ActionBar />}
                  />
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>
        )}
      </AnimatePresence>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  2 — Faculty feed                                                   */
/* ------------------------------------------------------------------ */

const TABS = ['For you', 'Following', 'Faculty'] as const;
type Tab = (typeof TABS)[number];

type FeedItem = {
  name: string;
  handle: string;
  time: string;
  faculty: string;
  body: string;
};

const FEED_FOR_YOU: FeedItem[] = [
  { name: 'Rina',  handle: '@rina',  time: '12m', faculty: 'School of Design', body: 'critique session di anggrek nanti malem, bawa prototype kalian ya' },
  { name: 'Reza',  handle: '@reza',  time: '34m', faculty: 'BINUS Business School', body: 'career fair binus minggu depan, udh siap-siap CV blm?' },
];
const FEED_FACULTY: FeedItem[] = [
  { name: 'Kevin', handle: '@kevin', time: '5m',  faculty: 'School of Computer Science', body: 'ada yg punya catatan compiler kemaren? gw ketinggalan kelas' },
  { name: 'Maya',  handle: '@maya',  time: '21m', faculty: 'School of Computer Science', body: 'plis bantu, stuck di soal DP problem D bnpc 😭' },
];

export function ThreadsDemoFacultyFeed() {
  const [tab, setTab] = useState<Tab>('For you');

  useEffect(() => {
    const id = setInterval(() => {
      setTab((t) => TABS[(TABS.indexOf(t) + 1) % TABS.length]);
    }, 2400);
    return () => clearInterval(id);
  }, []);

  const items = tab === 'Faculty' ? FEED_FACULTY : FEED_FOR_YOU;

  return (
    <DemoShell title="Home">
      <div style={{ display: 'flex', borderBottom: '1px solid var(--border)', background: 'var(--bg-card)' }}>
        {TABS.map((t) => (
          <button
            key={t}
            style={{
              flex: 1,
              padding: '14px 0',
              fontSize: 14,
              fontWeight: tab === t ? 700 : 500,
              color: tab === t ? 'var(--text)' : 'var(--text-tertiary)',
              background: 'transparent',
              border: 'none',
              position: 'relative',
              cursor: 'default',
            }}
          >
            {t}
            {tab === t && (
              <motion.span
                layoutId="feed-underline"
                style={{
                  position: 'absolute',
                  left: '25%',
                  right: '25%',
                  bottom: 0,
                  height: 4,
                  borderRadius: 2,
                  background: 'var(--accent)',
                }}
              />
            )}
          </button>
        ))}
      </div>
      <motion.div layout>
        <AnimatePresence mode="popLayout">
          {items.map((it) => (
            <motion.div
              key={tab + it.handle}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.3 }}
            >
              <ThreadRow
                name={it.name}
                handle={it.handle}
                time={it.time}
                faculty={it.faculty}
                body={it.body}
                meta={<ActionBar comments={2} reposts={1} likes={14} />}
              />
            </motion.div>
          ))}
        </AnimatePresence>
      </motion.div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  3 — Reposts & quotes                                               */
/* ------------------------------------------------------------------ */

export function ThreadsDemoReposts() {
  const [reposts, setReposts] = useState(2);
  const [quoted, setQuoted] = useState(false);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setReposts(2);
      setQuoted(false);
      timeouts.push(setTimeout(() => setReposts(7), 1200));
      timeouts.push(setTimeout(() => { setReposts(8); setQuoted(true); }, 2400));
      timeouts.push(setTimeout(() => setReposts(18), 3800));
      timeouts.push(setTimeout(cycle, 6200));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  return (
    <DemoShell title="Thread">
      <ThreadRow
        name="Kevin"
        handle="@kevin"
        time="2h"
        faculty="School of Computer Science"
        body="honestly kantin anggrek mah underrated, nasi gorengnya the best fight me"
        meta={<ActionBar comments={4} reposts={reposts} reposted={quoted} likes={67} />}
      />
      <AnimatePresence>
        {quoted && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.35 }}
          >
            <ThreadRow
              name="Rina"
              handle="@rina"
              time="1h"
              faculty="School of Design"
              body={
                <>
                  alsut juga enak loh jgn lupakan 😤
                  <div
                    style={{
                      marginTop: 10,
                      border: '1px solid var(--border)',
                      borderRadius: 12,
                      padding: '10px 12px',
                      background: 'var(--bg-subtle)',
                    }}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
                      <SgAvatar size={18} />
                      <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.01em' }}>
                        Kevin
                      </span>
                      <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@kevin</span>
                    </div>
                    <p
                      style={{
                        fontSize: 14,
                        color: 'var(--text)',
                        lineHeight: 1.5,
                        margin: 0,
                      }}
                    >
                      honestly kantin anggrek mah underrated, nasi gorengnya the best fight me
                    </p>
                  </div>
                </>
              }
              meta={<ActionBar comments={1} reposts={1} likes={9} />}
            />
          </motion.div>
        )}
      </AnimatePresence>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  4 — Polls                                                          */
/* ------------------------------------------------------------------ */

const POLL_OPTIONS = ['Anggrek', 'Alam Sutera', 'Bekasi', 'Malang'];
const POLL_TARGETS = [44, 27, 18, 11];

export function ThreadsDemoPolls() {
  const [pcts, setPcts] = useState<number[]>([0, 0, 0, 0]);
  const [votes, setVotes] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setPcts([0, 0, 0, 0]);
      setVotes(0);
      timeouts.push(setTimeout(() => { setPcts(POLL_TARGETS); setVotes(124); }, 700));
      timeouts.push(setTimeout(() => setVotes(238), 2000));
      timeouts.push(setTimeout(() => setVotes(361), 3300));
      timeouts.push(setTimeout(cycle, 5800));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  const maxPct = Math.max(...pcts);

  return (
    <DemoShell title="Thread">
      <ThreadRow
        name="kamu"
        handle="@kamu"
        time="3h"
        faculty="BINUS Business School"
        body={
          <>
            kantin terbaik di binus mana sih? 🤔
            <div style={{ marginTop: 10, display: 'flex', flexDirection: 'column', gap: 8 }}>
              {POLL_OPTIONS.map((label, i) => {
                const leading = pcts[i] > 0 && pcts[i] === maxPct;
                return (
                  <div
                    key={label}
                    style={{
                      position: 'relative',
                      border: '1px solid var(--border)',
                      borderRadius: 12,
                      padding: '10px 14px',
                      overflow: 'hidden',
                      fontSize: 14,
                      color: 'var(--text)',
                      background: 'var(--bg-card)',
                    }}
                  >
                    <motion.div
                      initial={{ width: 0 }}
                      animate={{ width: `${pcts[i]}%` }}
                      transition={{ duration: 0.9, ease: 'easeOut' }}
                      style={{
                        position: 'absolute',
                        inset: 0,
                        background: leading ? 'var(--accent-soft)' : 'var(--bg-subtle)',
                      }}
                    />
                    <div
                      style={{
                        position: 'relative',
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <span style={{ fontWeight: leading ? 700 : 500 }}>{label}</span>
                      <motion.span
                        key={pcts[i]}
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        style={{
                          color: leading ? 'var(--accent)' : 'var(--text-tertiary)',
                          fontVariantNumeric: 'tabular-nums',
                          fontSize: 13,
                          fontWeight: 700,
                        }}
                      >
                        {pcts[i]}%
                      </motion.span>
                    </div>
                  </div>
                );
              })}
            </div>
            <div style={{ marginTop: 8, fontSize: 12, color: 'var(--text-tertiary)' }}>
              {votes.toLocaleString()} votes · 18h left
            </div>
          </>
        }
        meta={<ActionBar comments={12} reposts={4} likes={47} />}
      />
    </DemoShell>
  );
}
