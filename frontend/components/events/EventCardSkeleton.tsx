export function EventCardSkeleton() {
  return (
    <div
      className="animate-pulse border-b px-5 py-4"
      style={{ borderColor: 'var(--border)', background: 'var(--bg-card)' }}
    >
      <div
        className="mb-4 h-44 rounded-3xl"
        style={{ background: 'var(--bg-subtle)' }}
      />
      <div className="mb-3 flex gap-2">
        <div className="h-6 w-24 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
        <div className="h-6 w-20 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
      </div>
      <div className="mb-3 h-7 w-2/3 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
      <div className="mb-2 h-4 w-full rounded-full" style={{ background: 'var(--bg-subtle)' }} />
      <div className="h-4 w-5/6 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
    </div>
  );
}
