import { FEEDBACK_STATUS_LABEL, type FeedbackStatus } from '@/lib/services/feedback';

interface FeedbackStatusBadgeProps {
  status: FeedbackStatus;
}

const STATUS_DOT: Record<FeedbackStatus, string> = {
  open: 'var(--text-secondary)',
  planned: 'var(--accent)',
  in_progress: 'var(--amber)',
  resolved: 'var(--green)',
  rejected: 'var(--text-tertiary)',
};

export function FeedbackStatusBadge({ status }: FeedbackStatusBadgeProps) {
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        padding: '2px 0',
        background: 'transparent',
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
          background: STATUS_DOT[status],
        }}
      />
      {FEEDBACK_STATUS_LABEL[status]}
    </span>
  );
}
