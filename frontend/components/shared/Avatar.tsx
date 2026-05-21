import type { CSSProperties } from 'react';
import { AvatarFallback, AvatarImage, AvatarRoot } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';

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

function SilhouetteFallback() {
  return (
    <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" style={{ width: '100%', height: '100%' }}>
      <circle cx="50" cy="50" r="50" fill="#2e3135" />
      <circle cx="50" cy="38" r="18" fill="#c8cdd4" />
      <ellipse cx="50" cy="95" rx="32" ry="26" fill="#c8cdd4" />
    </svg>
  );
}

export function Avatar({ username, avatarUrl, size = 'md', className, style }: AvatarProps) {
  const resolvedSize = resolveSize(size);

  return (
    <AvatarRoot
      className={cn('shrink-0 overflow-hidden rounded-full', className)}
      style={{ width: resolvedSize, height: resolvedSize, ...style }}
    >
      {avatarUrl?.trim() && (
        <AvatarImage src={avatarUrl} alt={`${username} profile image`} />
      )}
      <AvatarFallback className="p-0">
        <SilhouetteFallback />
      </AvatarFallback>
    </AvatarRoot>
  );
}
