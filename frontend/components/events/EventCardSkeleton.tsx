export function EventCardSkeleton() {
  return (
    <div
      className="animate-pulse"
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '14px 20px',
        borderBottom: '1px solid var(--border)',
      }}
    >
      {/* Thumbnail */}
      <div
        className="rounded-[10px]"
        style={{ width: 56, height: 56, flexShrink: 0, background: 'var(--bg-subtle)' }}
      />
      {/* Info */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div className="mb-2.5 h-4 w-1/2 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
        <div className="mb-1.5 h-3 w-2/3 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
        <div className="h-3 w-1/3 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
      </div>
    </div>
  );
}
