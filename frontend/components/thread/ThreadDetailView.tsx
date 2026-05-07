'use client';

import { useState } from 'react';
import { ArrowLeft } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { Avatar } from '@/components/shared/Avatar';
import type { ThreadItem } from '@/lib/services/threads';
import { ThreadActions } from './ThreadActions';
import { ThreadBody } from './ThreadBody';
import { ThreadComposer } from './ThreadComposer';
import { ThreadHeader } from './ThreadHeader';
import { ReplyTree } from './ReplyTree';
import { getAuthorAvatarUrl, getThreadAuthor } from './threadUtils';

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
  const author = getThreadAuthor(root);

  const sortedReplies =
    activeTab === 'latest'
      ? [...replies].sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at))
      : replies;

  return (
    <main
      className="mx-auto min-h-screen max-w-[600px]"
      style={{
        background: 'var(--bg-card)',
        borderLeft: '1px solid var(--border)',
        borderRight: '1px solid var(--border)',
      }}
    >
      <header
        className="sticky top-0 z-20 flex items-center gap-[18px] border-b px-4 py-3"
        style={{ background: 'var(--bg-card)', borderColor: 'var(--border)' }}
      >
        <button
          type="button"
          aria-label="Back"
          onClick={() => router.back()}
          className="inline-flex h-9 w-9 items-center justify-center rounded-full"
          style={{
            background: 'transparent',
            color: 'var(--text)',
          }}
        >
          <ArrowLeft size={20} strokeWidth={2.2} />
        </button>
        <div className="text-xl font-bold" style={{ color: 'var(--text)' }}>Thread</div>
      </header>

      <section className="border-b p-4" style={{ borderColor: 'var(--border)' }}>
        <div className="mb-3.5 flex gap-3">
          <Avatar username={author.username} avatarUrl={getAuthorAvatarUrl(author)} size={48} />
          <div className="min-w-0 flex-1">
            <ThreadHeader author={author} showFaculty />
          </div>
        </div>

        <ThreadBody content={root.content} media={root.media} />

        <div
          className="mt-3.5 border-b pb-3.5 text-[13px]"
          style={{ color: 'var(--text-tertiary)', borderColor: 'var(--border)' }}
        >
          {root.time_ago}
        </div>

        <div
          className="flex gap-5 pt-2.5 text-sm"
          style={{ color: 'var(--text-tertiary)' }}
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

        <ThreadActions
          likes={root.stats.likes}
          reposts={root.stats.reposts}
          bookmarks={root.stats.bookmarks}
          isLiked={root.is_liked === true}
          isReposted={root.is_reposted === true}
          isBookmarked={root.is_bookmarked === true}
          className="justify-start gap-1"
        />
      </section>

      <ThreadComposer placeholder={`Reply to @${root.author.username}...`} />

      <section className="border-b" style={{ borderColor: 'var(--border)' }}>
        <div className="px-4 pt-3.5 text-base font-bold" style={{ color: 'var(--text)' }}>
          {totalReplies} {totalReplies === 1 ? 'Reply' : 'Replies'}
        </div>
        <div className="mt-0.5 flex">
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
