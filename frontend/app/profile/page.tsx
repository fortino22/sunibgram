'use client';

import { useState, type ReactNode, type MouseEvent } from 'react';
import {
  ArrowLeft,
  Calendar,
  BookOpen,
  Heart,
  Repeat2,
  MessageCircle,
  Bookmark,
  MoreHorizontal,
} from 'lucide-react';
import type { ThreadItem } from '@/lib/services/threads';

// ─── Mock Data ────────────────────────────────────────────────────────────────

const MOCK_USER = {
  id: 1,
  name: 'Aria Putri',
  username: 'ariaputri',
  bio: 'Computer Science student at Binus 🎓 | Tech enthusiast, coffee addict, and occasional night owl. Building things that matter.',
  faculty: 'School of Computer Science',
  joinedMonth: 'March 2024',
  following: 142,
  followers: 1847,
  isCurrentUser: true,
};

function makeThread(
  id: number,
  content: string,
  overrides: Partial<ThreadItem> & { author?: ThreadItem['author'] } = {},
): ThreadItem {
  const { author, ...rest } = overrides;
  return {
    type: 'POST',
    id,
    parent_id: null,
    root_id: null,
    content,
    created_at: '2026-04-01T10:00:00Z',
    updated_at: '2026-04-01T10:00:00Z',
    time_ago: '1h',
    author: author ?? { id: 1, username: 'ariaputri', is_premium: false },
    stats: { likes: 24, reposts: 6, bookmarks: 3 },
    media: [],
    is_liked: false,
    is_bookmarked: false,
    is_reposted: false,
    is_pin: false,
    ...rest,
  };
}

const MOCK_THREADS: ThreadItem[] = [
  makeThread(
    1,
    'Just finished my final project on distributed systems and honestly it feels surreal. Four months of suffering compressed into 47 slides 😭 #BinusLife',
    { time_ago: '2h', stats: { likes: 84, reposts: 12, bookmarks: 31 }, is_liked: true, is_pin: true },
  ),
  makeThread(
    2,
    'Hot take: the library at 11 pm hits different. Something about the quiet desperation of everyone around you that really fuels productivity.',
    { time_ago: '1d', stats: { likes: 247, reposts: 38, bookmarks: 15 }, is_reposted: true },
  ),
  makeThread(
    3,
    "Starting a study group for Algorithms next week. DM me if you're in SOCS and struggling with dynamic programming — we're all struggling together.",
    { time_ago: '3d', stats: { likes: 56, reposts: 22, bookmarks: 8 } },
  ),
  makeThread(
    4,
    'Random thought: why does every group project have one person who disappears until the day before the deadline and then suddenly has "great ideas"?',
    { time_ago: '5d', stats: { likes: 412, reposts: 67, bookmarks: 29 }, is_liked: true },
  ),
];

type ReplyThread = ThreadItem & { replyingTo: string };

const MOCK_REPLIES: ReplyThread[] = [
  {
    ...makeThread(
      5,
      'Agree 100%! The vending machine by the computer lab is a lifesaver at 2 am 😂',
      { time_ago: '3h', stats: { likes: 18, reposts: 2, bookmarks: 1 } },
    ),
    replyingTo: 'binusofficialpage',
  },
  {
    ...makeThread(
      6,
      'This is exactly why I switched to Obsidian for notes. Game changer for organizing lecture slides.',
      { time_ago: '2d', stats: { likes: 29, reposts: 5, bookmarks: 11 } },
    ),
    replyingTo: 'notesnerdbinus',
  },
  {
    ...makeThread(
      7,
      "Honestly the grading rubric was unclear from the start but I get it, the TA's are overwhelmed too.",
      { time_ago: '4d', stats: { likes: 41, reposts: 3, bookmarks: 2 } },
    ),
    replyingTo: 'socs_helpdesk',
  },
];

