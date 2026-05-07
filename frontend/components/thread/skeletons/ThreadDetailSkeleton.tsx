export function ThreadDetailSkeleton() {
  return (
    <main
      className="mx-auto min-h-screen max-w-[600px]"
      style={{
        background: 'var(--bg-card)',
        borderLeft: '1px solid var(--border)',
        borderRight: '1px solid var(--border)',
      }}
    >
      <div
        className="sticky top-0 z-10 border-b px-4 py-3"
        style={{
          background: 'var(--bg-card)',
          borderColor: 'var(--border)',
        }}
      >
        <div className="h-[22px] w-[120px] rounded" style={{ background: 'var(--bg-subtle)' }} />
      </div>
      {Array.from({ length: 4 }, (_, index) => (
        <div key={index} className="flex gap-3 border-b p-4" style={{ borderColor: 'var(--border)' }}>
          <div className="h-11 w-11 rounded-full" style={{ background: 'var(--bg-subtle)' }} />
          <div className="flex-1 space-y-2">
            <div className="h-3 w-[45%] rounded" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-3 w-full rounded" style={{ background: 'var(--bg-subtle)' }} />
            <div className="h-3 w-[72%] rounded" style={{ background: 'var(--bg-subtle)' }} />
          </div>
        </div>
      ))}
    </main>
  );
}
