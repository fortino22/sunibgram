'use client';

import { useState } from 'react';
import { ArrowLeft, Bookmark, Heart, MessageCircle, Repeat2 } from 'lucide-react';
import { useRouter } from 'next/navigation';
import type { ThreadItem } from '@/lib/services/threads';
import { ActionButton } from './ActionButton';
import { ReplyTree } from './ReplyTree';

interface ThreadDetailViewProps {
  root: ThreadItem;
  replies: ThreadItem[];
  totalReplies: number;
}

type ReplyTab = 'top' | 'latest';

function tabStyle(active: boolean): React.CSSProperties {
  return {
    flex: 1,
    border: 0,
    borderBottom: active ? '2px solid var(--accent)' : '2px solid transparent',
    background: 'transparent',
    color: active ? 'var(--text)' : 'var(--text-tertiary)',
    cursor: 'pointer',
    font: 'inherit',
    fontSize: 14,
    fontWeight: active ? 700 : 500,
    padding: '14px 12px 12px',
  };
}

export function ThreadDetailView({ root, replies, totalReplies }: ThreadDetailViewProps) {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<ReplyTab>('top');
  const isLiked = root.is_liked === true;
  const isReposted = root.is_reposted === true;
  const isBookmarked = root.is_bookmarked === true;
  const authorInitial = root.author.username.charAt(0).toUpperCase();

  const sortedReplies =
    activeTab === 'latest'
      ? [...replies].sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at))
      : replies;

  return (
    <main
      style={{
        maxWidth: 600,
        margin: '0 auto',
        minHeight: '100vh',
        background: 'var(--bg-card)',
        borderLeft: '1px solid var(--border)',
        borderRight: '1px solid var(--border)',
      }}
    >
      <header
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 20,
          display: 'flex',
          alignItems: 'center',
          gap: 18,
          padding: '12px 16px',
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <button
          type="button"
          aria-label="Back"
          onClick={() => router.back()}
          style={{
            width: 36,
            height: 36,
            border: 0,
            borderRadius: '50%',
            background: 'transparent',
            color: 'var(--text)',
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
          }}
        >
          <ArrowLeft size={20} strokeWidth={2.2} />
        </button>
        <div style={{ color: 'var(--text)', fontSize: 20, fontWeight: 700 }}>Thread</div>
      </header>

      <section style={{ padding: 16, borderBottom: '1px solid var(--border)' }}>
        <div style={{ display: 'flex', gap: 12, marginBottom: 14 }}>
          <div
            aria-hidden="true"
            style={{
              width: 48,
              height: 48,
              borderRadius: '50%',
              background: 'oklch(58% 0.23 270)',
              color: '#fff',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
              fontWeight: 700,
              fontSize: 18,
            }}
          >
            {authorInitial}
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ color: 'var(--text)', fontWeight: 700, fontSize: 16 }}>
              @{root.author.username}
            </div>
            {root.author.faculty && (
              <span
                style={{
                  display: 'inline-block',
                  marginTop: 3,
                  color: 'var(--accent)',
                  background: 'var(--bg-subtle)',
                  border: '1px solid var(--border)',
                  borderRadius: 999,
                  padding: '2px 8px',
                  fontSize: 12,
                  fontWeight: 700,
                }}
              >
                {root.author.faculty.faculty_name}
              </span>
            )}
          </div>
        </div>

        <div
          style={{
            color: 'var(--text)',
            fontSize: 18,
            lineHeight: 1.6,
            whiteSpace: 'pre-wrap',
            overflowWrap: 'break-word',
          }}
        >
          {root.content}
        </div>

        {root.media.length > 0 && (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: root.media.length === 1 ? '1fr' : 'repeat(2, minmax(0, 1fr))',
              gap: 8,
              marginTop: 14,
            }}
          >
            {root.media.map((m) => (
              <img
                key={m.media_id}
                src={m.url}
                alt=""
                style={{
                  width: '100%',
                  maxHeight: root.media.length === 1 ? 420 : 220,
                  borderRadius: 8,
                  objectFit: 'cover',
                  background: 'var(--bg-subtle)',
                }}
              />
            ))}
          </div>
        )}

        <div
          style={{
            color: 'var(--text-tertiary)',
            fontSize: 13,
            marginTop: 14,
            paddingBottom: 14,
            borderBottom: '1px solid var(--border)',
          }}
        >
          {root.time_ago}
        </div>

        <div
          style={{
            display: 'flex',
            gap: 20,
            paddingTop: 10,
            color: 'var(--text-tertiary)',
            fontSize: 14,
          }}
        >
          <span>
            <strong style={{ color: 'var(--text)' }}>{root.stats.likes}</strong> Likes
          </span>
          <span>
            <strong style={{ color: 'var(--text)' }}>{root.stats.reposts}</strong> Reposts
          </span>
          <span>
            <strong style={{ color: 'var(--text)' }}>{root.stats.bookmarks}</strong> Bookmarks
          </span>
        </div>

        <div
          style={{
            display: 'flex',
            gap: 4,
            marginTop: 10,
            paddingTop: 10,
            borderTop: '1px solid var(--border)',
            color: 'var(--text-tertiary)',
          }}
        >
          <ActionButton aria-label="Like" color={isLiked ? '#f04452' : undefined}>
            <Heart size={20} strokeWidth={2} fill={isLiked ? '#f04452' : 'none'} />
          </ActionButton>
          <ActionButton aria-label="Comment">
            <MessageCircle size={20} strokeWidth={2} />
          </ActionButton>
          <ActionButton aria-label="Repost" color={isReposted ? 'var(--accent)' : undefined}>
            <Repeat2 size={20} strokeWidth={2} />
          </ActionButton>
          <ActionButton aria-label="Bookmark" color={isBookmarked ? 'var(--accent)' : undefined}>
            <Bookmark size={20} strokeWidth={2} fill={isBookmarked ? 'var(--accent)' : 'none'} />
          </ActionButton>
        </div>
      </section>

      <section style={{ display: 'flex', gap: 12, padding: 16, borderBottom: '1px solid var(--border)' }}>
        <div
          aria-hidden="true"
          style={{
            width: 40,
            height: 40,
            borderRadius: '50%',
            background: 'oklch(58% 0.23 270)',
            flexShrink: 0,
          }}
        />
        <textarea
          placeholder={`Reply to @${root.author.username}...`}
          rows={3}
          style={{
            flex: 1,
            resize: 'vertical',
            minHeight: 76,
            border: 0,
            outline: 'none',
            background: 'transparent',
            color: 'var(--text)',
            font: 'inherit',
            fontSize: 15,
            lineHeight: 1.5,
          }}
        />
      </section>

      <section style={{ borderBottom: '1px solid var(--border)' }}>
        <div style={{ padding: '14px 16px 0', color: 'var(--text)', fontWeight: 700, fontSize: 16 }}>
          {totalReplies} {totalReplies === 1 ? 'Reply' : 'Replies'}
        </div>
        <div style={{ display: 'flex', marginTop: 2 }}>
          <button type="button" onClick={() => setActiveTab('top')} style={tabStyle(activeTab === 'top')}>
            Top
          </button>
          <button type="button" onClick={() => setActiveTab('latest')} style={tabStyle(activeTab === 'latest')}>
            Latest
          </button>
        </div>
      </section>

      <ReplyTree threads={sortedReplies} rootId={root.id} />
    </main>
  );
}
