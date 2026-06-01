'use client';

import { useEffect, useState, type ReactNode } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Lock,
  Globe,
  Shield,
  UserCheck,
  Search,
  MessageCircle,
  Repeat2,
  Heart,
  Bookmark,
  Send,
  MoreHorizontal,
} from 'lucide-react';

/* ------------------------------------------------------------------ */
/*  Shared primitives — mirror the real CommunityListCard / Member row */
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

function CommunityThumb({ tint = 'var(--bg-subtle)' }: { tint?: string }) {
  return (
    <div
      style={{
        width: 56,
        height: 56,
        borderRadius: 10,
        background: tint,
        border: '1px solid var(--border)',
        flexShrink: 0,
      }}
    />
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

function CommunityRow({
  name,
  members,
  isPrivate = false,
  trailing,
  tint,
}: {
  name: string;
  members: number;
  isPrivate?: boolean;
  trailing?: ReactNode;
  tint?: string;
}) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        width: '100%',
        padding: '14px 20px',
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg-card)',
      }}
    >
      <CommunityThumb tint={tint} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.005em' }}>
            {name}
          </span>
          {isPrivate && <Lock size={12} color="var(--text-tertiary)" />}
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-tertiary)', marginTop: 2 }}>
          {members.toLocaleString()} members
        </div>
      </div>
      {trailing}
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  1 — Public & private rooms                                         */
/* ------------------------------------------------------------------ */

export function CommunitiesDemoVisibility() {
  const [joinedPublic, setJoinedPublic] = useState(false);
  const [requestedPrivate, setRequestedPrivate] = useState(false);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setJoinedPublic(false);
      setRequestedPrivate(false);
      timeouts.push(setTimeout(() => setJoinedPublic(true), 1100));
      timeouts.push(setTimeout(() => setRequestedPrivate(true), 2400));
      timeouts.push(setTimeout(cycle, 4800));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  return (
    <DemoShell title="Discover communities">
      <CommunityRow
        name="BNPC Coders"
        members={2412}
        tint="color-mix(in oklab, var(--accent) 18%, var(--bg-subtle))"
        trailing={
          <motion.button
            animate={{
              background: joinedPublic ? 'var(--accent-soft)' : 'var(--accent)',
              color: joinedPublic ? 'var(--accent)' : '#fff',
              borderColor: 'transparent',
            }}
            style={{
              padding: '7px 16px',
              borderRadius: 999,
              border: '1px solid',
              fontSize: 13,
              fontWeight: 700,
              cursor: 'default',
              minWidth: 84,
            }}
          >
            {joinedPublic ? (
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                <Globe size={12} /> Joined
              </span>
            ) : (
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                <Globe size={12} /> Join
              </span>
            )}
          </motion.button>
        }
      />
      <CommunityRow
        name="Tugas Akhir Survival"
        members={84}
        isPrivate
        tint="color-mix(in oklab, var(--coral) 15%, var(--bg-subtle))"
        trailing={
          <motion.button
            animate={{
              background: requestedPrivate ? 'var(--bg-subtle)' : 'var(--bg-card)',
              color: requestedPrivate ? 'var(--text-tertiary)' : 'var(--text)',
            }}
            style={{
              padding: '7px 14px',
              borderRadius: 999,
              border: '1px solid var(--border)',
              fontSize: 13,
              fontWeight: 700,
              cursor: 'default',
              minWidth: 84,
            }}
          >
            {requestedPrivate ? 'Requested' : 'Request'}
          </motion.button>
        }
      />
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  2 — Moderation: role change                                        */
/* ------------------------------------------------------------------ */

type Role = 'MEMBER' | 'ADMIN';
const ROLE_COLOR: Record<Role, string> = {
  MEMBER: 'var(--text-tertiary)',
  ADMIN: 'var(--accent)',
};

const MEMBERS: { username: string; role: Role }[] = [
  { username: 'rina',  role: 'ADMIN' },
  { username: 'kevin', role: 'MEMBER' },
  { username: 'maya',  role: 'MEMBER' },
];

export function CommunitiesDemoModeration() {
  const [promoted, setPromoted] = useState(false);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];
    function cycle() {
      if (!mounted) return;
      setPromoted(false);
      timeouts.push(setTimeout(() => setPromoted(true), 1600));
      timeouts.push(setTimeout(cycle, 4400));
    }
    cycle();
    return () => { mounted = false; timeouts.forEach(clearTimeout); };
  }, []);

  return (
    <DemoShell title="BNPC Coders · Members">
      {MEMBERS.map((m, i) => {
        const role: Role = i === 2 && promoted ? 'ADMIN' : m.role;
        return (
          <div
            key={m.username}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              padding: '12px 20px',
              borderBottom: '1px solid var(--border)',
              background: 'var(--bg-card)',
            }}
          >
            <SgAvatar size={40} />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)' }}>@{m.username}</div>
              <motion.div
                key={role}
                initial={{ opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: ROLE_COLOR[role],
                  marginTop: 2,
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 4,
                }}
              >
                {role === 'ADMIN' ? <Shield size={11} /> : <UserCheck size={11} />}
                {role}
              </motion.div>
            </div>
            <button
              style={{
                width: 28,
                height: 28,
                borderRadius: '50%',
                border: '1px solid var(--border)',
                background: 'transparent',
                color: 'var(--text-tertiary)',
                fontSize: 16,
                lineHeight: 1,
                cursor: 'default',
              }}
            >
              ⋯
            </button>
          </div>
        );
      })}
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  3 — Category discovery                                             */
/* ------------------------------------------------------------------ */

