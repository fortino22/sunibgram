'use client';

import { useState } from 'react';
import { ProfilePageHeader } from '@/components/profile/ProfilePageHeader';
import { ProfileBanner } from '@/components/profile/ProfileBanner';
import { ProfileInfo } from '@/components/profile/ProfileInfo';
import { ProfileTabs, type ProfileTab } from '@/components/profile/ProfileTabs';
import { ProfileThreadCard } from '@/components/profile/ProfileThreadCard';
import { EmptyState } from '@/components/shared/EmptyState';
import type { ThreadItem } from '@/lib/services/threads';

// ─── Mock Data ────────────────────────────────────────────────────────────────

const MOCK_USER = {
  name: 'Aria Putri',
  username: 'ariaputri',
  bio: 'Computer Science student at Binus 🎓 | Tech enthusiast, coffee addict, and occasional night owl. Building things that matter.',
  faculty: 'School of Computer Science',
  joinedMonth: 'March 2024',
  following: 142,
  followers: 1847,
  avatarUrl: null,
  bannerUrl: null,
};

function thread(id: number, content: string, overrides: Partial<ThreadItem> = {}): ThreadItem {
  return {
    type: 'POST', id, parent_id: null, root_id: null, content,
    created_at: '2026-04-01T10:00:00Z', updated_at: '2026-04-01T10:00:00Z', time_ago: '1h',
    author: { id: 1, username: 'ariaputri', is_premium: false },
    stats: { likes: 24, reposts: 6, bookmarks: 3 },
    media: [], is_liked: false, is_bookmarked: false, is_reposted: false, is_pin: false,
    ...overrides,
  };
}

const MOCK_THREADS: ThreadItem[] = [
  thread(1, 'Just finished my final project on distributed systems and honestly it feels surreal. Four months of suffering compressed into 47 slides 😭 #BinusLife', { time_ago: '2h', stats: { likes: 84, reposts: 12, bookmarks: 31 }, is_liked: true, is_pin: true }),
  thread(2, 'Hot take: the library at 11 pm hits different. Something about the quiet desperation of everyone around you that really fuels productivity.', { time_ago: '1d', stats: { likes: 247, reposts: 38, bookmarks: 15 }, is_reposted: true }),
  thread(3, "Starting a study group for Algorithms next week. DM me if you're in SOCS and struggling with dynamic programming — we're all struggling together.", { time_ago: '3d', stats: { likes: 56, reposts: 22, bookmarks: 8 } }),
  thread(4, 'Random thought: why does every group project have one person who disappears until the day before the deadline and then suddenly has "great ideas"?', { time_ago: '5d', stats: { likes: 412, reposts: 67, bookmarks: 29 }, is_liked: true }),
];

const MOCK_REPLIES: (ThreadItem & { replyingTo: string })[] = [
  { ...thread(5, 'Agree 100%! The vending machine by the computer lab is a lifesaver at 2 am 😂', { time_ago: '3h', stats: { likes: 18, reposts: 2, bookmarks: 1 } }), replyingTo: 'binusofficialpage' },
  { ...thread(6, 'This is exactly why I switched to Obsidian for notes. Game changer for organizing lecture slides.', { time_ago: '2d', stats: { likes: 29, reposts: 5, bookmarks: 11 } }), replyingTo: 'notesnerdbinus' },
  { ...thread(7, "Honestly the grading rubric was unclear from the start but I get it, the TA's are overwhelmed too.", { time_ago: '4d', stats: { likes: 41, reposts: 3, bookmarks: 2 } }), replyingTo: 'socs_helpdesk' },
];

const MOCK_LIKES: ThreadItem[] = [
  thread(8, "Reminder that the hackathon registration closes this Friday! Don't sleep on it 🔥 #BinusHackathon2026", { time_ago: '5h', stats: { likes: 198, reposts: 44, bookmarks: 27 }, is_liked: true, author: { id: 9, username: 'binusteach', is_premium: true } }),
  thread(9, 'The Binus cafeteria actually upgraded their menu and nobody is talking about this?? The new pasta section goes crazy', { time_ago: '1d', stats: { likes: 331, reposts: 78, bookmarks: 12 }, is_liked: true, author: { id: 2, username: 'campusfoodie', is_premium: false } }),
  thread(10, 'PSA: you can book private study rooms in the library app now. No more fighting for spots 🙌', { time_ago: '3d', stats: { likes: 612, reposts: 156, bookmarks: 89 }, is_liked: true, author: { id: 3, username: 'libraryupdate', is_premium: false } }),
];

