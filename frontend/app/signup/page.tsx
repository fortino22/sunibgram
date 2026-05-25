'use client';

import Link from 'next/link';
import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowLeft } from 'lucide-react';
import Image from 'next/image';
import { useForm } from 'react-hook-form';

import { OtpInput } from '@/components/auth/OtpInput';
import { RegisterForm } from '@/components/auth/RegisterForm';
import { signupSchema, type SignupInput, type SignupStep } from '@/components/auth/schemas';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import { register as registerAccount, sendVerification, verifyVerification } from '@/lib/services/auth';
import { getFaculties, type Faculty } from '@/lib/services/faculties';

export type UsernameStatus = 'idle' | 'checking' | 'available' | 'taken';

const STEP1_ERRORS = /username|email already|name is required/i;
const STEP2_ERRORS = /faculty/i;

export default function SignupPage() {
  const router = useRouter();
  const [step, setStep] = useState<SignupStep>(1);
  const [formData, setFormData] = useState<SignupInput | null>(null);
  const [otp, setOtp] = useState<string[]>(Array(6).fill(''));
  const [serverError, setServerError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPass, setShowPass] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [avatar, setAvatar] = useState<File | null>(null);
  const [banner, setBanner] = useState<File | null>(null);
  const [faculties, setFaculties] = useState<Faculty[]>([]);
  const [usernameStatus, setUsernameStatus] = useState<UsernameStatus>('idle');

  const { register, handleSubmit, watch, trigger, formState: { errors } } = useForm<SignupInput>({
    resolver: zodResolver(signupSchema),
    mode: 'onTouched',
  });

  useEffect(() => {
    getFaculties('').then((res) => setFaculties(res.data ?? [])).catch(() => {});
  }, []);

  // Debounced username availability check
  const usernameValue = watch('username');
  const usernameTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (usernameTimer.current) clearTimeout(usernameTimer.current);

    if (!usernameValue || usernameValue.length < 3) {
      setUsernameStatus('idle');
      return;
    }

    setUsernameStatus('checking');
    usernameTimer.current = setTimeout(async () => {
      try {
        const baseUrl = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
        const res = await fetch(`${baseUrl}/explore/people?q=${encodeURIComponent(usernameValue)}`);
        if (!res.ok) { setUsernameStatus('idle'); return; }
        const json = await res.json() as { data?: Array<{ username: string }> };
        const people = json.data ?? [];
        const taken = people.some((p) => p.username.toLowerCase() === usernameValue.toLowerCase());
        setUsernameStatus(taken ? 'taken' : 'available');
      } catch {
        setUsernameStatus('idle');
      }
    }, 600);

    return () => { if (usernameTimer.current) clearTimeout(usernameTimer.current); };
  }, [usernameValue]);

  const onStep1: React.FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();
    const valid = await trigger(['name', 'username', 'email', 'password', 'confirmPassword']);
    if (valid) {
      setServerError('');
      setStep(2);
    }
  };

  const onStep2 = handleSubmit(async (data) => {
    setServerError('');
    setLoading(true);
    try {
      await sendVerification(data.email, 'REGISTER');
      setFormData(data);
      setStep(3);
    } catch (err) {
      setServerError(err instanceof Error ? err.message : 'Failed to send code');
    } finally {
      setLoading(false);
    }
  });

  const onStep3 = async () => {
    if (!formData) return;
    const code = otp.join('');
    if (code.length < 6) return void setServerError('Enter the 6-digit code');
    setServerError('');
    setLoading(true);
    try {
      await verifyVerification(formData.email, code, 'REGISTER');
      await registerAccount({ ...formData, avatar, banner });
      router.replace('/login');
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Verification failed';
      if (STEP1_ERRORS.test(msg)) setStep(1);
      else if (STEP2_ERRORS.test(msg)) setStep(2);
      setServerError(msg);
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

  const stepTitles: Record<SignupStep, { heading: string; sub: string }> = {
    1: { heading: 'Create account', sub: 'Join the Binus campus community' },
    2: { heading: 'Your profile', sub: 'Tell us a bit about yourself' },
    3: { heading: 'Check your email', sub: '' },
  };

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '24px 16px' }}>
      <div style={{ width: '100%', maxWidth: 400 }}>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
          <Image src="/assets/sunibgram/logo.png" alt="Sunibgram" width={48} height={48} />
        </div>

        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          {step > 1 && (
            <button
              onClick={() => { setStep((step - 1) as SignupStep); setServerError(''); }}
              style={{ display: 'flex', alignItems: 'center', gap: 6, background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-tertiary)', fontSize: 14, margin: '0 auto 16px', fontFamily: 'inherit', padding: 0 }}
            >
              <ArrowLeft size={14} />
              Back
            </button>
          )}
          <h1 style={{ fontWeight: 600, fontSize: 28, color: 'var(--text)', letterSpacing: '-0.03em', marginBottom: 8 }}>
            {stepTitles[step].heading}
          </h1>
          {step === 3 ? (
            <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
              We sent a 6-digit code to <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{formData?.email}</span>
            </p>
          ) : (
            <p style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>{stepTitles[step].sub}</p>
          )}
        </div>

        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 16, padding: '28px 24px' }}>
          {(step === 1 || step === 2) && (
            <RegisterForm
              register={register}
              errors={errors}
              serverError={serverError}
              loading={loading}
              showPass={showPass}
              showConfirm={showConfirm}
              faculties={faculties}
              usernameStatus={usernameStatus}
              currentStep={step}
              onSubmit={step === 1 ? onStep1 : onStep2}
              onTogglePass={() => setShowPass((v) => !v)}
              onToggleConfirm={() => setShowConfirm((v) => !v)}
              onAvatarChange={setAvatar}
              onBannerChange={setBanner}
            />
          )}

          {step === 3 && (
            <>
              <ErrorAlert message={serverError} />
              <div style={{ marginBottom: 24 }}><OtpInput value={otp} onChange={setOtp} /></div>
              <button type="button" onClick={onStep3} disabled={loading} style={{ width: '100%', padding: '11px 0', background: loading ? 'var(--accent-soft)' : 'var(--accent)', border: 'none', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, cursor: loading ? 'not-allowed' : 'pointer', letterSpacing: '-0.01em', fontFamily: 'inherit', marginTop: 4, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
                {loading ? <><span style={{ width: 14, height: 14, borderRadius: '50%', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', animation: 'spin 0.7s linear infinite', display: 'inline-block' }} />Please wait…</> : 'Verify & Create Account'}
              </button>
              <div style={{ textAlign: 'center', marginTop: 16 }}>
                <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Didn&apos;t receive it? </span>
                <button onClick={resendOtp} disabled={loading} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--accent)', fontSize: 13, fontWeight: 500, fontFamily: 'inherit', padding: 0 }}>Resend code</button>
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
          Already have an account? <Link href="/login" style={{ color: 'var(--accent)', fontWeight: 500, textDecoration: 'none' }}>Sign in</Link>
        </div>
      </div>
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}
