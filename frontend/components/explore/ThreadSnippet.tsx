interface ThreadSnippetProps {
  content: string;
  showDivider?: boolean;
}

export function ThreadSnippet({ content, showDivider = true }: ThreadSnippetProps) {
  return (
    <div
      style={{
        padding: '14px 20px',
        borderBottom: showDivider ? '1px solid var(--border)' : 'none',
      }}
    >
      <p
        style={{
          fontSize: 15,
          color: 'var(--text)',
          lineHeight: 1.5,
          margin: 0,
          display: '-webkit-box',
          WebkitLineClamp: 3,
          WebkitBoxOrient: 'vertical',
          overflow: 'hidden',
        }}
      >
        {content}
      </p>
    </div>
  );
}
