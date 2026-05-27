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

const variantBackground: Record<ButtonVariant, string> = {
  default: 'var(--accent)',
  outline: 'var(--bg-card)',
  ghost: 'transparent',
};

const variantHoverBackground: Record<ButtonVariant, string> = {
  default: 'var(--accent-hover)',
  outline: 'var(--bg-hover)',
  ghost: 'var(--bg-hover)',
};

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'default', style, onMouseEnter, onMouseLeave, ...props }, ref) => {
    const [hovered, setHovered] = React.useState(false);
    const showHover = hovered && !props.disabled;

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
        onMouseEnter={(e) => {
          setHovered(true);
          onMouseEnter?.(e);
        }}
        onMouseLeave={(e) => {
          setHovered(false);
          onMouseLeave?.(e);
        }}
        style={{
          background: showHover ? variantHoverBackground[variant] : variantBackground[variant],
          ...style,
        }}
        {...props}
      />
    );
  }
);

Button.displayName = 'Button';
