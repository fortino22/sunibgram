interface VerifiedBadgeProps {
  size?: number;
  /** 'brand' = blue verified badge (default). 'mono' = theme-aware black/white,
      used only in the settings nav so the icon stays readable on both themes. */
  variant?: 'brand' | 'mono';
}

export function VerifiedBadge({ size = 16, variant = 'brand' }: VerifiedBadgeProps) {
  const mono = variant === 'mono';
  // var() is only valid in CSS properties (style), not in SVG presentation
  // attributes — so the mono variant sets fill/stroke via style.
  const circleFill = mono ? 'var(--premium-badge-bg)' : '#1D9BF0';
  const checkStroke = mono ? 'var(--premium-badge-fg)' : 'white';
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 22 22"
      aria-label="Verified"
      role="img"
      fill="none"
      style={{ flexShrink: 0 }}
    >
      <circle cx="11" cy="11" r="11" fill={mono ? undefined : circleFill} style={mono ? { fill: circleFill } : undefined} />
      <path
        stroke={mono ? undefined : checkStroke}
        style={mono ? { stroke: checkStroke } : undefined}
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M6.5 11.5L9.5 14.5L15.5 8"
      />
    </svg>
  );
}
