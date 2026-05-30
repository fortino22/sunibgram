'use client';

import { useState } from 'react';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import { SettingsPanel, SuccessBanner } from '@/components/settings/SettingsPanel';
import { useUpdateUserTheme, useUserSettings } from '@/hooks/useUserSettings';
import type { UserTheme } from '@/lib/services/userSettings';

const THEMES: { value: UserTheme; label: string; description: string }[] = [
  { value: 'black', label: 'Dark', description: 'The default dark theme.' },
  { value: 'white', label: 'Light', description: 'A bright, clean light theme.' },
];

export default function AppearanceSettingsPage() {
  const { data: settings, isLoading } = useUserSettings();
  const { mutate, isPending, error } = useUpdateUserTheme();
  const [success, setSuccess] = useState<string | null>(null);

  const current = settings?.user_theme ?? 'black';

  function handleSelect(next: UserTheme) {
    if (next === current) return;
    setSuccess(null);
    mutate(
      { user_theme: next },
      {
        onSuccess: () =>
          setSuccess(next === 'white' ? 'Light theme applied.' : 'Dark theme applied.'),
      }
    );
  }

  return (
    <SettingsPanel title="Appearance" description="Choose how Sunibgram looks to you.">
      <div style={{ maxWidth: 640 }}>
        <SuccessBanner message={success} />
        <ErrorAlert message={error ? (error as Error).message : null} />

        {isLoading || !settings ? (
          <SkeletonCards />
        ) : (
          <div style={{ display: 'flex', gap: 12 }}>
            {THEMES.map((t) => (
              <ThemeCard
                key={t.value}
                label={t.label}
                description={t.description}
                preview={t.value}
                selected={current === t.value}
                disabled={isPending}
                onClick={() => handleSelect(t.value)}
              />
            ))}
          </div>
        )}
      </div>
    </SettingsPanel>
  );
}

interface ThemeCardProps {
  label: string;
  description: string;
  preview: UserTheme;
  selected: boolean;
  disabled?: boolean;
  onClick: () => void;
}

function ThemeCard({ label, description, preview, selected, disabled, onClick }: ThemeCardProps) {
  // Hardcoded swatch colors are deliberate: this is a preview of each theme,
  // so it must show the target theme regardless of the active one.
  const swatchBg = preview === 'white' ? '#ffffff' : '#0a0a0c';
  const swatchText = preview === 'white' ? '#15171a' : '#f5f5f7';
  const swatchBar = preview === 'white' ? '#e2e4e8' : '#26282d';

  return (
    <button
      type="button"
      aria-pressed={selected}
      onClick={onClick}
      disabled={disabled}
      style={{
        flex: 1,
        textAlign: 'left',
        padding: 0,
        borderRadius: 14,
        overflow: 'hidden',
        cursor: disabled ? 'not-allowed' : 'pointer',
        background: 'var(--bg-card)',
        border: selected ? '2px solid var(--accent)' : '2px solid var(--border)',
        transition: 'border-color 0.12s ease, transform 0.12s ease',
        opacity: disabled && !selected ? 0.7 : 1,
      }}
    >
      <div style={{ padding: 14, background: swatchBg }}>
        <div style={{ height: 10, width: '55%', borderRadius: 4, background: swatchBar, marginBottom: 8 }} />
        <div style={{ height: 8, width: '80%', borderRadius: 4, background: swatchBar, marginBottom: 6 }} />
        <div style={{ height: 8, width: '40%', borderRadius: 4, background: swatchBar }} />
        <div style={{ marginTop: 12, fontSize: 12, fontWeight: 600, color: swatchText }}>Aa</div>
      </div>
      <div style={{ padding: '12px 14px' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontSize: 14, fontWeight: 700, color: 'var(--text)' }}>{label}</span>
          {selected && (
            <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--accent)' }}>SELECTED</span>
          )}
        </div>
        <p style={{ marginTop: 4, fontSize: 12.5, color: 'var(--text-secondary)', lineHeight: 1.5 }}>
          {description}
        </p>
      </div>
    </button>
  );
}

function SkeletonCards() {
  return (
    <div style={{ display: 'flex', gap: 12 }}>
      {[0, 1].map((i) => (
        <div
          key={i}
          style={{
            flex: 1,
            height: 180,
            borderRadius: 14,
            background: 'var(--bg-subtle)',
            border: '1px solid var(--border)',
          }}
        />
      ))}
    </div>
  );
}
