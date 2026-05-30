'use client';

import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import { OtpInput } from '@/components/auth/OtpInput';
import { useAuthStore } from '@/stores/authStore';
import { useMyProfile } from '@/hooks/useMyProfile';
import {
  changePasswordWithOtp,
  sendVerification,
  verifyVerification,
} from '@/lib/services/auth';
import {
  PrimaryButton,
  SettingsField,
  SettingsPanel,
  SuccessBanner,
  settingsInputStyle,
} from '@/components/settings/SettingsPanel';

type Step = 'form' | 'otp' | 'done';

const passwordSchema = z
  .object({
    newPassword: z.string().min(8, 'Min 8 characters').max(64, 'Max 64 characters'),
    confirmPassword: z.string().min(1, 'Confirm your new password'),
  })
  .refine((v) => v.newPassword === v.confirmPassword, {
    path: ['confirmPassword'],
    message: 'Passwords do not match',
  });

type PasswordValues = z.infer<typeof passwordSchema>;

export default function PasswordSettingsPage() {
  const { data: profile, isLoading: profileLoading } = useMyProfile();
  const storeEmail = useAuthStore((s) => s.user?.email);
  const token = useAuthStore((s) => s.token);
  const email = profile?.email ?? storeEmail;
  const [step, setStep] = useState<Step>('form');
  const [code, setCode] = useState<string[]>(Array(6).fill(''));
  const [pendingPassword, setPendingPassword] = useState<PasswordValues | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<PasswordValues>({
    resolver: zodResolver(passwordSchema),
    defaultValues: { newPassword: '', confirmPassword: '' },
  });

  const sendCodeMutation = useMutation({
    mutationFn: (email: string) => sendVerification(email, 'CHANGE_PASSWORD'),
  });

  const verifyAndChange = useMutation({
    mutationFn: async ({
      email,
      otp,
      newPassword,
      confirmPassword,
    }: {
      email: string;
      otp: string;
      newPassword: string;
      confirmPassword: string;
    }) => {
      if (!token) throw new Error('Not authenticated.');
      await verifyVerification(email, otp, 'CHANGE_PASSWORD');
      return changePasswordWithOtp({ newPassword, confirmPassword }, token);
    },
  });

  function onSubmitPassword(values: PasswordValues) {
    if (!email) {
      setErrorMsg('Cannot resolve account email.');
      return;
    }
    setErrorMsg(null);
    setPendingPassword(values);
    sendCodeMutation.mutate(email, {
      onSuccess: () => setStep('otp'),
      onError: (err) => setErrorMsg((err as Error).message),
    });
  }

  function onSubmitOtp() {
    if (!email || !pendingPassword) return;
    const otp = code.join('');
    if (otp.length !== 6) {
      setErrorMsg('Enter the 6-digit code.');
      return;
    }
    setErrorMsg(null);
    verifyAndChange.mutate(
      {
        email,
        otp,
        newPassword: pendingPassword.newPassword,
        confirmPassword: pendingPassword.confirmPassword,
      },
      {
        onSuccess: () => {
          setStep('done');
          reset({ newPassword: '', confirmPassword: '' });
          setCode(Array(6).fill(''));
          setPendingPassword(null);
        },
        onError: (err) => setErrorMsg((err as Error).message),
      }
    );
  }

  function resendCode() {
    if (!email) return;
    setErrorMsg(null);
    sendCodeMutation.mutate(email);
  }

  function backToForm() {
    setStep('form');
    setCode(Array(6).fill(''));
    setErrorMsg(null);
  }

  function startOver() {
    setStep('form');
    setCode(Array(6).fill(''));
    setPendingPassword(null);
    setErrorMsg(null);
  }

  const isBusy = sendCodeMutation.isPending || verifyAndChange.isPending;

  return (
    <SettingsPanel
      title="Change password"
      description="We'll send a verification code to your email before applying the change."
    >
      <div style={{ maxWidth: 560 }}>
        <ErrorAlert message={errorMsg} />

        {!email && profileLoading ? (
          <SkeletonForm />
        ) : !email ? (
          <ErrorAlert message="Cannot resolve account email. Please re-login and try again." />
        ) : step === 'form' ? (
          <form onSubmit={handleSubmit(onSubmitPassword)}>
            <SettingsField
              label="Email"
              htmlFor="accountEmail"
              hint="The verification code is sent here."
            >
              <input
                id="accountEmail"
                type="email"
                value={email}
                readOnly
                disabled
                style={{ ...settingsInputStyle(false), cursor: 'not-allowed', opacity: 0.7 }}
              />
            </SettingsField>

            <SettingsField
              label="New password"
              htmlFor="newPassword"
              error={errors.newPassword?.message}
              hint="Use at least 8 characters."
            >
              <input
                id="newPassword"
                type="password"
                autoComplete="new-password"
                {...register('newPassword')}
                style={settingsInputStyle(!!errors.newPassword)}
              />
            </SettingsField>

            <SettingsField
              label="Confirm new password"
              htmlFor="confirmPassword"
              error={errors.confirmPassword?.message}
            >
              <input
                id="confirmPassword"
                type="password"
                autoComplete="new-password"
                {...register('confirmPassword')}
                style={settingsInputStyle(!!errors.confirmPassword)}
              />
            </SettingsField>

            <div style={{ marginTop: 24, display: 'flex', justifyContent: 'flex-end' }}>
              <PrimaryButton disabled={isBusy}>
                {sendCodeMutation.isPending ? 'Sending code…' : 'Continue'}
              </PrimaryButton>
            </div>
          </form>
        ) : step === 'otp' ? (
          <div>
            <p style={{ fontSize: 14, color: 'var(--text-secondary)', marginBottom: 18 }}>
              We sent a 6-digit code to{' '}
              <strong style={{ color: 'var(--text)' }}>{email}</strong>. Enter it below to
              confirm the password change.
            </p>

            <div style={{ marginBottom: 18 }}>
              <OtpInput value={code} onChange={setCode} disabled={verifyAndChange.isPending} />
            </div>

            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 12,
                marginTop: 8,
              }}
            >
              <button
                type="button"
                onClick={backToForm}
                disabled={isBusy}
                style={linkButtonStyle}
              >
                Back
              </button>
              <button
                type="button"
                onClick={resendCode}
                disabled={isBusy}
                style={linkButtonStyle}
              >
                {sendCodeMutation.isPending ? 'Resending…' : 'Resend code'}
              </button>
            </div>

            <div style={{ marginTop: 24, display: 'flex', justifyContent: 'flex-end' }}>
              <PrimaryButton
                type="button"
                onClick={onSubmitOtp}
                disabled={verifyAndChange.isPending || code.join('').length !== 6}
              >
                {verifyAndChange.isPending ? 'Verifying…' : 'Verify and change password'}
              </PrimaryButton>
            </div>
          </div>
        ) : (
          <div>
            <SuccessBanner message="Password changed successfully." />
            <PrimaryButton type="button" onClick={startOver}>
              Change again
            </PrimaryButton>
          </div>
        )}
      </div>
    </SettingsPanel>
  );
}

const linkButtonStyle: React.CSSProperties = {
  background: 'none',
  border: 'none',
  color: 'var(--accent)',
  fontSize: 14,
  fontWeight: 500,
  cursor: 'pointer',
  padding: 0,
  fontFamily: 'inherit',
};

function SkeletonForm() {
  return (
    <>
      {[0, 1].map((i) => (
        <div key={i} style={{ marginBottom: 18 }}>
          <div style={{ width: 110, height: 12, background: 'var(--bg-subtle)', borderRadius: 4, marginBottom: 8 }} />
          <div style={{ width: '100%', height: 42, background: 'var(--bg-subtle)', borderRadius: 8 }} />
        </div>
      ))}
    </>
  );
}
