'use client';

import { useState } from 'react';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import {
  SettingRow,
  SettingsPanel,
  SuccessBanner,
  Toggle,
} from '@/components/settings/SettingsPanel';
import { useUpdateUserSettings, useUserSettings } from '@/hooks/useUserSettings';

export default function PrivacySettingsPage() {
  const { data: settings, isLoading } = useUserSettings();
  const { mutate, isPending, error } = useUpdateUserSettings();
  const [success, setSuccess] = useState<string | null>(null);

  function handleTogglePrivate(nextValue: boolean) {
    setSuccess(null);
    mutate(
      { is_private: nextValue },
      {
        onSuccess: () =>
          setSuccess(nextValue ? 'Your account is now private.' : 'Your account is now public.'),
      }
    );
  }

  return (
    <SettingsPanel
      title="Privacy and safety"
      description="Control who can see your posts and follow you."
    >
      <div style={{ maxWidth: 640 }}>
        <SuccessBanner message={success} />
        <ErrorAlert message={error ? (error as Error).message : null} />

        {isLoading || !settings ? (
          <SkeletonRow />
        ) : (
          <SettingRow
            title="Private account"
            description="When enabled, only approved followers can see your posts. Existing followers won't be affected."
          >
            <Toggle
              checked={settings.is_private}
              onChange={handleTogglePrivate}
              disabled={isPending}
              ariaLabel="Toggle private account"
            />
          </SettingRow>
        )}
      </div>
    </SettingsPanel>
  );
}

function SkeletonRow() {
  return (
    <div style={{ display: 'flex', gap: 16, padding: '14px 0' }}>
      <div style={{ flex: 1 }}>
        <div style={{ width: 160, height: 14, background: 'var(--bg-subtle)', borderRadius: 4, marginBottom: 8 }} />
        <div style={{ width: '90%', height: 12, background: 'var(--bg-subtle)', borderRadius: 4 }} />
      </div>
      <div style={{ width: 44, height: 26, background: 'var(--bg-subtle)', borderRadius: 9999 }} />
    </div>
  );
}
