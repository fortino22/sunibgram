'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { ArrowLeft } from 'lucide-react';
import { Logo } from '@/components/ui/Logo';
import { useForm } from 'react-hook-form';

import { AuthSplitLayout } from '@/components/auth/AuthSplitLayout';
import { LoginForm } from '@/components/auth/LoginForm';
import { OtpInput } from '@/components/auth/OtpInput';
import {
  loginSchema,
  verificationCodeSchema,
  type LoginInput,
  type LoginStep,
} from '@/components/auth/schemas';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import { login, loginProd, sendLoginOtp } from '@/lib/services/auth';
import { useAuthStore } from '@/stores/authStore';

export default function LoginPage() {
  const router = useRouter();
  const setToken = useAuthStore((state) => state.setToken);
  const [showPassword, setShowPassword] = useState(false);
  const [serverError, setServerError] = useState('');
  const [step, setStep] = useState<LoginStep>('credentials');
  const [formData, setFormData] = useState<LoginInput | null>(null);
  const [otp, setOtp] = useState<string[]>(Array(6).fill(''));
  const {
    register,
    handleSubmit,
    getValues,
    formState: { errors, isSubmitting },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
  });

  const sendLoginOtpMutation = useMutation({
    mutationFn: (data: LoginInput) =>
      sendLoginOtp({
        email: data.email,
        password: data.password,
      }),
  });
  const loginMutation = useMutation({
    mutationFn: login,
  });

  const setSessionCookie = async (sid: string) => {
    const response = await fetch('/auth/set-session', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sid }),
    });
    if (!response.ok) throw new Error('Failed to create session');
  };

  const onSubmit = handleSubmit(async (data) => {
    setServerError('');

    try {
      const prodEmails = new Set(['admin@binus.ac.id', 'devin@binus.ac.id', 'octa@binus.ac.id', 'chermiese@binus.ac.id']);
      if (prodEmails.has(data.email)) {
        const res = await loginProd(data.email, data.password);
        setToken(res.access_token, res.sid);
        await setSessionCookie(res.sid);
        router.replace('/');
        return;
      }

      await sendLoginOtpMutation.mutateAsync(data);
      setFormData(data);
      setOtp(Array(6).fill(''));
      setStep('otp');
    } catch (err) {
      setServerError(err instanceof Error ? err.message : 'Login failed');
    }
  });

  const onVerifyOtp = async () => {
    if (!formData) return;

    const parsed = verificationCodeSchema.safeParse({ code: otp.join('') });

    if (!parsed.success) {
      setServerError(parsed.error.issues[0]?.message ?? 'Enter the 6-digit code');
      return;
    }

    setServerError('');

    try {
      const res = await loginMutation.mutateAsync({
        email: formData.email,
        password: formData.password,
        otp: parsed.data.code,
      });
      setToken(res.access_token, res.sid);
      await setSessionCookie(res.sid);
      router.replace('/');
    } catch (err) {
      setServerError(err instanceof Error ? err.message : 'Verification failed');
    }
  };

  const onResendOtp = async () => {
    if (!formData) {
      return;
    }

    setServerError('');

    try {
      await sendLoginOtpMutation.mutateAsync(formData);
      setOtp(Array(6).fill(''));
    } catch (error) {
      setServerError(error instanceof Error ? error.message : 'Failed to resend code');
    }
  };

  return (
    <AuthSplitLayout>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
          <Logo size={48} style={{ color: 'var(--text)' }} />
        </div>

        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          {step === 'otp' && (
            <button onClick={() => { setStep('credentials'); setServerError(''); }} className="btn-ghost" style={{ display: 'flex', alignItems: 'center', gap: 6, border: 'none', cursor: 'pointer', color: 'var(--text-tertiary)', fontSize: 14, margin: '0 auto 16px', fontFamily: 'inherit', padding: 0 }}>
              <ArrowLeft size={14} />
              Back
            </button>
          )}
          <h1 style={{ fontWeight: 600, fontSize: 28, color: 'var(--text)', letterSpacing: '-0.03em', marginBottom: 8 }}>{step === 'credentials' ? 'Sign in to Sunibgram' : 'Check your email'}</h1>
          <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
            {step === 'credentials' ? 'Campus discourse for Binus students' : <>We sent a 6-digit code to <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{formData?.email}</span></>}
          </p>
          {step === 'otp' && (
            <p style={{ fontSize: 13, color: 'var(--text-tertiary)', lineHeight: 1.5, marginTop: 6 }}>
              Don&apos;t see it? Please check your <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>spam</span> or junk folder.
            </p>
          )}
        </div>

        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 16, padding: '28px 24px' }}>
          {step === 'credentials' ? (
            <LoginForm
              register={register}
              errors={errors}
              serverError={serverError}
              isSubmitting={isSubmitting || sendLoginOtpMutation.isPending}
              showPassword={showPassword}
              onSubmit={onSubmit}
              onTogglePassword={() => setShowPassword((value) => !value)}
              onForgotPassword={() => {
                setServerError('');
                const email = getValues('email').trim();
                const href = email.length > 0
                  ? `/forgot-password?${new URLSearchParams({ email }).toString()}`
                  : '/forgot-password';

                router.push(href);
              }}
            />
          ) : (
            <>
              <ErrorAlert message={serverError} />
              <div style={{ marginBottom: 24 }}>
                <OtpInput
                  value={otp}
                  onChange={setOtp}
                  disabled={loginMutation.isPending || sendLoginOtpMutation.isPending}
                />
              </div>
              <button type="button" onClick={onVerifyOtp} disabled={loginMutation.isPending || sendLoginOtpMutation.isPending} style={{ width: '100%', padding: '11px 0', background: loginMutation.isPending || sendLoginOtpMutation.isPending ? 'var(--accent-soft)' : 'var(--accent)', border: 'none', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, cursor: loginMutation.isPending || sendLoginOtpMutation.isPending ? 'not-allowed' : 'pointer', letterSpacing: '-0.01em', fontFamily: 'inherit', transition: 'background 0.15s ease', marginTop: 4, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
                {loginMutation.isPending ? <><span style={{ width: 14, height: 14, borderRadius: '50%', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', animation: 'spin 0.7s linear infinite', display: 'inline-block' }} />Signing in…</> : 'Verify & Sign In'}
              </button>
              <div style={{ textAlign: 'center', marginTop: 16 }}>
                <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Didn&apos;t receive it? </span>
                <button onClick={onResendOtp} disabled={loginMutation.isPending || sendLoginOtpMutation.isPending} className="hover-dim" style={{ background: 'none', border: 'none', cursor: loginMutation.isPending || sendLoginOtpMutation.isPending ? 'not-allowed' : 'pointer', color: 'var(--accent)', fontSize: 13, fontWeight: 500, fontFamily: 'inherit', padding: 0, opacity: loginMutation.isPending || sendLoginOtpMutation.isPending ? 0.7 : 1 }}>{sendLoginOtpMutation.isPending ? 'Sending…' : 'Resend code'}</button>
              </div>
            </>
          )}
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 12, margin: '20px 0' }}>
          <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
          <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>or</span>
          <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
        </div>
        <div style={{ textAlign: 'center', fontSize: 14, color: 'var(--text-tertiary)' }}>
          Don&apos;t have an account? <Link href="/signup" style={{ color: 'var(--accent)', fontWeight: 500, textDecoration: 'none' }}>Create account</Link>
        </div>
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </AuthSplitLayout>
  );
}
