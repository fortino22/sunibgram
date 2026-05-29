'use client';

interface PagerProps {
  page: number;
  totalPages: number;
  total?: number;
  onChange: (p: number) => void;
}

export function Pager({ page, totalPages, total, onChange }: PagerProps) {
  return (
    <div style={{ marginTop: 16, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
        {totalPages > 0
          ? `Page ${page} of ${totalPages}${typeof total === 'number' ? ` (${total} total)` : ''}`
          : ''}
      </span>
      <div style={{ display: 'flex', gap: 8 }}>
        <button
          type="button"
          disabled={page <= 1}
          onClick={() => onChange(Math.max(1, page - 1))}
          className="admin-btn admin-btn-outline admin-btn-sm"
        >
          Prev
        </button>
        <button
          type="button"
          disabled={page >= totalPages}
          onClick={() => onChange(page + 1)}
          className="admin-btn admin-btn-outline admin-btn-sm"
        >
          Next
        </button>
      </div>
    </div>
  );
}
