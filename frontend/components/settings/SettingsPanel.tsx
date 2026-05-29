'use client';

import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import type { ReactNode } from 'react';

interface SettingsPanelProps {
  title: string;
  description?: string;
  children: ReactNode;
}

export function SettingsPanel({ title, description, children }: SettingsPanelProps) {
  return (
    <div>
      <header
        style={{
          padding: '14px 20px',
          borderBottom: '1px solid var(--border)',
          position: 'sticky',
          top: 0,
          background: 'var(--bg)',
          backdropFilter: 'blur(8px)',
          zIndex: 10,
          display: 'flex',
          alignItems: 'center',
          gap: 14,
        }}
      >
        <Link
          href="/settings"
          aria-label="Back to settings"
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: 34,
            height: 34,
            borderRadius: '50%',
            color: 'var(--text)',
          }}
          onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-hover)'; }}
          onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
        >
          <ArrowLeft size={18} />
        </Link>
        <div style={{ flex: 1, minWidth: 0 }}>
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, letterSpacing: '-0.02em', color: 'var(--text)' }}>
            {title}
          </h2>
          {description && (
            <p style={{ margin: '2px 0 0', fontSize: 12, color: 'var(--text-tertiary)' }}>
              {description}
            </p>
          )}
        </div>
      </header>
      <div style={{ padding: '20px' }}>{children}</div>
    </div>
  );
}

interface SettingsFieldProps {
  label: string;
  htmlFor?: string;
  error?: string;
  hint?: string;
  children: ReactNode;
}

export function SettingsField({ label, htmlFor, error, hint, children }: SettingsFieldProps) {
  return (
    <div style={{ marginBottom: 18 }}>
      <label
        htmlFor={htmlFor}
        style={{ display: 'block', fontSize: 13, color: 'var(--text-tertiary)', marginBottom: 6, fontWeight: 500 }}
      >
        {label}
      </label>
      {children}
      {hint && !error && (
        <div style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 4 }}>{hint}</div>
      )}
      {error && (
        <div style={{ fontSize: 12, color: 'var(--coral)', marginTop: 4 }}>{error}</div>
      )}
    </div>
  );
}

export const settingsInputStyle = (hasError = false): React.CSSProperties => ({
  width: '100%',
  background: 'var(--bg-subtle)',
  border: `1px solid ${hasError ? 'var(--coral)' : 'var(--border)'}`,
  borderRadius: 8,
  padding: '10px 12px',
  color: 'var(--text)',
  fontSize: 15,
  fontFamily: 'inherit',
  outline: 'none',
});

export function PrimaryButton({
  children,
  disabled,
  type = 'submit',
  onClick,
}: {
  children: ReactNode;
  disabled?: boolean;
  type?: 'submit' | 'button';
  onClick?: () => void;
}) {
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      style={{
        background: 'var(--text)',
        color: 'var(--bg)',
        border: 'none',
        borderRadius: 10,
        padding: '11px 22px',
        fontWeight: 700,
        fontSize: 15,
        fontFamily: 'inherit',
        cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.6 : 1,
      }}
    >
      {children}
    </button>
  );
}

export function SuccessBanner({ message }: { message?: string | null }) {
  if (!message) return null;
  return (
    <div
      style={{
        padding: '10px 14px',
        background: 'oklch(70% 0.15 145 / 0.12)',
        border: '1px solid oklch(70% 0.15 145 / 0.3)',
        borderRadius: 10,
        fontSize: 13,
        color: 'oklch(78% 0.15 145)',
        marginBottom: 16,
      }}
    >
      {message}
    </div>
  );
}

export function Toggle({
  checked,
  onChange,
  disabled,
  ariaLabel,
}: {
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
  ariaLabel?: string;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={ariaLabel}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      style={{
        position: 'relative',
        width: 44,
        height: 26,
        borderRadius: 9999,
        background: checked ? 'var(--accent)' : 'var(--border)',
        border: 'none',
        cursor: disabled ? 'not-allowed' : 'pointer',
        transition: 'background 0.15s ease',
        opacity: disabled ? 0.6 : 1,
        flexShrink: 0,
      }}
    >
      <span
        style={{
          position: 'absolute',
          top: 3,
          left: checked ? 21 : 3,
          width: 20,
          height: 20,
          borderRadius: '50%',
          background: '#fff',
          transition: 'left 0.15s ease',
          boxShadow: '0 1px 3px rgba(0,0,0,0.3)',
        }}
      />
    </button>
  );
}

export function SettingRow({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 16,
        padding: '14px 0',
        borderBottom: '1px solid var(--border)',
      }}
    >
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 15, fontWeight: 600, color: 'var(--text)' }}>{title}</div>
        {description && (
          <div style={{ fontSize: 13, color: 'var(--text-tertiary)', marginTop: 2 }}>
            {description}
          </div>
        )}
      </div>
      {children}
    </div>
  );
}
