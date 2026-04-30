export default function BookmarksPage() {
  return (
    <div>
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'oklch(10.5% 0.005 260 / 0.92)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
          padding: '14px 20px',
        }}
      >
        <div style={{ fontWeight: 500, fontSize: 20, color: 'var(--text)', letterSpacing: '-0.02em' }}>
          Bookmarks
        </div>
      </div>
      <div
        style={{
          padding: '80px 24px',
          textAlign: 'center',
          color: 'var(--text-tertiary)',
          fontSize: 14,
        }}
      >
        Coming soon
      </div>
    </div>
  );
}
