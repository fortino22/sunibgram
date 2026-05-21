interface NotificationSkeletonProps {
  count?: number;
}

function SkeletonPulse({ width, height, style }: { width?: number | string; height?: number | string; style?: React.CSSProperties }) {
  return (
    <div
      style={{
        width,
        height,
        borderRadius: 6,
        background: 'var(--bg-subtle)',
        animation: 'skeleton-pulse 1.4s ease-in-out infinite',
        ...style,
      }}
    />
  );
}

export function NotificationSkeleton({ count = 6 }: NotificationSkeletonProps) {
  return (
    <>
      <style>{`
        @keyframes skeleton-pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.4; }
        }
      `}</style>
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            padding: '14px 20px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          {/* Avatar circle */}
          <SkeletonPulse width={40} height={40} style={{ borderRadius: '50%', flexShrink: 0 }} />

          {/* Text lines */}
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 8 }}>
            <SkeletonPulse width="60%" height={13} />
            <SkeletonPulse width="35%" height={11} />
          </div>
        </div>
      ))}
    </>
  );
}
