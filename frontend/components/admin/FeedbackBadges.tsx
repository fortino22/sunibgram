import { Bug, Hammer } from 'lucide-react';

const TYPE_LABEL: Record<string, string> = {
  bug: 'Bug',
  feature_request: 'Feature',
};

export function AdminFeedbackTypeBadge({ type }: { type: string }) {
  const isBug = type === 'bug';
  const color = isBug ? 'var(--coral)' : 'var(--accent)';
  const Glyph = isBug ? Bug : Hammer;

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 5,
        color,
        fontSize: 11,
        fontWeight: 600,
        letterSpacing: '-0.005em',
        lineHeight: 1,
      }}
    >
      <Glyph size={12} strokeWidth={2} />
      {TYPE_LABEL[type] ?? type}
    </span>
  );
}

const STATUS_DOT: Record<string, string> = {
  open: 'var(--text-secondary)',
  planned: 'var(--accent)',
  in_progress: 'var(--amber)',
  resolved: 'var(--green)',
  rejected: 'var(--text-tertiary)',
  dismissed: 'var(--text-tertiary)',
};

function statusLabel(status: string): string {
  return status
    .split('_')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

export function AdminFeedbackStatusBadge({ status }: { status: string }) {
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        padding: '3px 9px',
        borderRadius: 9999,
        background: 'var(--bg-subtle)',
        border: '1px solid var(--border)',
        color: 'var(--text-secondary)',
        fontSize: 11,
        fontWeight: 600,
        letterSpacing: '-0.005em',
        lineHeight: 1,
      }}
    >
      <span
        aria-hidden
        style={{
          width: 6,
          height: 6,
          borderRadius: '50%',
          background: STATUS_DOT[status] ?? 'var(--text-tertiary)',
        }}
      />
      {statusLabel(status)}
    </span>
  );
}
