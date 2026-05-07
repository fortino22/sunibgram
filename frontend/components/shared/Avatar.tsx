import { AvatarRoot, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';

const SIZE_MAP = {
  sm: { px: 32, font: 13 },
  md: { px: 40, font: 15 },
  lg: { px: 84, font: 30 },
} as const;

interface AvatarProps {
  username: string;
  avatarUrl?: string | null;
  size?: keyof typeof SIZE_MAP;
  className?: string;
  style?: React.CSSProperties;
}

export function Avatar({ username, avatarUrl, size = 'md', className, style }: AvatarProps) {
  const { px, font } = SIZE_MAP[size];
  const initial = username.charAt(0).toUpperCase();

  return (
    <AvatarRoot
      className={cn('shrink-0', className)}
      style={{ width: px, height: px, ...style }}
    >
      {avatarUrl && <AvatarImage src={avatarUrl} alt={username} />}
      <AvatarFallback
        style={{
          background: 'oklch(58% 0.23 270)',
          color: '#fff',
          fontWeight: 700,
          fontSize: font,
          letterSpacing: '-0.02em',
        }}
      >
        {initial}
      </AvatarFallback>
    </AvatarRoot>
  );
}
