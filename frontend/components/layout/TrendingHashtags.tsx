'use client';

import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { MoreHorizontal } from 'lucide-react';
import { useTopTrendingHashtags } from '@/hooks/useHashtags';

const SKELETON_COUNT = 4;

interface TrendingHashtagsProps {
  title?: string;
  /** Flat full-width section (no rounded card chrome) — for in-page use like X. */
  bare?: boolean;
}

function getHashtagHref(name: string) {
  return `/explore?q=${encodeURIComponent(`#${name}`)}`;
}

export function TrendingHashtags({ title = 'Trending on campus', bare = false }: TrendingHashtagsProps) {
  const router = useRouter();
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

  const hashtags = trendingQuery.data ?? [];
  if (hashtags.length === 0) return null;

  return (
    <section className={sectionClass} style={sectionStyle}>
      <h2 className={`${headCls} font-extrabold`} style={{ padding: headPad, color: 'var(--text)' }}>{title}</h2>

      <div>
        {hashtags.map((hashtag) => (
          <Link
            key={hashtag.name}
            href={getHashtagHref(hashtag.name)}
            className="group flex items-start justify-between transition-colors"
            style={{ padding: '12px 16px', color: 'inherit' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-hover)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <div className="min-w-0">
              <p className="text-[13px]" style={{ color: 'var(--text-tertiary)' }}>Trending · Campus</p>
              <p className="text-[15px] font-bold" style={{ marginTop: 2, color: 'var(--text)' }}>#{hashtag.name}</p>
              {hashtag.threadCount != null && (
                <p className="text-[13px]" style={{ marginTop: 2, color: 'var(--text-tertiary)' }}>
                  {hashtag.threadCount.toLocaleString()} {hashtag.threadCount === 1 ? 'post' : 'posts'}
                </p>
              )}
            </div>
            <button
              type="button"
              onClick={(e) => e.preventDefault()}
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full transition-colors"
              style={{ marginRight: -6, marginTop: -2, color: 'var(--text-tertiary)' }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--x-blue-soft)'; (e.currentTarget as HTMLElement).style.color = 'var(--x-blue)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; (e.currentTarget as HTMLElement).style.color = 'var(--text-tertiary)'; }}
              aria-label="More options"
            >
              <MoreHorizontal size={18} />
            </button>
          </Link>
        ))}
      </div>

      <button
        type="button"
        onClick={() => router.push('/explore?tab=trending')}
        className="w-full text-left text-[15px] transition-colors"
        style={{ padding: '12px 16px', color: 'var(--accent)' }}
        onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--bg-hover)'; }}
        onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
      >
        Show more
      </button>
    </section>
  );
}
