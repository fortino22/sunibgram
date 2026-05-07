interface EmptyStateProps {
  title: string;
  subtitle: string;
}

export function EmptyState({ title, subtitle }: EmptyStateProps) {
  return (
    <div
      style={{
        padding: '80px 24px',
        textAlign: 'center',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 10,
      }}
    >
      <div
        style={{
          fontSize: 20,
          fontWeight: 600,
          color: 'var(--text)',
          letterSpacing: '-0.02em',
        }}
      >
        {title}
      </div>
      <div
        style={{
          fontSize: 14,
          color: 'var(--text-tertiary)',
          maxWidth: 280,
          lineHeight: 1.6,
        }}
      >
        {subtitle}
      </div>
    </div>
  );
}
