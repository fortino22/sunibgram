interface VerifiedBadgeProps {
  size?: number;
}

export function VerifiedBadge({ size = 16 }: VerifiedBadgeProps) {
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
      <circle cx="11" cy="11" r="11" fill="#1D9BF0" />
      <path
        stroke="white"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M6.5 11.5L9.5 14.5L15.5 8"
      />
    </svg>
  );
}
