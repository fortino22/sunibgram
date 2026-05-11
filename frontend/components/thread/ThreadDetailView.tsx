'use client';

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
  viewFullThreadId?: number | null;
}

export function ThreadDetailView({ root, replies, viewFullThreadId = null }: ThreadDetailViewProps) {
  const router = useRouter();
  const author = getThreadAuthor(root);

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
        <div className="text-xl font-bold" style={{ color: 'var(--text)' }}>Post</div>
      </header>

      <article
        className="border-b"
        style={{
          borderColor: 'var(--border)',
          background: 'var(--bg-card)',
          padding: '14px 20px',
        }}
      >
        <div className="flex gap-3">
          <Avatar username={author.username} avatarUrl={getAuthorAvatarUrl(author)} size="md" />
          <div className="min-w-0 flex-1">
            <ThreadHeader author={author} timeAgo={root.time_ago} parentContext={root.parent_context} />
            <ThreadBody content={root.content} media={root.media} />

            <div
              className="mt-3 border-t pt-3 text-sm"
              style={{ color: 'var(--text-tertiary)', borderColor: 'var(--border)' }}
            >
              <div className="flex flex-wrap gap-x-5 gap-y-2">
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
            </div>

            <ThreadActions
              comments={root.stats.comments}
              likes={root.stats.likes}
              reposts={root.stats.reposts}
              bookmarks={root.stats.bookmarks}
              isLiked={root.is_liked === true}
              isReposted={root.is_reposted === true}
              isBookmarked={root.is_bookmarked === true}
            />
          </div>
        </div>
      </article>

      <ThreadComposer
        placeholder={`Post your reply to @${root.author.username}`}
        parentId={root.id}
      />

      {viewFullThreadId !== null && (
        <section
          className="border-b"
          style={{
            borderColor: 'var(--border)',
            background: 'var(--bg-card)',
            padding: '14px 20px',
          }}
        >
          <button
            type="button"
            onClick={() => router.push(`/thread/${viewFullThreadId}`)}
            className="rounded-full px-4 py-2 text-sm font-semibold"
            style={{
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
            }}
          >
            View full thread
          </button>
        </section>
      )}

      <ReplyTree threads={replies} />
    </main>
  );
}