const MOCK_LIKES: ThreadItem[] = [
  makeThread(
    8,
    "Reminder that the hackathon registration closes this Friday! Don't sleep on it 🔥 #BinusHackathon2026",
    {
      time_ago: '5h',
      stats: { likes: 198, reposts: 44, bookmarks: 27 },
      is_liked: true,
      author: { id: 9, username: 'binusteach', is_premium: true },
    },
  ),
  makeThread(
    9,
    "The Binus cafeteria actually upgraded their menu and nobody is talking about this?? The new pasta section goes crazy",
    {
      time_ago: '1d',
      stats: { likes: 331, reposts: 78, bookmarks: 12 },
      is_liked: true,
      author: { id: 2, username: 'campusfoodie', is_premium: false },
    },
  ),
  makeThread(
    10,
    'PSA: you can book private study rooms in the library app now. No more fighting for spots 🙌',
    {
      time_ago: '3d',
      stats: { likes: 612, reposts: 156, bookmarks: 89 },
      is_liked: true,
      author: { id: 3, username: 'libraryupdate', is_premium: false },
    },
  ),
];

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
            active={thread.is_reposted}
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
            active={thread.is_liked}
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
            active={thread.is_bookmarked}
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

export default function ProfilePage() {
  const [activeTab, setActiveTab] = useState<ProfileTab>('threads');
  const [isFollowing, setIsFollowing] = useState(false);

  const isOwn = MOCK_USER.isCurrentUser;
  const threadCount = MOCK_THREADS.length;

  return (
    <div>
      {/* ── Sticky Header ── */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 20,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '10px 16px',
          display: 'flex',
          alignItems: 'center',
          gap: 14,
        }}
      >
        <button
          style={{
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            color: 'var(--text)',
            padding: 6,
            borderRadius: '50%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}
        >
          <ArrowLeft size={20} strokeWidth={1.8} />
        </button>
        <div>
          <div
            style={{
              fontWeight: 700,
              fontSize: 17,
              color: 'var(--text)',
              letterSpacing: '-0.02em',
              lineHeight: 1.2,
            }}
          >
            {MOCK_USER.name}
          </div>
          <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
            {threadCount} thread{threadCount !== 1 ? 's' : ''}
          </div>
        </div>
      </div>

      {/* ── Banner ── */}
      <div
        style={{
          height: 200,
          background: 'linear-gradient(135deg, oklch(28% 0.09 260), oklch(18% 0.05 290), oklch(22% 0.07 270))',
          position: 'relative',
          flexShrink: 0,
        }}
      />

      {/* ── Avatar + Action Row ── */}
      <div style={{ position: 'relative', padding: '0 20px' }}>
        {/* Avatar — overlaps the banner */}
        <div
          style={{
            position: 'absolute',
            top: -44,
            left: 20,
            width: 84,
            height: 84,
            borderRadius: '50%',
            background: 'oklch(58% 0.23 270)',
            border: '4px solid var(--bg-card)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: 30,
            fontWeight: 700,
            color: '#fff',
            letterSpacing: '-0.02em',
            zIndex: 1,
          }}
        >
          {MOCK_USER.name.charAt(0)}
        </div>

        {/* Action button row */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'flex-end',
            gap: 10,
            paddingTop: 14,
          }}
        >
          {isOwn ? (
            <button
              style={{
                background: 'none',
                border: '1px solid var(--border-strong)',
                borderRadius: 20,
                padding: '7px 18px',
                fontSize: 14,
                fontWeight: 600,
                color: 'var(--text)',
                cursor: 'pointer',
                fontFamily: 'inherit',
                letterSpacing: '-0.01em',
                transition: 'border-color 0.15s',
              }}
            >
              Edit profile
            </button>
          ) : (
            <>
              <button
                style={{
                  background: 'none',
                  border: '1px solid var(--border-strong)',
                  borderRadius: '50%',
                  width: 36,
                  height: 36,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                  color: 'var(--text)',
                }}
              >
                <MoreHorizontal size={17} strokeWidth={1.8} />
              </button>
              <button
                onClick={() => setIsFollowing(!isFollowing)}
                style={{
                  background: isFollowing ? 'none' : 'var(--text)',
                  border: isFollowing ? '1px solid var(--border-strong)' : '1px solid transparent',
                  borderRadius: 20,
                  padding: '7px 20px',
                  fontSize: 14,
                  fontWeight: 700,
                  color: isFollowing ? 'var(--text)' : 'var(--bg)',
                  cursor: 'pointer',
                  fontFamily: 'inherit',
                  letterSpacing: '-0.01em',
                  transition: 'all 0.15s ease',
                }}
              >
                {isFollowing ? 'Following' : 'Follow'}
              </button>
            </>
          )}
        </div>

        {/* ── User Info ── */}
        <div style={{ marginTop: 46, paddingBottom: 16 }}>
          <div
            style={{
              fontWeight: 700,
              fontSize: 20,
              color: 'var(--text)',
              letterSpacing: '-0.03em',
              lineHeight: 1.2,
            }}
          >
            {MOCK_USER.name}
          </div>
          <div style={{ fontSize: 14, color: 'var(--text-tertiary)', marginTop: 2 }}>
            @{MOCK_USER.username}
          </div>

          {MOCK_USER.bio && (
            <div
              style={{
                fontSize: 15,
                color: 'var(--text)',
                lineHeight: 1.65,
                marginTop: 12,
              }}
            >
              {MOCK_USER.bio}
            </div>
          )}

          {/* Meta row */}
          <div
            style={{
              display: 'flex',
              flexWrap: 'wrap',
              gap: '8px 20px',
              marginTop: 12,
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                fontSize: 14,
                color: 'var(--text-tertiary)',
              }}
            >
              <BookOpen size={15} strokeWidth={1.6} />
              <span>{MOCK_USER.faculty}</span>
            </div>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                fontSize: 14,
                color: 'var(--text-tertiary)',
              }}
            >
              <Calendar size={15} strokeWidth={1.6} />
              <span>Joined {MOCK_USER.joinedMonth}</span>
            </div>
          </div>

          {/* Following / Followers */}
          <div style={{ display: 'flex', gap: 20, marginTop: 14 }}>
            <button
              style={{
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                padding: 0,
                fontFamily: 'inherit',
                display: 'flex',
                alignItems: 'baseline',
                gap: 5,
              }}
            >
              <span
                style={{
                  fontSize: 15,
                  fontWeight: 700,
                  color: 'var(--text)',
                  letterSpacing: '-0.02em',
                }}
              >
                {MOCK_USER.following.toLocaleString()}
              </span>
              <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>Following</span>
            </button>
            <button
              style={{
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                padding: 0,
                fontFamily: 'inherit',
                display: 'flex',
                alignItems: 'baseline',
                gap: 5,
              }}
            >
              <span
                style={{
                  fontSize: 15,
                  fontWeight: 700,
                  color: 'var(--text)',
                  letterSpacing: '-0.02em',
                }}
              >
                {MOCK_USER.followers.toLocaleString()}
              </span>
              <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>Followers</span>
            </button>
          </div>
        </div>
      </div>

      {/* ── Tab Bar ── */}
      <div
        style={{
          display: 'flex',
          borderBottom: '1px solid var(--border)',
          position: 'sticky',
          top: 57,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
        }}
      >
        {PROFILE_TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            style={{
              flex: 1,
              padding: '14px 0',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              fontSize: 14,
              fontWeight: activeTab === tab.id ? 700 : 400,
              color: activeTab === tab.id ? 'var(--text)' : 'var(--text-tertiary)',
              borderBottom: activeTab === tab.id ? '2px solid var(--accent)' : '2px solid transparent',
              fontFamily: 'inherit',
              letterSpacing: '-0.01em',
              transition: 'all 0.15s ease',
              marginBottom: -1,
            }}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* ── Tab Content ── */}
      {activeTab === 'threads' && (
        <div>
          {MOCK_THREADS.map((t) => (
            <ProfileThreadCard key={t.id} thread={t} />
          ))}
        </div>
      )}

      {activeTab === 'replies' && (
        <div>
          {MOCK_REPLIES.map((r) => (
            <ReplyCard key={r.id} thread={r} replyingTo={r.replyingTo} />
          ))}
        </div>
      )}

      {activeTab === 'media' && (
        <EmptyTab
          message="No media yet"
          sub="Photos and videos from threads will appear here."
        />
      )}

      {activeTab === 'likes' && (
        <div>
          {MOCK_LIKES.map((t) => (
            <ProfileThreadCard key={t.id} thread={t} />
          ))}
        </div>
      )}
    </div>
  );
}
