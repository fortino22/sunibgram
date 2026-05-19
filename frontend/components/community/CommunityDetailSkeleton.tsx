export function CommunityDetailSkeleton() {
  return (
    <div>
      {/* Banner */}
      <div style={{ height: 140, background: 'var(--bg-subtle)' }} />

      {/* Avatar + header info */}
      <div style={{ padding: '0 20px 16px', position: 'relative' }}>
        <div
          style={{
            width: 72,
            height: 72,
            borderRadius: '50%',
            background: 'var(--bg-card)',
            border: '3px solid var(--bg)',
            marginTop: -36,
            marginBottom: 12,
          }}
        />
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{ width: '50%', height: 18, borderRadius: 4, background: 'var(--bg-subtle)' }} />
          <div style={{ width: '30%', height: 13, borderRadius: 4, background: 'var(--bg-subtle)' }} />
          <div style={{ width: '70%', height: 13, borderRadius: 4, background: 'var(--bg-subtle)' }} />
        </div>
      </div>

      {/* Tab bar */}
      <div
        style={{
          display: 'flex',
          borderBottom: '1px solid var(--border)',
          padding: '0 20px',
          gap: 24,
        }}
      >
        {Array.from({ length: 3 }).map((_, i) => (
          <div
            key={i}
            style={{ width: 50, height: 36, borderRadius: 4, background: 'var(--bg-subtle)' }}
          />
        ))}
      </div>

      {/* Thread skeletons */}
      {Array.from({ length: 4 }).map((_, i) => (
        <div
          key={i}
          style={{
            display: 'flex',
            gap: 12,
            padding: '14px 20px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: '50%',
              background: 'var(--bg-subtle)',
              flexShrink: 0,
            }}
          />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div style={{ width: '35%', height: 13, borderRadius: 4, background: 'var(--bg-subtle)' }} />
            <div style={{ width: '100%', height: 13, borderRadius: 4, background: 'var(--bg-subtle)' }} />
            <div style={{ width: '80%', height: 13, borderRadius: 4, background: 'var(--bg-subtle)' }} />
          </div>
        </div>
      ))}
    </div>
  );
}
