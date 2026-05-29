import { Bug, Hammer } from 'lucide-react';
import { FEEDBACK_TYPE_LABEL, type FeedbackType } from '@/lib/services/feedback';

interface FeedbackTypeBadgeProps {
  type: FeedbackType;
}

export function FeedbackTypeBadge({ type }: FeedbackTypeBadgeProps) {
  const isBug = type === 'bug';
  const color = isBug ? 'var(--coral)' : 'var(--accent)';
  const Glyph = isBug ? Bug : Hammer;

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 5,
        padding: '2px 0',
        background: 'transparent',
        color,
        fontSize: 11,
        fontWeight: 600,
        letterSpacing: '-0.005em',
        lineHeight: 1,
      }}
    >
      <Glyph size={12} strokeWidth={2} />
      {FEEDBACK_TYPE_LABEL[type]}
    </span>
  );
}
