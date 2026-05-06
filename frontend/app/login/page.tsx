'use client';

import { useRef, useState, type ClipboardEvent, type KeyboardEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { ArrowLeft, Eye, EyeOff } from 'lucide-react';
import Image from 'next/image';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import { login, loginProd, sendLoginOtp } from '@/lib/services/auth';
import { useAuthStore } from '@/stores/authStore';

const loginSchema = z.object({
  email: z
    .string()
    .email('Invalid email address')
    .endsWith('@binus.ac.id', 'Must use your @binus.ac.id email'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

type LoginInput = z.infer<typeof loginSchema>;
type Step = 'credentials' | 'otp';

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

  const handlePaste = (e: ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const digits = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
    const next = [...value];
    digits.split('').forEach((d, i) => {
      next[i] = d;
    });
    onChange(next);
    const focusIdx = Math.min(digits.length, 5);
    refs.current[focusIdx]?.focus();
  };

  return (
    <div style={{ display: 'flex', gap: 10, justifyContent: 'center' }}>
      {value.map((digit, i) => (
        <input
          key={i}
          ref={(el) => {
            refs.current[i] = el;
          }}
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

export default function LoginPage() {
  const router = useRouter();
  const setToken = useAuthStore((s) => s.setToken);
  const [showPassword, setShowPassword] = useState(false);
  const [serverError, setServerError] = useState('');
  const [step, setStep] = useState<Step>('credentials');
  const [formData, setFormData] = useState<LoginInput | null>(null);
  const [otp, setOtp] = useState<string[]>(Array(6).fill(''));
  const [loading, setLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginInput>({ resolver: zodResolver(loginSchema) });

  const setSessionCookie = async (sid: string) => {
    const response = await fetch('/api/auth/set-session', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sid }),
    });

    if (!response.ok) {
      throw new Error('Failed to create session');
    }
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
    if (code.length < 6) {
      setServerError('Enter the 6-digit code');
      return;
    }

    setServerError('');
    setLoading(true);
    try {
      const res = await login({
        email: formData.email,
        password: formData.password,
        otp: code,
      });
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

        {/* Heading */}
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          {step === 'otp' && (
            <button
              onClick={() => {
                setStep('credentials');
                setServerError('');
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
          )}
          <h1
            style={{
              fontWeight: 600,
              fontSize: 28,
              color: 'var(--text)',
              letterSpacing: '-0.03em',
              marginBottom: 8,
            }}
          >
            {step === 'credentials' ? 'Sign in to Binusgram' : 'Check your email'}
          </h1>
          <p style={{ fontSize: 14, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
            {step === 'credentials' ? (
              'Campus discourse for Binus students'
            ) : (
              <>
                We sent a 6-digit code to{' '}
                <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>
                  {formData?.email}
                </span>
              </>
            )}
          </p>
        </div>

        {/* Card */}
        <div
          style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 16,
            padding: '28px 24px',
          }}
        >
          <ErrorAlert message={serverError} />

          {step === 'credentials' ? (
          <form onSubmit={onSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {/* Email */}
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
                Email
              </label>
              <input
                {...register('email')}
                type="email"
                placeholder="you@binus.ac.id"
                autoComplete="email"
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  background: 'var(--bg-subtle)',
                  border: `1.5px solid ${errors.email ? 'var(--coral)' : 'var(--border)'}`,
                  borderRadius: 10,
                  fontSize: 14,
                  color: 'var(--text)',
                  outline: 'none',
                  fontFamily: 'inherit',
                  transition: 'border-color 0.15s ease',
                }}
                onFocus={(e) => {
                  if (!errors.email) e.target.style.borderColor = 'var(--accent)';
                }}
                onBlur={(e) => {
                  if (!errors.email) e.target.style.borderColor = 'var(--border)';
                }}
              />
              {errors.email && (
                <p style={{ fontSize: 12, color: 'var(--coral)', marginTop: 5 }}>
                  {errors.email.message}
                </p>
              )}
            </div>

            {/* Password */}
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
                Password
              </label>
              <div style={{ position: 'relative' }}>
                <input
                  {...register('password')}
                  type={showPassword ? 'text' : 'password'}
                  placeholder="••••••••"
                  autoComplete="current-password"
                  style={{
                    width: '100%',
                    padding: '10px 42px 10px 14px',
                    background: 'var(--bg-subtle)',
                    border: `1.5px solid ${errors.password ? 'var(--coral)' : 'var(--border)'}`,
                    borderRadius: 10,
                    fontSize: 14,
                    color: 'var(--text)',
                    outline: 'none',
                    fontFamily: 'inherit',
                    transition: 'border-color 0.15s ease',
                  }}
                  onFocus={(e) => {
                    if (!errors.password) e.target.style.borderColor = 'var(--accent)';
                  }}
                  onBlur={(e) => {
                    if (!errors.password) e.target.style.borderColor = 'var(--border)';
                  }}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
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
                  {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
              {errors.password && (
                <p style={{ fontSize: 12, color: 'var(--coral)', marginTop: 5 }}>
                  {errors.password.message}
                </p>
              )}
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={isSubmitting}
              style={{
                width: '100%',
                padding: '11px 0',
                background: isSubmitting ? 'var(--accent-soft)' : 'var(--accent)',
                border: 'none',
                borderRadius: 10,
                color: '#fff',
                fontWeight: 600,
                fontSize: 15,
                cursor: isSubmitting ? 'not-allowed' : 'pointer',
                letterSpacing: '-0.01em',
                fontFamily: 'inherit',
                transition: 'background 0.15s ease',
                marginTop: 4,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 8,
              }}
            >
              {isSubmitting ? (
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
                  Signing in…
                </>
              ) : (
                'Sign In'
              )}
            </button>
          </form>
          ) : (
            <>
              <div style={{ marginBottom: 24 }}>
                <OtpInput value={otp} onChange={setOtp} />
              </div>

              <button
                type="button"
                onClick={onVerifyOtp}
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
                  transition: 'background 0.15s ease',
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
                    Signing in…
                  </>
                ) : (
                  'Verify & Sign In'
                )}
              </button>
            </>
          )}
        </div>

        {/* Divider */}
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

        {/* Signup link */}
        <div
          style={{
            textAlign: 'center',
            fontSize: 14,
            color: 'var(--text-tertiary)',
          }}
        >
          Don&apos;t have an account?{' '}
          <Link
            href="/signup"
            style={{
              color: 'var(--accent)',
              fontWeight: 500,
              textDecoration: 'none',
            }}
          >
            Create account
          </Link>
        </div>
      </div>

      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}
