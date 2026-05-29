export function CollectionCardSkeleton() {
  return (
    <div
      style={{
        border: '1px solid var(--border)',
        borderRadius: 16,
        background: 'var(--bg-card)',
        padding: '16px 18px',
        display: 'flex',
        flexDirection: 'column',
        gap: 10,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <div
          className="animate-pulse"
          style={{ width: 36, height: 36, borderRadius: 10, background: 'var(--bg-hover)' }}
        />
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
          <div
            className="animate-pulse"
            style={{ width: '60%', height: 14, borderRadius: 6, background: 'var(--bg-hover)' }}
          />
          <div
            className="animate-pulse"
            style={{ width: '30%', height: 10, borderRadius: 6, background: 'var(--bg-hover)' }}
          />
        </div>
      </div>
      <div
        className="animate-pulse"
        style={{ width: '85%', height: 12, borderRadius: 6, background: 'var(--bg-hover)' }}
      />
    </div>
  );
}
