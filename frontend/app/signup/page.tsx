'use client';

import { useState, useRef, type KeyboardEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Eye, EyeOff, ArrowLeft, CheckCircle2 } from 'lucide-react';
import Image from 'next/image';
import { register as registerAccount, sendVerification, verifyVerification } from '@/lib/services/auth';

// Schemas

const step1Schema = z
  .object({
    name: z.string().min(2, 'Name must be at least 2 characters'),
    username: z
      .string()
      .min(3, 'Username must be at least 3 characters')
      .max(30, 'Username too long')
      .regex(/^[a-zA-Z0-9_]+$/, 'Only letters, numbers, underscores'),
    email: z
      .string()
      .email('Invalid email address')
      .endsWith('@binus.ac.id', 'Must use your @binus.ac.id email'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string(),
  })
  .refine((d) => d.password === d.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

type Step1Input = z.infer<typeof step1Schema>;

// Input component

interface FieldProps {
  label: string;
  error?: string;
  children: React.ReactNode;
}

function Field({ label, error, children }: FieldProps) {
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
      {error && (
        <p style={{ fontSize: 12, color: 'var(--coral)', marginTop: 5 }}>{error}</p>
      )}
    </div>
  );
}

const inputStyle = (hasError: boolean): React.CSSProperties => ({
  width: '100%',
  padding: '10px 14px',
  background: 'var(--bg-subtle)',
  border: `1.5px solid ${hasError ? 'var(--coral)' : 'var(--border)'}`,
  borderRadius: 10,
  fontSize: 14,
  color: 'var(--text)',
  outline: 'none',
  fontFamily: 'inherit',
  boxSizing: 'border-box',
});

// OTP Input

interface OtpInputProps {
  value: string[];
  onChange: (v: string[]) => void;
}

function OtpInput({ value, onChange }: OtpInputProps) {
  const refs = useRef<(HTMLInputElement | null)[]>([]);

  const handleChange = (i: number, char: string) => {
    const digit = char.replace(/\D/g, '').slice(-1);
    const next = [...value];
    next[i] = digit;
    onChange(next);
    if (digit && i < 5) refs.current[i + 1]?.focus();
  };

  const handleKeyDown = (i: number, e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !value[i] && i > 0) {
      refs.current[i - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const digits = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
    const next = [...value];
    digits.split('').forEach((d, i) => { next[i] = d; });
    onChange(next);
    const focusIdx = Math.min(digits.length, 5);
    refs.current[focusIdx]?.focus();
  };

  return (
    <div style={{ display: 'flex', gap: 10, justifyContent: 'center' }}>
      {value.map((digit, i) => (
        <input
          key={i}
          ref={(el) => { refs.current[i] = el; }}
          type="text"
          inputMode="numeric"
          maxLength={1}
          value={digit}
          onChange={(e) => handleChange(i, e.target.value)}
          onKeyDown={(e) => handleKeyDown(i, e)}
          onPaste={handlePaste}
          style={{
            width: 48,
            height: 56,
            textAlign: 'center',
            fontSize: 22,
            fontWeight: 600,
            background: 'var(--bg-subtle)',
            border: `1.5px solid ${digit ? 'var(--accent)' : 'var(--border)'}`,
            borderRadius: 12,
            color: 'var(--text)',
            outline: 'none',
            fontFamily: 'inherit',
            caretColor: 'var(--accent)',
          }}
        />
      ))}
    </div>
  );
}

// Page

type Step = 1 | 2 | 3;

export default function SignupPage() {
  const router = useRouter();
  const [step, setStep] = useState<Step>(1);
  const [formData, setFormData] = useState<Step1Input | null>(null);
  const [otp, setOtp] = useState<string[]>(Array(6).fill(''));
  const [serverError, setServerError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPass, setShowPass] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Step1Input>({ resolver: zodResolver(step1Schema) });

  // Step 1 → send OTP
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

  // Step 2 → verify OTP → register
  const onStep2 = async () => {
    if (!formData) return;
    const code = otp.join('');
    if (code.length < 6) {
      setServerError('Enter the 6-digit code');
      return;
    }
    setServerError('');
    setLoading(true);
    try {
      await verifyVerification(formData.email, code, 'REGISTER');
      await registerAccount({
        name: formData.name,
        username: formData.username,
        email: formData.email,
        password: formData.password,
        confirmPassword: formData.confirmPassword,
      });
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

  const submitBtn = (label: string, onClick?: () => void) => (
    <button
      type={onClick ? 'button' : 'submit'}
      onClick={onClick}
      disabled={loading}
      style={{
        width: '100%',
        padding: '11px 0',
        background: loading ? 'var(--accent-soft)' : 'var(--accent)',
        border: 'none',
        borderRadius: 10,
        color: '#fff',
        fontWeight: 600,
        fontSize: 15,
        cursor: loading ? 'not-allowed' : 'pointer',
        letterSpacing: '-0.01em',
        fontFamily: 'inherit',
        marginTop: 4,
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
          Please wait…
        </>
      ) : (
        label
      )}
    </button>
  );

  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'var(--bg)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px 16px',
      }}
    >
      <div style={{ width: '100%', maxWidth: 400 }}>
        {/* Logo */}
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
          <Image src="/assets/sunibgram/logo.png" alt="Sunibgram" width={48} height={48} />
        </div>

        {/* Step 1: Registration form */}
        {step === 1 && (
          <>
            <div style={{ textAlign: 'center', marginBottom: 32 }}>
              <h1
                style={{
                  fontWeight: 600,
                  fontSize: 28,
                  color: 'var(--text)',
                  letterSpacing: '-0.03em',
                  marginBottom: 8,
                }}
              >
                Create account
              </h1>
              <p style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>
                Join the Binus campus community
              </p>
            </div>

            <div
              style={{
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 16,
                padding: '28px 24px',
              }}
            >
              {serverError && (
                <div
                  style={{
                    padding: '10px 14px',
                    background: 'var(--coral-soft)',
                    border: '1px solid oklch(64% 0.18 25 / 0.3)',
                    borderRadius: 10,
                    fontSize: 13,
                    color: 'var(--coral)',
                    marginBottom: 20,
                  }}
                >
                  {serverError}
                </div>
              )}

              <form onSubmit={onStep1} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                <Field label="Full Name" error={errors.name?.message}>
                  <input
                    {...register('name')}
                    type="text"
                    placeholder="Your full name"
                    autoComplete="name"
                    style={inputStyle(!!errors.name)}
                  />
                </Field>

                <Field label="Username" error={errors.username?.message}>
                  <div style={{ position: 'relative' }}>
                    <span
                      style={{
                        position: 'absolute',
                        left: 14,
                        top: '50%',
                        transform: 'translateY(-50%)',
                        color: 'var(--text-tertiary)',
                        fontSize: 14,
                        pointerEvents: 'none',
                      }}
                    >
                      @
                    </span>
                    <input
                      {...register('username')}
                      type="text"
                      placeholder="username"
                      autoComplete="username"
                      style={{ ...inputStyle(!!errors.username), paddingLeft: 28 }}
                    />
                  </div>
                </Field>

                <Field label="Binus Email" error={errors.email?.message}>
                  <input
                    {...register('email')}
                    type="email"
                    placeholder="you@binus.ac.id"
                    autoComplete="email"
                    style={inputStyle(!!errors.email)}
                  />
                </Field>

                <Field label="Password" error={errors.password?.message}>
                  <div style={{ position: 'relative' }}>
                    <input
                      {...register('password')}
                      type={showPass ? 'text' : 'password'}
                      placeholder="••••••••"
                      autoComplete="new-password"
                      style={{ ...inputStyle(!!errors.password), paddingRight: 42 }}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPass((v) => !v)}
                      style={{
                        position: 'absolute',
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
                      }}
                    >
                      {showPass ? <EyeOff size={16} /> : <Eye size={16} />}
                    </button>
                  </div>
                </Field>

                <Field label="Confirm Password" error={errors.confirmPassword?.message}>
                  <div style={{ position: 'relative' }}>
                    <input
                      {...register('confirmPassword')}
                      type={showConfirm ? 'text' : 'password'}
                      placeholder="••••••••"
                      autoComplete="new-password"
                      style={{ ...inputStyle(!!errors.confirmPassword), paddingRight: 42 }}
                    />
                    <button
                      type="button"
                      onClick={() => setShowConfirm((v) => !v)}
                      style={{
                        position: 'absolute',
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
                      }}
                    >
                      {showConfirm ? <EyeOff size={16} /> : <Eye size={16} />}
                    </button>
                  </div>
                </Field>

                {submitBtn('Continue')}
              </form>
            </div>
          </>
        )}

        {/* Step 2: OTP verification */}
        {step === 2 && (
          <>
            <div style={{ textAlign: 'center', marginBottom: 32 }}>
              <button
                onClick={() => { setStep(1); setServerError(''); }}
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
                <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>
                  {formData?.email}
                </span>
              </p>
            </div>

            <div
              style={{
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 16,
                padding: '28px 24px',
              }}
            >
              {serverError && (
                <div
                  style={{
                    padding: '10px 14px',
                    background: 'var(--coral-soft)',
                    border: '1px solid oklch(64% 0.18 25 / 0.3)',
                    borderRadius: 10,
                    fontSize: 13,
                    color: 'var(--coral)',
                    marginBottom: 20,
                  }}
                >
                  {serverError}
                </div>
              )}

              <div style={{ marginBottom: 24 }}>
                <OtpInput value={otp} onChange={setOtp} />
              </div>

              {submitBtn('Verify & Create Account', onStep2)}

              <div style={{ textAlign: 'center', marginTop: 16 }}>
                <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
                  Didn&apos;t receive it?{' '}
                </span>
                <button
                  onClick={resendOtp}
                  disabled={loading}
                  style={{
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    color: 'var(--accent)',
                    fontSize: 13,
                    fontWeight: 500,
                    fontFamily: 'inherit',
                    padding: 0,
                  }}
                >
                  Resend code
                </button>
              </div>
            </div>
          </>
        )}

        {/* Step 3: Success */}
        {step === 3 && (
          <div
            style={{
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: 16,
              padding: '40px 24px',
              textAlign: 'center',
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
              Account created!
            </h2>
            <p style={{ fontSize: 14, color: 'var(--text-tertiary)', marginBottom: 28 }}>
              Welcome to Binusgram. Sign in to get started.
            </p>
            <button
              onClick={() => router.replace('/login')}
              style={{
                width: '100%',
                padding: '11px 0',
                background: 'var(--accent)',
                border: 'none',
                borderRadius: 10,
                color: '#fff',
                fontWeight: 600,
                fontSize: 15,
                cursor: 'pointer',
                letterSpacing: '-0.01em',
                fontFamily: 'inherit',
              }}
            >
              Sign In
            </button>
          </div>
        )}

        {/* Divider + login link (steps 1 & 2 only) */}
        {step !== 3 && (
          <>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                margin: '20px 0',
              }}
            >
              <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
              <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>or</span>
              <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
            </div>

            <div style={{ textAlign: 'center', fontSize: 14, color: 'var(--text-tertiary)' }}>
              Already have an account?{' '}
              <Link
                href="/login"
                style={{ color: 'var(--accent)', fontWeight: 500, textDecoration: 'none' }}
              >
                Sign in
              </Link>
            </div>
          </>
        )}
      </div>

      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}
