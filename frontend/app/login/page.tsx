'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { Logo } from '@/components/ui/Logo';
import { useForm } from 'react-hook-form';

import { AuthSplitLayout } from '@/components/auth/AuthSplitLayout';
import { LoginForm } from '@/components/auth/LoginForm';
import { loginSchema, type LoginInput } from '@/components/auth/schemas';
import { login } from '@/lib/services/auth';
import { useAuthStore } from '@/stores/authStore';

export default function LoginPage() {
  const router = useRouter();
  const setToken = useAuthStore((state) => state.setToken);
  const [showPassword, setShowPassword] = useState(false);
  const [serverError, setServerError] = useState('');
  const {
    register,
    handleSubmit,
    getValues,
    formState: { errors, isSubmitting },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
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
      const res = await loginMutation.mutateAsync(data);
      setToken(res.access_token, res.sid);
      await setSessionCookie(res.sid);
      router.replace('/');
    } catch (err) {
      setServerError(err instanceof Error ? err.message : 'Login failed');
    }
  });

  return (
    <AuthSplitLayout>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
          <Logo size={48} style={{ color: 'var(--text)' }} />
        </div>

        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <h1 style={{ fontWeight: 600, fontSize: 28, color: 'var(--text)', letterSpacing: '-0.03em', marginBottom: 8 }}>Sign in to Sunibgram</h1>
          <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
            Campus discourse for Binus students
          </p>
        </div>

        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 16, padding: '28px 24px' }}>
          <LoginForm
            register={register}
            errors={errors}
            serverError={serverError}
            isSubmitting={isSubmitting || loginMutation.isPending}
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
