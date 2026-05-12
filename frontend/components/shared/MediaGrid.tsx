'use client';

import { useRouter } from 'next/navigation';
import type { ThreadItem } from '@/lib/services/threads';

function isVideo(url: string): boolean {
  const ext = url.toLowerCase().split('?')[0];
  return ext.endsWith('.mp4') || ext.endsWith('.webm') || ext.endsWith('.mov') || ext.endsWith('.ogg');
}

function MultipleMediaIcon() {
  return (
    <div
      style={{
        position: 'absolute',
        top: 6,
        right: 6,
        background: 'rgba(0,0,0,0.65)',
        borderRadius: 5,
        width: 24,
        height: 24,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {/* Stacked squares — same as X's multiple-media indicator */}
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="8" y="8" width="13" height="13" rx="2" />
        <path d="M4 16V5a2 2 0 0 1 2-2h11" />
      </svg>
    </div>
  );
}

interface MediaGridProps {
  threads: ThreadItem[];
}

export function MediaGrid({ threads }: MediaGridProps) {
  const router = useRouter();

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 2 }}>
      {threads.map((thread) => {
        const first = thread.media[0];
        return (
          <div
            key={thread.id}
            onClick={() => router.push(`/thread/${thread.id}`)}
            style={{
              position: 'relative',
              aspectRatio: '1 / 1',
              overflow: 'hidden',
              cursor: 'pointer',
              background: 'var(--bg-subtle)',
            }}
          >
            {isVideo(first.url) ? (
              <video
                src={first.url}
                muted
                preload="metadata"
                style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block', pointerEvents: 'none' }}
              />
            ) : (
              <img
                src={first.url}
                alt=""
                loading="lazy"
                style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
              />
            )}
            {thread.media.length > 1 && <MultipleMediaIcon />}
          </div>
        );
      })}
    </div>
  );
}
