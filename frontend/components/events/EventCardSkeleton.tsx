export function EventCardSkeleton() {
  return (
    <div
      className="animate-pulse"
      style={{
        display: 'flex',
        flexDirection: 'column',
        borderRadius: 16,
        border: '1px solid var(--border)',
        background: 'var(--bg-secondary)',
        overflow: 'hidden',
      }}
    >
      <div
        style={{
          width: '100%',
          aspectRatio: '16 / 10',
          background: 'var(--bg-subtle)',
        }}
      />
      <div style={{ padding: '14px 14px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
        <div className="h-3 w-1/2 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
        <div className="h-4 w-3/4 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
        <div className="h-3 w-2/5 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
      </div>
    </div>
  );
}
