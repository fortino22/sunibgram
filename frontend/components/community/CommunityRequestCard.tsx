import { Avatar } from '@/components/shared/Avatar';
import type { CommunityRequestData } from '@/lib/services/communities';
import { timeAgo } from '@/lib/utils';

interface CommunityRequestCardProps {
  request: CommunityRequestData;
  onApprove: (userId: number) => void | Promise<void>;
  onDecline: (userId: number) => void | Promise<void>;
  approvePending: boolean;
  declinePending: boolean;
  showDivider?: boolean;
}

export function CommunityRequestCard({
  request,
  onApprove,
  onDecline,
  approvePending,
  declinePending,
  showDivider = true,
}: CommunityRequestCardProps) {
  const btnBase: React.CSSProperties = {
    padding: '6px 14px',
    borderRadius: 999,
    fontSize: 13,
    fontWeight: 600,
    cursor: 'pointer',
    border: 'none',
  };

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '12px 20px',
        borderBottom: showDivider ? '1px solid var(--border)' : 'none',
      }}
    >
      <Avatar username={request.username} avatarUrl={request.avatar_url || null} size="md" />

      <div style={{ flex: 1 }}>
        <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)' }}>
          @{request.username}
        </div>
        <div style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 2 }}>
          Requested {timeAgo(request.created_at) || request.created_at}
        </div>
      </div>

      <div style={{ display: 'flex', gap: 8 }}>
        <button
          disabled={declinePending}
          onClick={() => onDecline(request.user_id)}
          style={{
            ...btnBase,
            background: 'transparent',
            border: '1px solid var(--border)',
            color: 'var(--text)',
            opacity: declinePending ? 0.6 : 1,
          }}
        >
          Decline
        </button>
        <button
          disabled={approvePending}
          onClick={() => onApprove(request.user_id)}
          style={{
            ...btnBase,
            background: 'var(--text)',
            color: 'var(--bg)',
            opacity: approvePending ? 0.6 : 1,
          }}
        >
          Approve
        </button>
      </div>
    </div>
  );
}