<<<<<<< HEAD
// ─── Page ─────────────────────────────────────────────────────────────────────
=======
// ─── Tab Config ───────────────────────────────────────────────────────────────

type ProfileTab = 'threads' | 'replies' | 'media' | 'likes';

const PROFILE_TABS: { id: ProfileTab; label: string }[] = [
  { id: 'threads', label: 'Threads' },
  { id: 'replies', label: 'Replies' },
  { id: 'media', label: 'Media' },
  { id: 'likes', label: 'Likes' },
];

// ─── Thread Card ──────────────────────────────────────────────────────────────

function ProfileThreadCard({ thread }: { thread: ThreadItem }) {
  const initial = thread.author.username.charAt(0).toUpperCase();

  return (
    <article
      style={{
        borderBottom: '1px solid var(--border)',
        padding: '16px 20px',
        display: 'flex',
        gap: 12,
        background: 'var(--bg-card)',
        cursor: 'pointer',
        transition: 'background 0.1s',
      }}
      onMouseEnter={(e: MouseEvent<HTMLElement>) => ((e.currentTarget as HTMLElement).style.background = 'var(--bg-subtle)')}
      onMouseLeave={(e: MouseEvent<HTMLElement>) => ((e.currentTarget as HTMLElement).style.background = 'var(--bg-card)')}
    >
      {/* Avatar */}
      <div
        style={{
          width: 40,
          height: 40,
          borderRadius: '50%',
          background: 'oklch(58% 0.23 270)',
          color: '#fff',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
          fontWeight: 600,
          fontSize: 15,
          letterSpacing: '-0.01em',
        }}
      >
        {initial}
      </div>

      <div style={{ flex: 1, minWidth: 0 }}>
        {/* Author row */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
          <span style={{ fontWeight: 600, fontSize: 14, color: 'var(--text)', letterSpacing: '-0.01em' }}>
            {thread.author.username}
          </span>
          {thread.author.is_premium && (
            <span
              style={{
                fontSize: 10,
                background: 'var(--accent-soft)',
                color: 'var(--accent)',
                padding: '1px 5px',
                borderRadius: 4,
                fontWeight: 600,
                letterSpacing: '0.02em',
              }}
            >
              PRO
            </span>
          )}
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>· {thread.time_ago}</span>
          {thread.is_pin && (
            <span
              style={{
                fontSize: 11,
                color: 'var(--text-tertiary)',
                marginLeft: 'auto',
                letterSpacing: '-0.01em',
              }}
            >
              Pinned
            </span>
          )}
          <button
            style={{
              marginLeft: thread.is_pin ? 8 : 'auto',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text-tertiary)',
              padding: 2,
              borderRadius: 4,
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <MoreHorizontal size={15} strokeWidth={1.8} />
          </button>
        </div>

        {/* Content */}
        <div
          style={{
            fontSize: 15,
            color: 'var(--text)',
            lineHeight: 1.6,
            whiteSpace: 'pre-wrap',
            overflowWrap: 'break-word',
          }}
        >
          {thread.content}
        </div>

        {/* Media */}
        {thread.media.length > 0 && (
          <div style={{ display: 'grid', gap: 8, marginTop: 12 }}>
            {thread.media.map((m) => (
              <img
                key={m.media_id}
                src={m.url}
                alt=""
                style={{ maxWidth: '100%', borderRadius: 10, display: 'block', background: 'var(--bg-subtle)' }}
              />
            ))}
          </div>
        )}

        {/* Action buttons */}
        <div
          style={{
            display: 'flex',
            gap: 4,
            marginTop: 12,
          }}
        >
          <ActionBtn icon={<MessageCircle size={15} strokeWidth={1.8} />} label="Reply" active={false} />
          <ActionBtn
            icon={<Repeat2 size={15} strokeWidth={1.8} />}
            label={String(thread.stats.reposts)}
            active={thread.is_reposted ?? false}
            activeColor="var(--green)"
          />
          <ActionBtn
            icon={
              <Heart
                size={15}
                strokeWidth={1.8}
                fill={thread.is_liked ? 'var(--coral)' : 'none'}
                color={thread.is_liked ? 'var(--coral)' : 'currentColor'}
              />
            }
            label={String(thread.stats.likes)}
            active={thread.is_liked ?? false}
            activeColor="var(--coral)"
          />
          <ActionBtn
            icon={
              <Bookmark
                size={15}
                strokeWidth={1.8}
                fill={thread.is_bookmarked ? 'var(--accent)' : 'none'}
                color={thread.is_bookmarked ? 'var(--accent)' : 'currentColor'}
              />
            }
            label={String(thread.stats.bookmarks)}
            active={thread.is_bookmarked ?? false}
            activeColor="var(--accent)"
          />
        </div>
      </div>
    </article>
  );
}

interface ActionBtnProps {
  icon: ReactNode;
  label: string;
  active: boolean;
  activeColor?: string;
}

function ActionBtn({ icon, label, active, activeColor = 'var(--accent)' }: ActionBtnProps) {
  return (
    <button
      style={{
        background: 'none',
        border: 'none',
        cursor: 'pointer',
        display: 'flex',
        alignItems: 'center',
        gap: 5,
        color: active ? activeColor : 'var(--text-tertiary)',
        fontFamily: 'inherit',
        fontSize: 13,
        padding: '4px 8px',
        borderRadius: 6,
        marginLeft: -4,
        transition: 'color 0.1s',
      }}
    >
      {icon}
      {label !== 'Reply' && <span>{label}</span>}
    </button>
  );
}

// ─── Reply Card ───────────────────────────────────────────────────────────────

function ReplyCard({ thread, replyingTo }: { thread: ThreadItem; replyingTo: string }) {
  return (
    <div style={{ borderBottom: '1px solid var(--border)', background: 'var(--bg-card)' }}>
      <div
        style={{
          padding: '10px 20px 0 72px',
          fontSize: 13,
          color: 'var(--text-tertiary)',
          letterSpacing: '-0.01em',
        }}
      >
        Replying to{' '}
        <span style={{ color: 'var(--accent)', cursor: 'pointer' }}>@{replyingTo}</span>
      </div>
      <div style={{ borderBottom: 'none' }}>
        <ProfileThreadCard thread={thread} />
      </div>
    </div>
  );
}

// ─── Empty State ──────────────────────────────────────────────────────────────

function EmptyTab({ message, sub }: { message: string; sub: string }) {
  return (
    <div
      style={{
        padding: '80px 24px',
        textAlign: 'center',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 10,
      }}
    >
      <div style={{ fontSize: 20, fontWeight: 600, color: 'var(--text)', letterSpacing: '-0.02em' }}>
        {message}
      </div>
      <div style={{ fontSize: 14, color: 'var(--text-tertiary)', maxWidth: 280, lineHeight: 1.6 }}>
        {sub}
      </div>
    </div>
  );
}

// ─── Profile Page ─────────────────────────────────────────────────────────────
>>>>>>> 5622a22b23a8bf77c2e077de1b2cdfaaafe47e34

export default function ProfilePage() {
  const [activeTab, setActiveTab] = useState<ProfileTab>('threads');
  const [isFollowing, setIsFollowing] = useState(false);

  return (
    <div>
      <ProfilePageHeader name={MOCK_USER.name} threadCount={MOCK_THREADS.length} />
      <ProfileBanner bannerUrl={MOCK_USER.bannerUrl} />
      <ProfileInfo
        user={MOCK_USER}
        isOwn
        isFollowing={isFollowing}
        onFollowToggle={() => setIsFollowing(!isFollowing)}
      />
      <ProfileTabs activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'threads' && MOCK_THREADS.map((t) => (
        <ProfileThreadCard key={t.id} thread={t} />
      ))}
      {activeTab === 'replies' && MOCK_REPLIES.map((r) => (
        <ProfileThreadCard key={r.id} thread={r} replyingTo={r.replyingTo} />
      ))}
      {activeTab === 'media' && (
        <EmptyState title="No media yet" subtitle="Photos and videos from threads will appear here." />
      )}
      {activeTab === 'likes' && MOCK_LIKES.map((t) => (
        <ProfileThreadCard key={t.id} thread={t} />
      ))}
    </div>
  );
}
