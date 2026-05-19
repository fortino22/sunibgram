export function CommunitySkeleton() {
  return (
    <div>
      {Array.from({ length: 6 }).map((_, i) => (
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
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: '50%',
              background: 'var(--bg-subtle)',
              flexShrink: 0,
            }}
          />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
            <div style={{ width: '45%', height: 14, borderRadius: 4, background: 'var(--bg-subtle)' }} />
            <div style={{ width: '30%', height: 12, borderRadius: 4, background: 'var(--bg-subtle)' }} />
          </div>
        </div>
      ))}
    </div>
  );
}
