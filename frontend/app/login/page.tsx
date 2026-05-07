'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowLeft } from 'lucide-react';
import Image from 'next/image';
import { useForm } from 'react-hook-form';

import { LoginForm } from '@/components/auth/LoginForm';
import { OtpInput } from '@/components/auth/OtpInput';
import { loginSchema, type LoginInput, type LoginStep } from '@/components/auth/schemas';
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
  const [loading, setLoading] = useState(false);
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<LoginInput>({ resolver: zodResolver(loginSchema) });

  const setSessionCookie = async (sid: string) => {
    const response = await fetch('/api/auth/set-session', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sid }),
    });
    if (!response.ok) throw new Error('Failed to create session');
  };
  const onSubmit = handleSubmit(async (data) => {
    setServerError('');
    try {
      if (data.email === 'admin@binus.ac.id') {
        const res = await loginProd(data.email, data.password);
        setToken(res.access_token, res.sid);
        await setSessionCookie(res.sid);
        router.replace('/');
        return;
      }
      await sendLoginOtp(data.email);
      setFormData(data);
      setOtp(Array(6).fill(''));
      setStep('otp');
    } catch (err) {
      setServerError(err instanceof Error ? err.message : 'Login failed');
    }
  });
  const onVerifyOtp = async () => {
    if (!formData) return;
    const code = otp.join('');
    if (code.length < 6) return void setServerError('Enter the 6-digit code');
    setServerError('');
    setLoading(true);
    try {
      const res = await login({ email: formData.email, password: formData.password, otp: code });
      setToken(res.access_token, res.sid);
      await setSessionCookie(res.sid);
      router.replace('/');
    } catch (err) {
      setServerError(err instanceof Error ? err.message : 'Verification failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '24px 16px' }}>
      <div style={{ width: '100%', maxWidth: 400 }}>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
          <Image src="/assets/sunibgram/logo.png" alt="Sunibgram" width={48} height={48} />
        </div>

        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          {step === 'otp' && (
            <button onClick={() => { setStep('credentials'); setServerError(''); }} style={{ display: 'flex', alignItems: 'center', gap: 6, background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-tertiary)', fontSize: 14, margin: '0 auto 16px', fontFamily: 'inherit', padding: 0 }}>
              <ArrowLeft size={14} />
              Back
            </button>
          )}
          <h1 style={{ fontWeight: 600, fontSize: 28, color: 'var(--text)', letterSpacing: '-0.03em', marginBottom: 8 }}>{step === 'credentials' ? 'Sign in to Binusgram' : 'Check your email'}</h1>
          <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
            {step === 'credentials' ? 'Campus discourse for Binus students' : <>We sent a 6-digit code to <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{formData?.email}</span></>}
          </p>
        </div>

        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 16, padding: '28px 24px' }}>
          {step === 'credentials' ? (
            <LoginForm register={register} errors={errors} serverError={serverError} isSubmitting={isSubmitting} showPassword={showPassword} onSubmit={onSubmit} onTogglePassword={() => setShowPassword((value) => !value)} />
          ) : (
            <>
              <ErrorAlert message={serverError} />
              <div style={{ marginBottom: 24 }}><OtpInput value={otp} onChange={setOtp} /></div>
              <button type="button" onClick={onVerifyOtp} disabled={loading} style={{ width: '100%', padding: '11px 0', background: loading ? 'var(--accent-soft)' : 'var(--accent)', border: 'none', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, cursor: loading ? 'not-allowed' : 'pointer', letterSpacing: '-0.01em', fontFamily: 'inherit', transition: 'background 0.15s ease', marginTop: 4, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
                {loading ? <><span style={{ width: 14, height: 14, borderRadius: '50%', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', animation: 'spin 0.7s linear infinite', display: 'inline-block' }} />Signing in…</> : 'Verify & Sign In'}
              </button>
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
      </div>
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}
