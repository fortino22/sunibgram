'use client';

import { useState } from 'react';
import { ThreadCard } from '@/components/thread/ThreadCard';
import { Icon } from '@/components/ui/Icon';
import { useThreads } from '@/hooks/useThreads';
import type { ThreadItem } from '@/lib/services/threads';

type FeedTabId = 'for-you' | 'following' | 'cs' | 'events';

interface Tab {
  id: FeedTabId;
  label: string;
}

const TABS: Tab[] = [
  { id: 'for-you',   label: 'For You' },
  { id: 'following', label: 'Following' },
  { id: 'cs',        label: 'Comp Sci' },
  { id: 'events',    label: 'Events' },
];

const EMPTY_MESSAGES: Record<FeedTabId, { title: string; subtitle: string }> = {
  'for-you':   { title: 'Nothing here yet',                   subtitle: 'Posts from campus will appear here.' },
  'following': { title: 'No posts from people you follow',    subtitle: 'Follow more people to see their posts here.' },
  'cs':        { title: 'No Comp Sci posts yet',              subtitle: 'Posts from Computer Science students will appear here.' },
  'events':    { title: 'No events or announcements yet',     subtitle: 'Campus events and announcements will appear here.' },
};

interface FeedTabButtonProps {
  tab: Tab;
  active: boolean;
  onClick: () => void;
}

function FeedTabButton({ tab, active, onClick }: FeedTabButtonProps) {
  return (
    <button
      onClick={onClick}
      style={{
        padding: '8px 16px',
        background: 'none',
        border: 'none',
        cursor: 'pointer',
        fontWeight: active ? 700 : 400,
        fontSize: 14,
        color: active ? 'var(--text)' : 'var(--text-tertiary)',
        borderBottom: active ? '2px solid var(--accent)' : '2px solid transparent',
        marginBottom: -1,
        transition: 'all 0.15s ease',
        letterSpacing: '-0.01em',
        fontFamily: 'inherit',
      }}
    >
      {tab.label}
    </button>
  );
}

function EmptyFeed({ tab }: { tab: FeedTabId }) {
  const { title, subtitle } = EMPTY_MESSAGES[tab];
  return (
    <div
      style={{
        padding: '80px 24px',
        textAlign: 'center',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 12,
      }}
    >
      <div
        style={{
          width: 52,
          height: 52,
          borderRadius: '50%',
          background: 'var(--bg-subtle)',
          border: '1px solid var(--border)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          marginBottom: 4,
        }}
      >
        <Icon name="reply" size={22} color="var(--border)" strokeWidth={1.2} />
      </div>
      <div
        style={{
          fontWeight: 500,
          fontSize: 18,
          color: 'var(--text)',
          letterSpacing: '-0.02em',
        }}
      >
        {title}
      </div>
      <div
        style={{
          fontSize: 14,
          color: 'var(--text-tertiary)',
          maxWidth: 280,
          lineHeight: 1.6,
        }}
      >
        {subtitle}
      </div>
    </div>
  );
}

function PostComposerShell() {
  const ATTACH_BUTTONS = [
    { icon: 'image', label: 'Image' },
    { icon: 'video', label: 'Video' },
    { icon: 'poll',  label: 'Poll' },
    { icon: 'hash',  label: 'Topic' },
  ];

  return (
    <div
      style={{
        padding: '16px 20px',
        background: 'var(--bg-card)',
        borderBottom: '1px solid var(--border)',
      }}
    >
      <div style={{ display: 'flex', gap: 12 }}>
        {/* Current user avatar placeholder */}
        <div
          style={{
            width: 40,
            height: 40,
            borderRadius: '50%',
            background: 'oklch(58% 0.23 270)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontWeight: 500,
            fontSize: 15,
            color: '#fff',
            flexShrink: 0,
            letterSpacing: '-0.02em',
          }}
        >
          AP
        </div>
        <div style={{ flex: 1 }}>
          <div
            style={{
              width: '100%',
              fontSize: 15,
              lineHeight: 1.65,
              color: 'var(--text-tertiary)',
              padding: '6px 0 10px',
              cursor: 'text',
            }}
          >
            What's happening on campus?
          </div>
          <div style={{ height: 1, background: 'var(--border)', margin: '4px 0 10px' }} />
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', gap: 2 }}>
              {ATTACH_BUTTONS.map((btn) => (
                <button
                  key={btn.label}
                  title={btn.label}
                  style={{
                    background: 'none',
                    border: 'none',
                    cursor: 'not-allowed',
                    color: 'var(--text-tertiary)',
                    padding: '5px 7px',
                    borderRadius: 6,
                    display: 'flex',
                    alignItems: 'center',
                    opacity: 0.5,
                  }}
                  disabled
                >
                  <Icon name={btn.icon} size={16} color="currentColor" />
                </button>
              ))}
            </div>
            <button
              disabled
              style={{
                background: 'var(--accent)',
                border: 'none',
                cursor: 'not-allowed',
                borderRadius: 10,
                color: '#fff',
                fontWeight: 500,
                fontSize: 13,
                padding: '6px 14px',
                opacity: 0.45,
                fontFamily: 'inherit',
              }}
            >
              Post
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function HomePage() {
  const [activeTab, setActiveTab] = useState<FeedTabId>('for-you');
  const threadsQuery = useThreads();

  return (
    <div>
      {/* Sticky header with tabs */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div style={{ padding: '14px 20px 0' }}>
          <div
            style={{
              fontWeight: 500,
              fontSize: 20,
              color: 'var(--text)',
              letterSpacing: '-0.02em',
              marginBottom: 12,
            }}
          >
            Home
          </div>
          <div style={{ display: 'flex', gap: 0 }}>
            {TABS.map((t) => (
              <FeedTabButton
                key={t.id}
                tab={t}
                active={activeTab === t.id}
                onClick={() => setActiveTab(t.id)}
              />
            ))}
          </div>
        </div>
      </div>

      {/* Composer shell */}
      <PostComposerShell />

      {activeTab === 'for-you' ? (
        <div>
          {threadsQuery.isLoading
            ? Array.from({ length: 3 }, (_, index) => (
                <div
                  key={index}
                  style={{
                    background: 'var(--bg-subtle)',
                    borderRadius: '8px',
                    height: 80,
                    marginBottom: 8,
                  }}
                />
              ))
            : threadsQuery.data?.data.map((thread: ThreadItem) => (
                <ThreadCard key={thread.id} thread={thread} />
              ))}
        </div>
      ) : (
        <EmptyFeed tab={activeTab} />
      )}
    </div>
  );
}
