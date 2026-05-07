import { Eye, EyeOff } from 'lucide-react';
import type { FormEventHandler, ReactNode } from 'react';
import type { FieldErrors, UseFormRegister } from 'react-hook-form';

import { ErrorAlert } from '@/components/ui/ErrorAlert';
import type { SignupInput } from '@/components/auth/schemas';

interface RegisterFormProps {
  register: UseFormRegister<SignupInput>;
  errors: FieldErrors<SignupInput>;
  serverError: string;
  loading: boolean;
  showPass: boolean;
  showConfirm: boolean;
  onSubmit: FormEventHandler<HTMLFormElement>;
  onTogglePass: () => void;
  onToggleConfirm: () => void;
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

export function RegisterForm({
  register,
  errors,
  serverError,
  loading,
  showPass,
  showConfirm,
  onSubmit,
  onTogglePass,
  onToggleConfirm,
}: RegisterFormProps) {
  return (
    <>
      <ErrorAlert message={serverError} />
      <form onSubmit={onSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <FormField label="Full Name" error={errors.name?.message}>
          <input {...register('name')} type="text" placeholder="Your full name" autoComplete="name" style={inputStyle(!!errors.name)} />
        </FormField>

        <FormField label="Username" error={errors.username?.message}>
          <div style={{ position: 'relative' }}>
            <span style={{ position: 'absolute', left: 14, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', fontSize: 14, pointerEvents: 'none' }}>@</span>
            <input {...register('username')} type="text" placeholder="username" autoComplete="username" style={{ ...inputStyle(!!errors.username), paddingLeft: 28 }} />
          </div>
        </FormField>

        <FormField label="Binus Email" error={errors.email?.message}>
          <input {...register('email')} type="email" placeholder="you@binus.ac.id" autoComplete="email" style={inputStyle(!!errors.email)} />
        </FormField>

        <FormField label="Password" error={errors.password?.message}>
          <div style={{ position: 'relative' }}>
            <input {...register('password')} type={showPass ? 'text' : 'password'} placeholder="••••••••" autoComplete="new-password" style={{ ...inputStyle(!!errors.password), paddingRight: 42 }} />
            <button type="button" onClick={onTogglePass} style={toggleStyle}>
              {showPass ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
        </FormField>

        <FormField label="Confirm Password" error={errors.confirmPassword?.message}>
          <div style={{ position: 'relative' }}>
            <input {...register('confirmPassword')} type={showConfirm ? 'text' : 'password'} placeholder="••••••••" autoComplete="new-password" style={{ ...inputStyle(!!errors.confirmPassword), paddingRight: 42 }} />
            <button type="button" onClick={onToggleConfirm} style={toggleStyle}>
              {showConfirm ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
        </FormField>

        <button type="submit" disabled={loading} style={{ width: '100%', padding: '11px 0', background: loading ? 'var(--accent-soft)' : 'var(--accent)', border: 'none', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, cursor: loading ? 'not-allowed' : 'pointer', letterSpacing: '-0.01em', fontFamily: 'inherit', marginTop: 4, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
          {loading ? (
            <>
              <span style={{ width: 14, height: 14, borderRadius: '50%', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', animation: 'spin 0.7s linear infinite', display: 'inline-block' }} />
              Please wait…
            </>
          ) : (
            'Continue'
          )}
        </button>
      </form>
    </>
  );
}
