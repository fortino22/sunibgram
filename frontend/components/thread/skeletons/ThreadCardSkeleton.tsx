export function ThreadCardSkeleton() {
  return (
    <div className="animate-pulse border-b p-4" style={{ borderColor: 'var(--border)', background: 'var(--bg-card)' }}>
      <div className="flex gap-3">
        <div className="h-10 w-10 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
        <div className="flex-1 space-y-2">
          <div className="h-3 w-40 rounded" style={{ background: 'var(--bg-subtle)' }} />
          <div className="h-3 w-full rounded" style={{ background: 'var(--bg-subtle)' }} />
          <div className="h-3 w-3/4 rounded" style={{ background: 'var(--bg-subtle)' }} />
        </div>
      </div>
    </div>
  );
}
