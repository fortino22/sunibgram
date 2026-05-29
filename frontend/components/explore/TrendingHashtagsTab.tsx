'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { MoreHorizontal } from 'lucide-react';
import { useTrendingHashtags } from '@/hooks/useHashtags';
import { EmptyState } from '@/components/shared/EmptyState';
import { ReportTrendModal, type AnchorRect } from '@/components/hashtags/ReportTrendModal';
import { useAuthStore } from '@/stores/authStore';

export function TrendingHashtagsTab() {
  const router = useRouter();
  const token = useAuthStore((s) => s.token);
  const [reporting, setReporting] = useState<string | null>(null);
  const [anchorRect, setAnchorRect] = useState<AnchorRect | null>(null);
  const [reported, setReported] = useState(false);
  const { data, isLoading, isError } = useTrendingHashtags();

  if (isLoading) {
    return (
      <div>
        {Array.from({ length: 10 }, (_, i) => (
          <div key={i} style={{ padding: '8px 16px', display: 'flex', flexDirection: 'column', gap: 4 }}>
            <div className="h-2.5 w-24 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-3.5 w-40 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-2.5 w-16 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
          </div>
        ))}
      </div>
    );
  }

  if (isError) {
    return <EmptyState title="Couldn't load trends" subtitle="Try again in a moment." />;
  }

  const hashtags = data ?? [];
  if (hashtags.length === 0) {
    return <EmptyState title="No trends yet" subtitle="Trending hashtags from campus will appear here." />;
  }

  const goToHashtag = (name: string) => {
    router.push(`/explore?q=${encodeURIComponent(`#${name}`)}`);
  };

  return (
    <div>
      {reported && (
        <div
          style={{
            margin: '8px 12px',
            padding: '10px 12px',
            borderRadius: 12,
            background: 'var(--bg-subtle)',
            color: 'var(--text-secondary)',
            fontSize: 13,
            lineHeight: 1.4,
          }}
        >
          Thanks. Refresh this page to update these trends.
        </div>
      )}

      {hashtags.map((hashtag, idx) => (
        <div
          key={hashtag.name}
          role="link"
          tabIndex={0}
          onClick={() => goToHashtag(hashtag.name)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              goToHashtag(hashtag.name);
            }
          }}
          style={{
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'space-between',
            gap: 12,
            width: '100%',
            padding: '6px 16px',
            background: 'transparent',
            cursor: 'pointer',
            textAlign: 'left',
          }}
          onMouseEnter={(e) => { e.currentTarget.style.background = 'rgba(255,255,255,0.03)'; }}
          onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
        >
          <div style={{ minWidth: 0 }}>
            <p className="text-[12px]" style={{ color: 'var(--text-tertiary)' }}>
              {idx + 1} · Trending · Campus
            </p>
            <p className="text-[14px] font-bold" style={{ marginTop: 1, color: 'var(--text)' }}>
              #{hashtag.name}
            </p>
            {hashtag.threadCount != null && (
              <p className="text-[12px]" style={{ marginTop: 1, color: 'var(--text-tertiary)' }}>
                {hashtag.threadCount.toLocaleString()} {hashtag.threadCount === 1 ? 'post' : 'posts'}
              </p>
            )}
          </div>
          <button
            type="button"
            aria-label="More options"
            onClick={(e) => {
              e.stopPropagation();
              if (!token) {
                router.push('/login');
                return;
              }
              const r = (e.currentTarget as HTMLElement).getBoundingClientRect();
              setAnchorRect({ top: r.top, left: r.left, bottom: r.bottom, right: r.right, width: r.width, height: r.height });
              setReporting(hashtag.name);
            }}
            className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full transition-colors"
            style={{ marginRight: -4, marginTop: -2, color: 'var(--text-tertiary)', background: 'transparent', border: 'none', cursor: 'pointer' }}
            onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--x-blue-soft)'; e.currentTarget.style.color = 'var(--x-blue)'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--text-tertiary)'; }}
          >
            <MoreHorizontal size={16} />
          </button>
        </div>
      ))}

      <ReportTrendModal
        open={reporting !== null}
        onOpenChange={(o) => { if (!o) { setReporting(null); setAnchorRect(null); } }}
        hashtagName={reporting}
        anchorRect={anchorRect}
        onReported={() => setReported(true)}
      />
    </div>
  );
}
