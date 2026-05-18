'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';

export const Textarea = React.forwardRef<HTMLTextAreaElement, React.TextareaHTMLAttributes<HTMLTextAreaElement>>(
  ({ className, style, ...props }, ref) => (
    <textarea
      ref={ref}
      className={cn(
        'flex min-h-[120px] w-full rounded-xl border bg-transparent px-3 py-2 text-sm outline-none transition-colors',
        'placeholder:text-[color:var(--text-tertiary)] focus-visible:ring-2 focus-visible:ring-[color:var(--accent)]',
        className
      )}
      style={{
        borderColor: 'var(--border)',
        color: 'var(--text)',
        background: 'var(--bg-card)',
        ...style,
      }}
      {...props}
    />
  )
);

Textarea.displayName = 'Textarea';