const CATS = ['All', 'Academic', 'Sports', 'Tech', 'Arts'];
type CatItem = { name: string; cat: string; members: number; tint: string };
const COMMUNITIES: CatItem[] = [
  { name: 'BNPC Coders',           cat: 'Tech',     members: 2412, tint: 'color-mix(in oklab, var(--accent) 18%, var(--bg-subtle))' },
  { name: 'ML Study Group',        cat: 'Tech',     members: 894,  tint: 'color-mix(in oklab, var(--accent) 14%, var(--bg-subtle))' },
  { name: 'Alam Sutera Futsal',    cat: 'Sports',   members: 612,  tint: 'color-mix(in oklab, var(--green) 18%, var(--bg-subtle))' },
  { name: 'Anggrek Run Club',      cat: 'Sports',   members: 478,  tint: 'color-mix(in oklab, var(--green) 14%, var(--bg-subtle))' },
  { name: 'Design Critique',       cat: 'Arts',     members: 314,  tint: 'color-mix(in oklab, var(--coral) 18%, var(--bg-subtle))' },
  { name: 'Tugas Akhir Survival',  cat: 'Academic', members: 1207, tint: 'color-mix(in oklab, var(--amber) 18%, var(--bg-subtle))' },
];

export function CommunitiesDemoDiscovery() {
  const [idx, setIdx] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setIdx((i) => (i + 1) % CATS.length), 2400);
    return () => clearInterval(id);
  }, []);

  const active = CATS[idx];
  const visible = active === 'All' ? COMMUNITIES.slice(0, 3) : COMMUNITIES.filter((c) => c.cat === active);

  return (
    <DemoShell title="Discover">
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          padding: '12px 20px',
          borderBottom: '1px solid var(--border)',
          background: 'var(--bg-card)',
        }}
      >
        <Search size={16} color="var(--text-tertiary)" />
        <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>Search communities…</span>
      </div>
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
        {CATS.map((c) => (
          <motion.span
            key={c}
            animate={{
              background: c === active ? 'var(--text)' : 'var(--bg-subtle)',
              color: c === active ? 'var(--bg)' : 'var(--text)',
            }}
            style={{
              padding: '6px 14px',
              borderRadius: 999,
              fontSize: 13,
              fontWeight: 600,
              whiteSpace: 'nowrap',
            }}
          >
            {c}
          </motion.span>
        ))}
      </div>
      <motion.div layout>
        <AnimatePresence mode="popLayout">
          {visible.map((c) => (
            <motion.div
              layout
              key={c.name}
              initial={{ opacity: 0, x: 16 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -16 }}
              transition={{ duration: 0.3 }}
            >
              <CommunityRow name={c.name} members={c.members} tint={c.tint} />
            </motion.div>
          ))}
        </AnimatePresence>
      </motion.div>
    </DemoShell>
  );
}

