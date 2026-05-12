'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { ArrowLeft, CheckCircle2, Eye, EyeOff } from 'lucide-react';
import { useState, type FormEventHandler, type ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';

import { OtpInput } from '@/components/auth/OtpInput';
import {
  forgotPasswordEmailSchema,
  forgotPasswordResetSchema,
  verificationCodeSchema,
  type ForgotPasswordEmailInput,
  type ForgotPasswordResetInput,
  type ForgotPasswordStep,
} from '@/components/auth/schemas';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import { forgotPassword, sendVerification, verifyVerification } from '@/lib/services/auth';

interface ForgotPasswordFlowProps {
  initialEmail?: string;
}

function FormField({
  label,
  error,
  children,
}: {
  label: string;
  error?: string;
  children: ReactNode;
}) {
  return (
    <div>
      <label
        style={{
          display: 'block',
          fontSize: 13,
          fontWeight: 500,
          color: 'var(--text-secondary)',
          marginBottom: 6,
        }}
      >
        {label}
      </label>
      {children}
      {error && <p style={{ fontSize: 12, color: 'var(--coral)', marginTop: 5 }}>{error}</p>}
    </div>
  );
}

const inputStyle = (hasError: boolean) => ({
  width: '100%',
  padding: '10px 14px',
  background: 'var(--bg-subtle)',
  border: `1.5px solid ${hasError ? 'var(--coral)' : 'var(--border)'}`,
  borderRadius: 10,
  fontSize: 14,
  color: 'var(--text)',
  outline: 'none',
  fontFamily: 'inherit',
  boxSizing: 'border-box' as const,
});

const toggleStyle = {
  position: 'absolute' as const,
  right: 12,
  top: '50%',
  transform: 'translateY(-50%)',
  background: 'none',
  border: 'none',
  cursor: 'pointer',
  color: 'var(--text-tertiary)',
  display: 'flex',
  alignItems: 'center',
  padding: 0,
};

function LoadingButton({
  disabled,
  loading,
  idleLabel,
  loadingLabel,
  onClick,
  type = 'button',
}: {
  disabled: boolean;
  loading: boolean;
  idleLabel: string;
  loadingLabel: string;
  onClick?: () => void;
  type?: 'button' | 'submit';
}) {
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      style={{
        width: '100%',
        padding: '11px 0',
        background: disabled ? 'var(--accent-soft)' : 'var(--accent)',
        border: 'none',
        borderRadius: 10,
        color: '#fff',
        fontWeight: 600,
        fontSize: 15,
        cursor: disabled ? 'not-allowed' : 'pointer',
        letterSpacing: '-0.01em',
        fontFamily: 'inherit',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 8,
      }}
    >
      {loading ? (
        <>
          <span
            style={{
              width: 14,
              height: 14,
              borderRadius: '50%',
              border: '2px solid rgba(255,255,255,0.3)',
              borderTopColor: '#fff',
              animation: 'spin 0.7s linear infinite',
              display: 'inline-block',
            }}
          />
          {loadingLabel}
        </>
      ) : (
        idleLabel
      )}
    </button>
  );
}

function createInitialOtp(): string[] {
  return Array(6).fill('');
}

