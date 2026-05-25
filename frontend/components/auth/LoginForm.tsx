import { Eye, EyeOff } from 'lucide-react';
import type { FormEventHandler, ReactNode } from 'react';
import type { FieldErrors, UseFormRegister } from 'react-hook-form';

import { ErrorAlert } from '@/components/ui/ErrorAlert';
import type { LoginInput } from '@/components/auth/schemas';

interface LoginFormProps {
  register: UseFormRegister<LoginInput>;
  errors: FieldErrors<LoginInput>;
  serverError: string;
  isSubmitting: boolean;
  showPassword: boolean;
  onSubmit: FormEventHandler<HTMLFormElement>;
  onTogglePassword: () => void;
  onForgotPassword: () => void;
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
      <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text-secondary)', marginBottom: 6 }}>
        {label}
      </label>
      {children}
      {error && <p style={{ fontSize: 12, color: 'var(--coral)', marginTop: 5 }}>{error}</p>}
    </div>
  );
}

export function LoginForm({
  register,
  errors,
  serverError,
  isSubmitting,
  showPassword,
  onSubmit,
  onTogglePassword,
  onForgotPassword,
}: LoginFormProps) {
  return (
    <>
      <ErrorAlert message={serverError} />
      <form onSubmit={onSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <FormField label="Email" error={errors.email?.message}>
          <input
            {...register('email')}
            type="email"
            placeholder="you@binus.ac.id"
            autoComplete="email"
            style={{ width: '100%', padding: '10px 14px', background: 'var(--bg-subtle)', border: `1.5px solid ${errors.email ? 'var(--coral)' : 'var(--border)'}`, borderRadius: 10, fontSize: 14, color: 'var(--text)', outline: 'none', fontFamily: 'inherit', transition: 'border-color 0.15s ease' }}
            onFocus={(event) => {
              if (!errors.email) event.target.style.borderColor = 'var(--accent)';
            }}
            onBlur={(event) => {
              if (!errors.email) event.target.style.borderColor = 'var(--border)';
            }}
          />
        </FormField>

        <FormField label="Password" error={errors.password?.message}>
          <div style={{ position: 'relative' }}>
            <input
              {...register('password')}
              type={showPassword ? 'text' : 'password'}
              placeholder="••••••••"
              autoComplete="current-password"
              style={{ width: '100%', padding: '10px 42px 10px 14px', background: 'var(--bg-subtle)', border: `1.5px solid ${errors.password ? 'var(--coral)' : 'var(--border)'}`, borderRadius: 10, fontSize: 14, color: 'var(--text)', outline: 'none', fontFamily: 'inherit', transition: 'border-color 0.15s ease' }}
              onFocus={(event) => {
                if (!errors.password) event.target.style.borderColor = 'var(--accent)';
              }}
              onBlur={(event) => {
                if (!errors.password) event.target.style.borderColor = 'var(--border)';
              }}
            />
            <button type="button" onClick={onTogglePassword} style={{ position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)', background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center', padding: 0 }}>
              {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
        </FormField>

        <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: -6 }}>
          <button
            type="button"
            onClick={onForgotPassword}
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
            Forgot password?
          </button>
        </div>

        <button type="submit" disabled={isSubmitting} style={{ width: '100%', padding: '11px 0', background: isSubmitting ? 'var(--accent-soft)' : 'var(--accent)', border: 'none', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, cursor: isSubmitting ? 'not-allowed' : 'pointer', letterSpacing: '-0.01em', fontFamily: 'inherit', transition: 'background 0.15s ease', marginTop: 4, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
          {isSubmitting ? (
            <>
              <span style={{ width: 14, height: 14, borderRadius: '50%', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', animation: 'spin 0.7s linear infinite', display: 'inline-block' }} />
              Verifying…
            </>
          ) : (
            'Login'
          )}
        </button>
      </form>
    </>
  );
}