/* ------------------------------------------------------------------ */
/*  4 — Member-only threads                                            */
/* ------------------------------------------------------------------ */

export function CommunitiesDemoMemberThreads() {
  const [joined, setJoined] = useState(false);

  useEffect(() => {
    const id = setInterval(() => setJoined((j) => !j), 3000);
    return () => clearInterval(id);
  }, []);

  return (
    <DemoShell title="Tugas Akhir Survival">
      <CommunityRow
        name="Tugas Akhir Survival"
        members={84}
        isPrivate
        tint="color-mix(in oklab, var(--amber) 18%, var(--bg-subtle))"
        trailing={
          <motion.button
            animate={{
              background: joined ? 'var(--accent-soft)' : 'var(--bg-card)',
              color: joined ? 'var(--accent)' : 'var(--text)',
            }}
            style={{
              padding: '7px 14px',
              borderRadius: 999,
              border: '1px solid var(--border)',
              fontSize: 13,
              fontWeight: 700,
              cursor: 'default',
              minWidth: 84,
            }}
          >
            {joined ? 'Member' : 'Request'}
          </motion.button>
        }
      />

      <div style={{ position: 'relative' }}>
        <div
          style={{
            background: 'var(--bg-card)',
            padding: '14px 20px',
            display: 'flex',
            gap: 12,
            borderBottom: '1px solid var(--border)',
          }}
        >
          <SgAvatar size={40} />
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 4, flexWrap: 'wrap' }}>
              <span style={{ fontWeight: 700, fontSize: 14, color: 'var(--text)', letterSpacing: '-0.01em' }}>
                Kevin
              </span>
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@kevin</span>
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>·</span>
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>2h</span>
              <span style={{ marginLeft: 'auto' }}>
                <MoreHorizontal size={16} color="var(--text-tertiary)" />
              </span>
            </div>
            <motion.div
              animate={{ filter: joined ? 'blur(0px)' : 'blur(6px)' }}
              transition={{ duration: 0.4 }}
              style={{
                marginTop: 6,
                fontSize: 14,
                lineHeight: 1.45,
                color: 'var(--text)',
                userSelect: joined ? 'auto' : 'none',
              }}
            >
              ada yg udh sidang minggu ini? share tipsnya plis 🙏 gw masih revisi bab 3 nih
            </motion.div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                marginTop: 10,
                color: 'var(--text-tertiary)',
                maxWidth: 360,
              }}
            >
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                <MessageCircle size={18} strokeWidth={2} />
                <span style={{ fontSize: 13, fontVariantNumeric: 'tabular-nums' }}>7</span>
              </span>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                <Repeat2 size={18} strokeWidth={2} />
                <span style={{ fontSize: 13, fontVariantNumeric: 'tabular-nums' }}>2</span>
              </span>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                <Heart size={18} strokeWidth={2} />
                <span style={{ fontSize: 13, fontVariantNumeric: 'tabular-nums' }}>24</span>
              </span>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                <Bookmark size={18} strokeWidth={2} />
              </span>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                <Send size={18} strokeWidth={2} />
              </span>
            </div>
          </div>
        </div>

        <AnimatePresence>
          {!joined && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              style={{
                position: 'absolute',
                inset: 0,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 8,
                background: 'color-mix(in oklab, var(--bg-card) 80%, transparent)',
              }}
            >
              <Lock size={22} color="var(--text-tertiary)" />
              <span style={{ fontSize: 13, color: 'var(--text-muted)', fontWeight: 500 }}>
                Join to read this thread
              </span>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </DemoShell>
  );
}
