'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowLeft, CheckCircle2 } from 'lucide-react';
import Image from 'next/image';
import { useForm } from 'react-hook-form';

import { OtpInput } from '@/components/auth/OtpInput';
import { RegisterForm } from '@/components/auth/RegisterForm';
import { signupSchema, type SignupInput, type SignupStep } from '@/components/auth/schemas';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import { register as registerAccount, sendVerification, verifyVerification } from '@/lib/services/auth';

export default function SignupPage() {
  const router = useRouter();
  const [step, setStep] = useState<SignupStep>(1);
  const [formData, setFormData] = useState<SignupInput | null>(null);
  const [otp, setOtp] = useState<string[]>(Array(6).fill(''));
  const [serverError, setServerError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPass, setShowPass] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const { register, handleSubmit, formState: { errors } } = useForm<SignupInput>({ resolver: zodResolver(signupSchema) });

  const onStep1 = handleSubmit(async (data) => {
    setServerError('');
    setLoading(true);
    try {
      await sendVerification(data.email, 'REGISTER');
      setFormData(data);
      setStep(2);
    } catch (err) {
      setServerError(err instanceof Error ? err.message : 'Failed to send code');
    } finally {
      setLoading(false);
    }
  });
  const onStep2 = async () => {
    if (!formData) return;
    const code = otp.join('');
    if (code.length < 6) return void setServerError('Enter the 6-digit code');
    setServerError('');
    setLoading(true);
    try {
      await verifyVerification(formData.email, code, 'REGISTER');
      await registerAccount(formData);
      router.replace('/login');
    } catch (err) {
      setServerError(err instanceof Error ? err.message : 'Verification failed');
    } finally {
      setLoading(false);
    }
  };
  const resendOtp = async () => {
    if (!formData) return;
    setServerError('');
    setLoading(true);
    try {
      await sendVerification(formData.email, 'REGISTER');
      setOtp(Array(6).fill(''));
    } catch (err) {
      setServerError(err instanceof Error ? err.message : 'Failed to resend');
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

        {step === 1 && (
          <>
            <div style={{ textAlign: 'center', marginBottom: 32 }}>
              <h1 style={{ fontWeight: 600, fontSize: 28, color: 'var(--text)', letterSpacing: '-0.03em', marginBottom: 8 }}>Create account</h1>
              <p style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>Join the Binus campus community</p>
            </div>
            <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 16, padding: '28px 24px' }}>
              <RegisterForm register={register} errors={errors} serverError={serverError} loading={loading} showPass={showPass} showConfirm={showConfirm} onSubmit={onStep1} onTogglePass={() => setShowPass((value) => !value)} onToggleConfirm={() => setShowConfirm((value) => !value)} />
            </div>
          </>
        )}

        {step === 2 && (
          <>
            <div style={{ textAlign: 'center', marginBottom: 32 }}>
              <button onClick={() => { setStep(1); setServerError(''); }} style={{ display: 'flex', alignItems: 'center', gap: 6, background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-tertiary)', fontSize: 14, margin: '0 auto 16px', fontFamily: 'inherit', padding: 0 }}>
                <ArrowLeft size={14} />
                Back
              </button>
              <h1 style={{ fontWeight: 600, fontSize: 28, color: 'var(--text)', letterSpacing: '-0.03em', marginBottom: 8 }}>Check your email</h1>
              <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>We sent a 6-digit code to <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{formData?.email}</span></p>
            </div>
            <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 16, padding: '28px 24px' }}>
              <ErrorAlert message={serverError} />
              <div style={{ marginBottom: 24 }}><OtpInput value={otp} onChange={setOtp} /></div>
              <button type="button" onClick={onStep2} disabled={loading} style={{ width: '100%', padding: '11px 0', background: loading ? 'var(--accent-soft)' : 'var(--accent)', border: 'none', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, cursor: loading ? 'not-allowed' : 'pointer', letterSpacing: '-0.01em', fontFamily: 'inherit', marginTop: 4, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
                {loading ? <><span style={{ width: 14, height: 14, borderRadius: '50%', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', animation: 'spin 0.7s linear infinite', display: 'inline-block' }} />Please wait…</> : 'Verify & Create Account'}
              </button>
              <div style={{ textAlign: 'center', marginTop: 16 }}>
                <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Didn&apos;t receive it? </span>
                <button onClick={resendOtp} disabled={loading} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--accent)', fontSize: 13, fontWeight: 500, fontFamily: 'inherit', padding: 0 }}>Resend code</button>
              </div>
            </div>
          </>
        )}

        {step === 3 && (
          <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 16, padding: '40px 24px', textAlign: 'center' }}>
            <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 16 }}><CheckCircle2 size={52} color="var(--accent)" strokeWidth={1.5} /></div>
            <h2 style={{ fontWeight: 600, fontSize: 22, color: 'var(--text)', letterSpacing: '-0.03em', marginBottom: 8 }}>Account created!</h2>
            <p style={{ fontSize: 14, color: 'var(--text-tertiary)', marginBottom: 28 }}>Welcome to Binusgram. Sign in to get started.</p>
            <button onClick={() => router.replace('/login')} style={{ width: '100%', padding: '11px 0', background: 'var(--accent)', border: 'none', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, cursor: 'pointer', letterSpacing: '-0.01em', fontFamily: 'inherit' }}>Sign In</button>
          </div>
        )}

        {step !== 3 && (
          <>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, margin: '20px 0' }}>
              <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
              <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>or</span>
              <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
            </div>
            <div style={{ textAlign: 'center', fontSize: 14, color: 'var(--text-tertiary)' }}>
              Already have an account? <Link href="/login" style={{ color: 'var(--accent)', fontWeight: 500, textDecoration: 'none' }}>Sign in</Link>
            </div>
          </>
        )}
      </div>
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}
