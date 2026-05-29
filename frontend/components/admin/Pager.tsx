'use client';

import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PagerProps {
  page: number;
  totalPages: number;
  total?: number;
  onChange: (p: number) => void;
}

type PageEntry = number | 'ellipsis-left' | 'ellipsis-right';

export function getPageList(current: number, total: number): PageEntry[] {
  if (total <= 1) return [1];
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);

  const entries: PageEntry[] = [1];
  const leftBound = Math.max(2, current - 1);
  const rightBound = Math.min(total - 1, current + 1);

  if (leftBound > 2) entries.push('ellipsis-left');
  for (let i = leftBound; i <= rightBound; i++) entries.push(i);
  if (rightBound < total - 1) entries.push('ellipsis-right');

  entries.push(total);
  return entries;
}

export function Pager({ page, totalPages, total, onChange }: PagerProps) {
  if (totalPages <= 1 && (!total || total === 0)) {
    return null;
  }
  const entries = getPageList(page, totalPages);

  return (
    <nav className="admin-pager" aria-label="Pagination">
      <span className="admin-pager-meta">
        {totalPages > 0
          ? `Page ${page} of ${totalPages}${typeof total === 'number' ? ` (${total} total)` : ''}`
          : ''}
      </span>
      <div className="admin-pager-pages">
        <button
          type="button"
          className="admin-pager-btn admin-pager-step"
          disabled={page <= 1}
          onClick={() => onChange(Math.max(1, page - 1))}
          aria-label="Previous page"
        >
          <ChevronLeft size={15} strokeWidth={2.2} />
          <span>Prev</span>
        </button>
        {entries.map((entry, i) =>
          typeof entry === 'number' ? (
            <button
              key={`p-${entry}`}
              type="button"
              className={`admin-pager-btn${entry === page ? ' admin-pager-btn-active' : ''}`}
              onClick={() => entry !== page && onChange(entry)}
              aria-current={entry === page ? 'page' : undefined}
              aria-label={`Go to page ${entry}`}
            >
              {entry}
            </button>
          ) : (
            <span key={`${entry}-${i}`} className="admin-pager-ellipsis" aria-hidden>
              …
            </span>
          ),
        )}
        <button
          type="button"
          className="admin-pager-btn admin-pager-step"
          disabled={page >= totalPages}
          onClick={() => onChange(page + 1)}
          aria-label="Next page"
        >
          <span>Next</span>
          <ChevronRight size={15} strokeWidth={2.2} />
        </button>
      </div>
    </nav>
  );
}