export function ForgotPasswordFlow({
  initialEmail = '',
}: ForgotPasswordFlowProps) {
  const router = useRouter();
  const [step, setStep] = useState<ForgotPasswordStep>('email');
  const [email, setEmail] = useState(initialEmail);
  const [otp, setOtp] = useState<string[]>(createInitialOtp());
  const [serverError, setServerError] = useState('');
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const emailForm = useForm<ForgotPasswordEmailInput>({
    resolver: zodResolver(forgotPasswordEmailSchema),
    defaultValues: { email: initialEmail },
  });
  const resetForm = useForm<ForgotPasswordResetInput>({
    resolver: zodResolver(forgotPasswordResetSchema),
    defaultValues: {
      newPassword: '',
      confirmPassword: '',
    },
  });

  const sendCodeMutation = useMutation({
    mutationFn: (values: ForgotPasswordEmailInput) =>
      sendVerification(values.email, 'FORGOT_PASSWORD'),
  });
  const verifyCodeMutation = useMutation({
    mutationFn: (code: string) => verifyVerification(email, code, 'FORGOT_PASSWORD'),
  });
  const resetPasswordMutation = useMutation({
    mutationFn: (values: ForgotPasswordResetInput) =>
      forgotPassword({
        email,
        newPassword: values.newPassword,
        confirmPassword: values.confirmPassword,
      }),
  });

  const handleSendCode: FormEventHandler<HTMLFormElement> = emailForm.handleSubmit(
    async (values) => {
      setServerError('');

      try {
        await sendCodeMutation.mutateAsync(values);
        setEmail(values.email);
        setOtp(createInitialOtp());
        setStep('otp');
      } catch (error) {
        setServerError(
          error instanceof Error ? error.message : 'Failed to send verification code'
        );
      }
    }
  );

  const handleVerifyCode = async () => {
    const parsed = verificationCodeSchema.safeParse({ code: otp.join('') });

    if (!parsed.success) {
      setServerError(parsed.error.issues[0]?.message ?? 'Enter the 6-digit code');
      return;
    }

    setServerError('');

    try {
      await verifyCodeMutation.mutateAsync(parsed.data.code);
      resetForm.reset();
      setStep('reset');
    } catch (error) {
      setServerError(error instanceof Error ? error.message : 'Verification failed');
    }
  };

  const handleResetPassword: FormEventHandler<HTMLFormElement> = resetForm.handleSubmit(
    async (values) => {
      setServerError('');

      try {
        await resetPasswordMutation.mutateAsync(values);
        setStep('done');
      } catch (error) {
        setServerError(error instanceof Error ? error.message : 'Failed to reset password');
      }
    }
  );

  const handleResendCode = async () => {
    setServerError('');

    try {
      await sendCodeMutation.mutateAsync({ email });
      setOtp(createInitialOtp());
    } catch (error) {
      setServerError(error instanceof Error ? error.message : 'Failed to resend code');
    }
  };

  if (step === 'email') {
    return (
      <>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <h1
            style={{
              fontWeight: 600,
              fontSize: 28,
              color: 'var(--text)',
              letterSpacing: '-0.03em',
              marginBottom: 8,
            }}
          >
            Reset your password
          </h1>
          <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
            Enter your Binus email and we&apos;ll send a verification code.
          </p>
        </div>

        <ErrorAlert message={serverError} />

        <form onSubmit={handleSendCode} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <FormField label="Email" error={emailForm.formState.errors.email?.message}>
            <input
              {...emailForm.register('email')}
              type="email"
              placeholder="you@binus.ac.id"
              autoComplete="email"
              style={inputStyle(!!emailForm.formState.errors.email)}
            />
          </FormField>

          <LoadingButton
            type="submit"
            disabled={sendCodeMutation.isPending}
            loading={sendCodeMutation.isPending}
            idleLabel="Send Code"
            loadingLabel="Sending code…"
          />
        </form>

        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </>
    );
  }

  if (step === 'otp') {
    const isSendingCode = sendCodeMutation.isPending;
    const isVerifyingCode = verifyCodeMutation.isPending;

    return (
      <>
        <button
          type="button"
          onClick={() => {
            setServerError('');
            setStep('email');
          }}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            color: 'var(--text-tertiary)',
            fontSize: 14,
            margin: '0 auto 16px',
            fontFamily: 'inherit',
            padding: 0,
          }}
        >
          <ArrowLeft size={14} />
          Back
        </button>

        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <h1
            style={{
              fontWeight: 600,
              fontSize: 28,
              color: 'var(--text)',
              letterSpacing: '-0.03em',
              marginBottom: 8,
            }}
          >
            Check your email
          </h1>
          <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
            We sent a 6-digit code to{' '}
            <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{email}</span>
          </p>
        </div>

        <ErrorAlert message={serverError} />

        <div style={{ marginBottom: 24 }}>
          <OtpInput value={otp} onChange={setOtp} disabled={isSendingCode || isVerifyingCode} />
        </div>

        <LoadingButton
          disabled={isSendingCode || isVerifyingCode}
          loading={isVerifyingCode}
          idleLabel="Verify"
          loadingLabel="Verifying…"
          onClick={handleVerifyCode}
        />

        <div style={{ textAlign: 'center', marginTop: 16 }}>
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
            Didn&apos;t receive it?{' '}
          </span>
          <button
            type="button"
            onClick={handleResendCode}
            disabled={isSendingCode || isVerifyingCode}
            style={{
              background: 'none',
              border: 'none',
              cursor: isSendingCode || isVerifyingCode ? 'not-allowed' : 'pointer',
              color: 'var(--accent)',
              fontSize: 13,
              fontWeight: 500,
              fontFamily: 'inherit',
              padding: 0,
              opacity: isSendingCode || isVerifyingCode ? 0.7 : 1,
            }}
          >
            {isSendingCode ? 'Sending…' : 'Resend code'}
          </button>
        </div>

        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </>
    );
  }

  if (step === 'reset') {
    return (
      <>
        <button
          type="button"
          onClick={() => {
            setServerError('');
            setStep('otp');
          }}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            color: 'var(--text-tertiary)',
            fontSize: 14,
            margin: '0 auto 16px',
            fontFamily: 'inherit',
            padding: 0,
          }}
        >
          <ArrowLeft size={14} />
          Back
        </button>

        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <h1
            style={{
              fontWeight: 600,
              fontSize: 28,
              color: 'var(--text)',
              letterSpacing: '-0.03em',
              marginBottom: 8,
            }}
          >
            Create a new password
          </h1>
          <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
            Set a new password for{' '}
            <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{email}</span>
          </p>
        </div>

        <ErrorAlert message={serverError} />

        <form
          onSubmit={handleResetPassword}
          style={{ display: 'flex', flexDirection: 'column', gap: 16 }}
        >
          <FormField
            label="New Password"
            error={resetForm.formState.errors.newPassword?.message}
          >
            <div style={{ position: 'relative' }}>
              <input
                {...resetForm.register('newPassword')}
                type={showNewPassword ? 'text' : 'password'}
                placeholder="••••••••"
                autoComplete="new-password"
                style={{
                  ...inputStyle(!!resetForm.formState.errors.newPassword),
                  paddingRight: 42,
                }}
              />
              <button
                type="button"
                onClick={() => setShowNewPassword((value) => !value)}
                style={toggleStyle}
              >
                {showNewPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
          </FormField>

          <FormField
            label="Confirm Password"
            error={resetForm.formState.errors.confirmPassword?.message}
          >
            <div style={{ position: 'relative' }}>
              <input
                {...resetForm.register('confirmPassword')}
                type={showConfirmPassword ? 'text' : 'password'}
                placeholder="••••••••"
                autoComplete="new-password"
                style={{
                  ...inputStyle(!!resetForm.formState.errors.confirmPassword),
                  paddingRight: 42,
                }}
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword((value) => !value)}
                style={toggleStyle}
              >
                {showConfirmPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
          </FormField>

          <LoadingButton
            type="submit"
            disabled={resetPasswordMutation.isPending}
            loading={resetPasswordMutation.isPending}
            idleLabel="Submit"
            loadingLabel="Updating password…"
          />
        </form>

        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </>
    );
  }

  return (
    <>
      <div
        style={{
          textAlign: 'center',
          paddingTop: 12,
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 16 }}>
          <CheckCircle2 size={52} color="var(--accent)" strokeWidth={1.5} />
        </div>
        <h2
          style={{
            fontWeight: 600,
            fontSize: 22,
            color: 'var(--text)',
            letterSpacing: '-0.03em',
            marginBottom: 8,
          }}
        >
          Password updated
        </h2>
        <p style={{ fontSize: 14, color: 'var(--text-tertiary)', marginBottom: 28 }}>
          You can now sign in with your new password.
        </p>
        <LoadingButton
          disabled={false}
          loading={false}
          idleLabel="Back to sign in"
          loadingLabel="Back to sign in"
          onClick={() => router.replace('/login')}
        />
      </div>

      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </>
  );
}
