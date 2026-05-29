'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { MoreHorizontal } from 'lucide-react';
import { useTopTrendingHashtags } from '@/hooks/useHashtags';
import { ReportTrendModal, type AnchorRect } from '@/components/hashtags/ReportTrendModal';
import { useAuthStore } from '@/stores/authStore';

const SKELETON_COUNT = 4;

interface TrendingHashtagsProps {
  title?: string;
  /** Flat full-width section (no rounded card chrome) — for in-page use like X. */
  bare?: boolean;
  /** Cap the number of hashtags shown (right panel = 4, explore = 6). */
  limit?: number;
}

function getHashtagHref(name: string) {
  return `/explore?q=${encodeURIComponent(`#${name}`)}`;
}

export function TrendingHashtags({ title = 'Trending on campus', bare = false, limit }: TrendingHashtagsProps) {
  const router = useRouter();
  const token = useAuthStore((s) => s.token);
  const [reporting, setReporting] = useState<string | null>(null);
  const [anchorRect, setAnchorRect] = useState<AnchorRect | null>(null);
  const [reported, setReported] = useState(false);
  const trendingQuery = useTopTrendingHashtags();

  const sectionClass = bare ? '' : 'overflow-hidden rounded-2xl';
  const sectionStyle = bare
    ? { borderBottom: '1px solid var(--border)' }
    : { background: 'var(--bg-card)', border: '1px solid var(--border)' };
  const headCls = bare ? 'text-[20px]' : 'text-[18px]';
  const headPad = bare ? '16px 16px 6px' : '14px 16px 8px';

  if (trendingQuery.isLoading) {
    return (
      <div className={sectionClass} style={sectionStyle}>
        <h2 className={`${headCls} font-extrabold`} style={{ padding: headPad, color: 'var(--text)' }}>{title}</h2>
        <div>
          {Array.from({ length: SKELETON_COUNT }, (_, i) => (
            <div key={i} style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 6 }}>
              <div className="h-2.5 w-20 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
              <div className="h-3.5 w-32 animate-pulse rounded" style={{ background: 'var(--bg-subtle)' }} />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (trendingQuery.isError) {
    return (
      <div className={sectionClass} style={sectionStyle}>
        <h2 className={`${headCls} font-extrabold`} style={{ padding: headPad, color: 'var(--text)' }}>{title}</h2>
        <p className="text-sm" style={{ padding: '4px 16px 14px', color: 'var(--text-tertiary)' }}>Unable to load trending hashtags.</p>
      </div>
    );
  }

  const allHashtags = trendingQuery.data ?? [];
  const hashtags = typeof limit === 'number' ? allHashtags.slice(0, limit) : allHashtags;
  if (hashtags.length === 0) return null;

  return (
    <section className={sectionClass} style={sectionStyle}>
      <h2 className={`${headCls} font-extrabold`} style={{ padding: headPad, color: 'var(--text)' }}>{title}</h2>

      {reported && (
        <div
          style={{
            margin: '4px 12px 8px',
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

      <div>
        {hashtags.map((hashtag) => (
          <Link
            key={hashtag.name}
            href={getHashtagHref(hashtag.name)}
            className="group flex items-start justify-between transition-colors"
            style={{ padding: '6px 16px', color: 'inherit' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.03)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <div className="min-w-0">
              <p className="text-[12px]" style={{ color: 'var(--text-tertiary)' }}>Trending · Campus</p>
              <p className="text-[14px] font-bold" style={{ marginTop: 1, color: 'var(--text)' }}>#{hashtag.name}</p>
              {hashtag.threadCount != null && (
                <p className="text-[12px]" style={{ marginTop: 1, color: 'var(--text-tertiary)' }}>
                  {hashtag.threadCount.toLocaleString()} {hashtag.threadCount === 1 ? 'post' : 'posts'}
                </p>
              )}
            </div>
            <button
              type="button"
              onClick={(e) => {
                e.preventDefault();
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
              style={{ marginRight: -4, marginTop: -2, color: 'var(--text-tertiary)' }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--x-blue-soft)'; (e.currentTarget as HTMLElement).style.color = 'var(--x-blue)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; (e.currentTarget as HTMLElement).style.color = 'var(--text-tertiary)'; }}
              aria-label="More options"
            >
              <MoreHorizontal size={16} />
            </button>
          </Link>
        ))}
      </div>

      <button
        type="button"
        onClick={() => router.push('/explore?tab=trending')}
        className="w-full text-left text-[14px] transition-colors"
        style={{ padding: '10px 16px', color: 'var(--accent)' }}
        onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.03)'; }}
        onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
      >
        Show more
      </button>

      <ReportTrendModal
        open={reporting !== null}
        onOpenChange={(o) => { if (!o) { setReporting(null); setAnchorRect(null); } }}
        hashtagName={reporting}
        anchorRect={anchorRect}
        onReported={() => setReported(true)}
      />
    </section>
  );
}
