export function ErrorAlert({ message }: { message?: string | null }) {
  if (!message) return null;
  return (
    <div
      style={{
        padding: '10px 14px',
        background: 'var(--coral-soft)',
        border: '1px solid oklch(64% 0.18 25 / 0.3)',
        borderRadius: 10,
        fontSize: 13,
        color: 'var(--coral)',
        marginBottom: 20,
      }}
    >
      {message}
    </div>
  );
}
