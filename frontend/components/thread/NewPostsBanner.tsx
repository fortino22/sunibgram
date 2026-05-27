interface NewPostsBannerProps {
  count: number;
  onClick: () => void;
}

export function NewPostsBanner({ count, onClick }: NewPostsBannerProps) {
  const label = count >= 1000 ? '999+' : String(count);
  return (
    <button
      type="button"
      onClick={onClick}
      className="btn-ghost"
      style={{
        display: 'block',
        width: '100%',
        padding: '12px',
        border: 'none',
        borderBottom: '1px solid var(--border)',
        cursor: 'pointer',
        fontSize: 14,
        fontWeight: 600,
        color: 'var(--accent)',
        textAlign: 'center',
        fontFamily: 'inherit',
      }}
    >
      Show {label} posts
    </button>
  );
}
