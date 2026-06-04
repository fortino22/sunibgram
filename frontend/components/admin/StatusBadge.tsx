'use client';

interface StatusBadgeProps {
  status: string;
}

const COLORS: Record<string, { bg: string; fg: string }> = {
  approved:    { bg: 'var(--green-soft)',  fg: 'var(--green)' },
  pending:     { bg: 'var(--amber-soft)',  fg: 'var(--amber)' },
  declined:    { bg: 'var(--coral-soft)',  fg: 'var(--coral)' },
  revision:    { bg: 'var(--accent-soft)', fg: 'var(--accent)' },
  resolved:    { bg: 'var(--green-soft)',  fg: 'var(--green)' },
  in_progress: { bg: 'var(--accent-soft)', fg: 'var(--accent)' },
  dismissed:   { bg: 'var(--bg-subtle)',   fg: 'var(--text-secondary)' },
  open:        { bg: 'var(--coral-soft)',  fg: 'var(--coral)' },
  paid:        { bg: 'var(--green-soft)',  fg: 'var(--green)' },
  cancelled:   { bg: 'var(--bg-subtle)',   fg: 'var(--text-secondary)' },
  expired:     { bg: 'var(--bg-subtle)',   fg: 'var(--text-tertiary)' },
};

const ACTIONABLE = new Set(['pending', 'open', 'revision']);

export function StatusBadge({ status }: StatusBadgeProps) {
  const key = status.toLowerCase();
  const c = COLORS[key] ?? { bg: 'var(--bg-subtle)', fg: 'var(--text-secondary)' };
  const pulse = ACTIONABLE.has(key);
  return (
    <span
      className={pulse ? 'status-pulse' : undefined}
      style={{
        display: 'inline-block',
        padding: '3px 10px',
        borderRadius: 9999,
        background: c.bg,
        color: c.fg,
        fontSize: 12,
        fontWeight: 600,
        textTransform: 'capitalize',
        letterSpacing: '0.02em',
      }}
    >
      {status.replace(/_/g, ' ')}
    </span>
  );
}
