import type { CSSProperties } from 'react';
import { AvatarFallback, AvatarImage, AvatarRoot } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';

const DEFAULT_AVATAR_SRC = '/assets/sunibgram/default-profile.jpg';

const SIZE_MAP = {
  sm: 32,
  md: 40,
  lg: 84,
} as const;

interface AvatarProps {
  username: string;
  avatarUrl?: string | null;
  size?: keyof typeof SIZE_MAP | number;
  className?: string;
  style?: CSSProperties;
}

function resolveSize(size: AvatarProps['size']): number {
  return typeof size === 'number' ? size : SIZE_MAP[size ?? 'md'];
}

export function Avatar({ username, avatarUrl, size = 'md', className, style }: AvatarProps) {
  const resolvedSize = resolveSize(size);

  return (
    <AvatarRoot
      className={cn('shrink-0', className)}
      style={{ width: resolvedSize, height: resolvedSize, ...style }}
    >
      <AvatarImage src={avatarUrl ?? DEFAULT_AVATAR_SRC} alt={`${username} profile image`} />
      <AvatarFallback
        className="text-xs font-semibold"
        style={{
          background: 'var(--bg-subtle)',
          color: 'var(--text-secondary)',
        }}
      >
        {username.charAt(0).toUpperCase()}
      </AvatarFallback>
    </AvatarRoot>
  );
}
