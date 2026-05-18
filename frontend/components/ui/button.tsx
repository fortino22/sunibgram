'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';

type ButtonVariant = 'default' | 'outline' | 'ghost';
type ButtonSize = 'default' | 'sm' | 'lg' | 'icon';

const variantClasses: Record<ButtonVariant, string> = {
  default: 'border-transparent text-white shadow-sm',
  outline: 'border-[color:var(--border)] bg-transparent text-[color:var(--text)]',
  ghost: 'border-transparent bg-transparent text-[color:var(--text)]',
};

const sizeClasses: Record<ButtonSize, string> = {
  default: 'h-10 px-4 py-2 text-sm',
  sm: 'h-9 px-3 text-sm',
  lg: 'h-11 px-5 text-sm',
  icon: 'h-10 w-10',
};

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'default', style, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          'inline-flex items-center justify-center gap-2 rounded-xl border font-medium transition-colors',
          'disabled:pointer-events-none disabled:opacity-50',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[color:var(--accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-[color:var(--bg-card)]',
          variantClasses[variant],
          sizeClasses[size],
          className
        )}
        style={{
          background:
            variant === 'default'
              ? 'var(--accent)'
              : variant === 'ghost'
                ? 'transparent'
                : 'var(--bg-card)',
          ...style,
        }}
        {...props}
      />
    );
  }
);

Button.displayName = 'Button';
