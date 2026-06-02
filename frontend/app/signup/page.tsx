'use client';

import Link from 'next/link';
import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowLeft } from 'lucide-react';
import { Logo } from '@/components/ui/Logo';
import { useForm } from 'react-hook-form';

import { AuthSplitLayout } from '@/components/auth/AuthSplitLayout';
import { OtpInput } from '@/components/auth/OtpInput';
import { RegisterForm } from '@/components/auth/RegisterForm';
import { signupSchema, type SignupInput, type SignupStep } from '@/components/auth/schemas';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import { register as registerAccount, sendVerification, verifyVerification, checkUsernameAvailable } from '@/lib/services/auth';
import { getFaculties, type Faculty } from '@/lib/services/faculties';
import { loadDraft, saveDraft, clearDraft } from '@/hooks/useSignupDraft';

export type UsernameStatus = 'idle' | 'checking' | 'available' | 'taken';

// Maps a backend validation error to the field + step that owns it, so the
// message lands where the user can fix it instead of in a generic banner.
function mapServerError(msg: string): { field?: keyof SignupInput; step?: SignupStep } {
  const m = msg.toLowerCase();
  if (m.includes('email already')) return { field: 'email', step: 1 };
  if (m.includes('username')) return { field: 'username', step: 1 };
  if (m.includes('password')) return { field: 'password', step: 1 };
  if (m.includes('gender')) return { field: 'gender', step: 2 };
  if (m.includes('birthday')) return { field: 'birthdayDate', step: 2 };
  if (m.includes('faculty')) return { field: 'facultyId', step: 2 };
  if (m.includes('terms')) return { field: 'agreeToTerms', step: 1 };
  return {};
}

