'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';

export function Badge({
  className,
  style,
  ...props
}: React.HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-medium',
        className
      )}
      style={{
        borderColor: 'var(--border)',
        color: 'var(--text-secondary)',
        background: 'var(--bg-subtle)',
        ...style,
      }}
      {...props}
    />
  );
}
