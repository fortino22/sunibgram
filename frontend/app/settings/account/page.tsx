'use client';

import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useUpdateProfile } from '@/hooks/useUpdateProfile';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import {
  PrimaryButton,
  SettingsField,
  SettingsPanel,
  SuccessBanner,
  settingsInputStyle,
} from '@/components/settings/SettingsPanel';

const schema = z.object({
  name: z.string().min(1, 'Name is required').max(50, 'Max 50 characters'),
  username: z
    .string()
    .min(3, 'Min 3 characters')
    .max(30, 'Max 30 characters')
    .regex(/^[a-zA-Z0-9._]+$/, 'Only letters, numbers, dots, underscores'),
});

type FormValues = z.infer<typeof schema>;

export default function AccountSettingsPage() {
  const { data: profile, isLoading } = useMyProfile();
  const { mutate, isPending, error } = useUpdateProfile();
  const [success, setSuccess] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isDirty },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', username: '' },
  });

  useEffect(() => {
    if (profile) {
      reset({ name: profile.name ?? '', username: profile.username ?? '' });
    }
  }, [profile, reset]);

  function onSubmit(values: FormValues) {
    setSuccess(null);
    mutate(values, {
      onSuccess: () => {
        setSuccess('Account updated.');
        reset(values);
      },
    });
  }

  return (
    <SettingsPanel
      title="Your account"
      description="See information about your account, edit display name and username."
    >
      {isLoading || !profile ? (
        <SkeletonForm />
      ) : (
        <form onSubmit={handleSubmit(onSubmit)} style={{ maxWidth: 560 }}>
          <SuccessBanner message={success} />
          <ErrorAlert message={error ? (error as Error).message : null} />

          <SettingsField label="Display name" htmlFor="name" error={errors.name?.message}>
            <input id="name" type="text" {...register('name')} style={settingsInputStyle(!!errors.name)} />
          </SettingsField>

          <SettingsField label="Username" htmlFor="username" error={errors.username?.message}>
            <input id="username" type="text" {...register('username')} style={settingsInputStyle(!!errors.username)} />
          </SettingsField>

          <SettingsField label="Email" hint="Email is verified and cannot be changed.">
            <input
              type="text"
              value={maskEmail(profile.email)}
              disabled
              style={{ ...settingsInputStyle(), opacity: 0.6, cursor: 'not-allowed' }}
            />
          </SettingsField>

          <div style={{ marginTop: 24, display: 'flex', justifyContent: 'flex-end' }}>
            <PrimaryButton disabled={isPending || !isDirty}>
              {isPending ? 'Saving…' : 'Save'}
            </PrimaryButton>
          </div>
        </form>
      )}
    </SettingsPanel>
  );
}

function maskEmail(email: string | null | undefined): string {
  if (!email) return '';
  const [local, domain] = email.split('@');
  if (!domain) return email;
  const visible = local.slice(0, 4);
  const hidden = '*'.repeat(Math.max(local.length - 4, 3));
  return `${visible}${hidden}@${domain}`;
}

function SkeletonForm() {
  return (
    <div style={{ maxWidth: 560 }}>
      {[0, 1, 2].map((i) => (
        <div key={i} style={{ marginBottom: 18 }}>
          <div style={{ width: 90, height: 12, background: 'var(--bg-subtle)', borderRadius: 4, marginBottom: 8 }} />
          <div style={{ width: '100%', height: 42, background: 'var(--bg-subtle)', borderRadius: 8 }} />
        </div>
      ))}
    </div>
  );
}