export default function SignupPage() {
  const router = useRouter();

  // Start empty so the first client render matches the server HTML; the draft is
  // restored after mount (see hydration effect) to avoid an SSR hydration mismatch.
  const [step, setStep] = useState<SignupStep>(1);
  const hydratedRef = useRef(false);

  const [otp, setOtp] = useState<string[]>(Array(6).fill(''));
  const [serverError, setServerError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPass, setShowPass] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [avatar, setAvatar] = useState<File | null>(null);
  const [banner, setBanner] = useState<File | null>(null);
  const [faculties, setFaculties] = useState<Faculty[]>([]);
  const [facultiesLoading, setFacultiesLoading] = useState(true);
  const [usernameStatus, setUsernameStatus] = useState<UsernameStatus>('idle');

  const { register, handleSubmit, watch, trigger, setError, getValues, reset, formState: { errors } } =
    useForm<SignupInput>({
      resolver: zodResolver(signupSchema),
      mode: 'onChange',
      defaultValues: {},
    });

  // Restore the saved draft once, after mount — keeps SSR markup identical to the
  // first client render. hydratedRef gates the save effects below so the empty
  // pre-hydration state never clobbers the stored draft.
  useEffect(() => {
    const d = loadDraft();
    if (Object.keys(d.values).length > 0) reset(d.values);
    // A restored profile/OTP step is meaningless without the step-1 essentials.
    const safeStep = d.step >= 2 && !(d.values.email && d.values.password && d.values.name) ? 1 : d.step;
    setStep(safeStep);
    hydratedRef.current = true;
  }, [reset]);

  // Persist the draft on every change (sessionStorage). stepRef keeps the
  // subscription closure pointed at the current step without re-subscribing.
  const stepRef = useRef(step);
  useEffect(() => {
    stepRef.current = step;
    if (hydratedRef.current) saveDraft(getValues(), step);
  }, [step, getValues]);

  useEffect(() => {
    const sub = watch((values) => {
      if (hydratedRef.current) saveDraft(values, stepRef.current);
    });
    return () => sub.unsubscribe();
  }, [watch]);

  useEffect(() => {
    getFaculties()
      .then((res) => setFaculties(res.data ?? []))
      .catch(() => {})
      .finally(() => setFacultiesLoading(false));
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
        const { available } = await checkUsernameAvailable(usernameValue);
        setUsernameStatus(available ? 'available' : 'taken');
      } catch {
        setUsernameStatus('idle');
      }
    }, 600);

    return () => { if (usernameTimer.current) clearTimeout(usernameTimer.current); };
  }, [usernameValue]);

  function handleServerError(err: unknown, fallback: string) {
    const msg = err instanceof Error ? err.message : fallback;
    const { field, step: target } = mapServerError(msg);
    if (field) {
      setError(field, { message: msg });
      if (target) setStep(target);
      setServerError('');
    } else {
      setServerError(msg);
    }
  }

  const onStep1: React.FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();
    const valid = await trigger(['name', 'username', 'email', 'password', 'confirmPassword', 'agreeToTerms']);
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
      setStep(3);
    } catch (err) {
      handleServerError(err, 'Failed to send code');
    } finally {
      setLoading(false);
    }
  });

  const onStep3 = async () => {
    const data = getValues();
    const code = otp.join('');
    if (code.length < 6) return void setServerError('Enter the 6-digit code');
    setServerError('');
    setLoading(true);
    try {
      await verifyVerification(data.email, code, 'REGISTER');
      await registerAccount({ ...data, avatar, banner });
      clearDraft();
      router.replace('/login');
    } catch (err) {
      handleServerError(err, 'Verification failed');
    } finally {
      setLoading(false);
    }
  };

  const resendOtp = async () => {
    setServerError('');
    setLoading(true);
    try {
      await sendVerification(getValues('email'), 'REGISTER');
      setOtp(Array(6).fill(''));
    } catch (err) {
      handleServerError(err, 'Failed to resend');
    } finally {
      setLoading(false);
    }
  };

  const email = watch('email');

  const stepTitles: Record<SignupStep, { heading: string; sub: string }> = {
    1: { heading: 'Create account', sub: 'Join the Binus campus community' },
    2: { heading: 'Your profile', sub: 'Tell us a bit about yourself' },
    3: { heading: 'Check your email', sub: '' },
  };

  return (
    <AuthSplitLayout>
      <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 28 }}>
        <Logo size={48} style={{ color: 'var(--text)' }} />
      </div>

      <div style={{ textAlign: 'center', marginBottom: step < 3 ? 20 : 28 }}>
        <h1 style={{ fontWeight: 600, fontSize: 28, color: 'var(--text)', letterSpacing: '-0.03em', marginBottom: 8 }}>
          {stepTitles[step].heading}
        </h1>
        {step === 3 ? (
          <>
            <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
              We sent a 6-digit code to <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{email}</span>
            </p>
            <p style={{ fontSize: 13, color: 'var(--text-tertiary)', lineHeight: 1.5, marginTop: 6 }}>
              Don&apos;t see it? Please check your <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>spam</span> or junk folder.
            </p>
          </>
        ) : (
          <p style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>{stepTitles[step].sub}</p>
        )}
      </div>

      {/* Step indicator — below the heading, only for the two data-entry steps */}
      {step < 3 && (
        <div style={{ marginBottom: 28 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
            <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-tertiary)', letterSpacing: '0.02em' }}>
              STEP {step} OF 2
            </span>
            <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>
              {step === 1 ? 'Account' : 'Profile'}
            </span>
          </div>
          <div style={{ display: 'flex', gap: 6 }}>
            {[1, 2].map((i) => (
              <div
                key={i}
                style={{
                  flex: 1,
                  height: 4,
                  borderRadius: 999,
                  background: i <= step ? 'var(--accent)' : 'var(--border)',
                  transition: 'background 0.2s ease',
                }}
              />
            ))}
          </div>
        </div>
      )}

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
            facultiesLoading={facultiesLoading}
            usernameStatus={usernameStatus}
            currentStep={step}
            onSubmit={step === 1 ? onStep1 : onStep2}
            onBack={step === 2 ? () => { setStep(1); setServerError(''); } : undefined}
            onTogglePass={() => setShowPass((v) => !v)}
            onToggleConfirm={() => setShowConfirm((v) => !v)}
            onAvatarChange={setAvatar}
            onBannerChange={setBanner}
            passwordValue={watch('password') ?? ''}
          />
        )}

        {step === 3 && (
          <>
            <ErrorAlert message={serverError} />
            <div style={{ marginBottom: 24 }}><OtpInput value={otp} onChange={setOtp} /></div>
            <div style={{ display: 'flex', gap: 12 }}>
              <button
                type="button"
                onClick={() => { setStep(2); setServerError(''); }}
                disabled={loading}
                className="btn-ghost"
                style={{ flex: '0 0 auto', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6, padding: '11px 18px', border: '1px solid var(--border)', borderRadius: 10, background: 'transparent', color: 'var(--text)', fontWeight: 600, fontSize: 15, cursor: loading ? 'not-allowed' : 'pointer', fontFamily: 'inherit' }}
              >
                <ArrowLeft size={16} /> Back
              </button>
              <button
                type="button"
                onClick={onStep3}
                disabled={loading}
                className="btn-solid"
                style={{ flex: 1, padding: '11px 0', background: loading ? 'var(--accent-soft)' : 'var(--accent)', border: 'none', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, cursor: loading ? 'not-allowed' : 'pointer', letterSpacing: '-0.01em', fontFamily: 'inherit', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}
              >
                {loading ? <><span style={{ width: 14, height: 14, borderRadius: '50%', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', animation: 'spin 0.7s linear infinite', display: 'inline-block' }} />Please wait…</> : 'Verify & Create Account'}
              </button>
            </div>
            <div style={{ textAlign: 'center', marginTop: 16 }}>
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Didn&apos;t receive it? </span>
              <button onClick={resendOtp} disabled={loading} className="hover-dim" style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--accent)', fontSize: 13, fontWeight: 500, fontFamily: 'inherit', padding: 0 }}>Resend code</button>
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
    </AuthSplitLayout>
  );
}
