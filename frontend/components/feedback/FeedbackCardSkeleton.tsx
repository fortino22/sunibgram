interface FeedbackCardSkeletonProps {
  view?: 'grid' | 'list';
}

export function FeedbackCardSkeleton({ view = 'grid' }: FeedbackCardSkeletonProps) {
  const isList = view === 'list';
  return (
    <div
      style={{
        padding: isList ? '14px 16px' : '16px',
        border: '1px solid var(--border)',
        borderRadius: isList ? 12 : 16,
        background: 'var(--bg-card)',
        display: 'flex',
        flexDirection: 'column',
        gap: isList ? 8 : 12,
        minHeight: isList ? 'auto' : 220,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <div
          style={{
            width: 36,
            height: 36,
            borderRadius: '50%',
            background: 'var(--bg-hover)',
            animation: 'pulse 1.4s ease-in-out infinite',
          }}
        />
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6, flex: 1 }}>
          <div
            style={{
              width: '60%',
              height: 11,
              borderRadius: 4,
              background: 'var(--bg-hover)',
              animation: 'pulse 1.4s ease-in-out infinite',
            }}
          />
          <div
            style={{
              width: '40%',
              height: 10,
              borderRadius: 4,
              background: 'var(--bg-hover)',
              animation: 'pulse 1.4s ease-in-out infinite',
            }}
          />
        </div>
      </div>

      <div
        style={{
          width: '80%',
          height: 14,
          borderRadius: 4,
          background: 'var(--bg-hover)',
          animation: 'pulse 1.4s ease-in-out infinite',
        }}
      />
      <div
        style={{
          width: '100%',
          height: 10,
          borderRadius: 4,
          background: 'var(--bg-hover)',
          animation: 'pulse 1.4s ease-in-out infinite',
        }}
      />
      <div
        style={{
          width: '70%',
          height: 10,
          borderRadius: 4,
          background: 'var(--bg-hover)',
          animation: 'pulse 1.4s ease-in-out infinite',
        }}
      />

      <div style={{ marginTop: isList ? 4 : 'auto', display: 'flex', gap: 6 }}>
        <div
          style={{
            width: 56,
            height: 28,
            borderRadius: 9999,
            background: 'var(--bg-hover)',
            animation: 'pulse 1.4s ease-in-out infinite',
          }}
        />
        <div
          style={{
            width: 56,
            height: 28,
            borderRadius: 9999,
            background: 'var(--bg-hover)',
            animation: 'pulse 1.4s ease-in-out infinite',
          }}
        />
      </div>
    </div>
  );
}
