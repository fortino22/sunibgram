'use client';

import { useRef, useState } from 'react';
import { Eye, EyeOff, Camera, X, ArrowLeft, Check, ChevronDown } from 'lucide-react';
import type { FormEventHandler, ReactNode } from 'react';
import type { FieldErrors, UseFormRegister } from 'react-hook-form';

import { ErrorAlert } from '@/components/ui/ErrorAlert';
import { ImageCropModal } from '@/components/profile/ImageCropModal';
import type { SignupInput } from '@/components/auth/schemas';
import type { Faculty } from '@/lib/services/faculties';
import type { UsernameStatus } from '@/app/signup/page';

interface RegisterFormProps {
  register: UseFormRegister<SignupInput>;
  errors: FieldErrors<SignupInput>;
  serverError: string;
  loading: boolean;
  showPass: boolean;
  showConfirm: boolean;
  faculties: Faculty[];
  facultiesLoading?: boolean;
  usernameStatus: UsernameStatus;
  currentStep: 1 | 2;
  passwordValue?: string;
  onSubmit: FormEventHandler<HTMLFormElement>;
  onBack?: () => void;
  onTogglePass: () => void;
  onToggleConfirm: () => void;
  onAvatarChange: (file: File | null) => void;
  onBannerChange: (file: File | null) => void;
}

const PASSWORD_RULES: { label: string; test: (v: string) => boolean }[] = [
  { label: 'At least 10 characters', test: (v) => v.length >= 10 },
  { label: 'One uppercase letter', test: (v) => /[A-Z]/.test(v) },
  { label: 'One lowercase letter', test: (v) => /[a-z]/.test(v) },
  { label: 'One number', test: (v) => /[0-9]/.test(v) },
];

function PasswordHints({ value }: { value: string }) {
  if (!value) return null;
  return (
    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '4px 12px', marginTop: 8 }}>
      {PASSWORD_RULES.map((rule) => {
        const ok = rule.test(value);
        return (
          <div key={rule.label} style={{ display: 'flex', alignItems: 'center', gap: 5, fontSize: 11.5, color: ok ? '#22c55e' : 'var(--text-tertiary)' }}>
            <Check size={12} strokeWidth={3} style={{ opacity: ok ? 1 : 0.35 }} />
            {rule.label}
          </div>
        );
      })}
    </div>
  );
}

type PendingCrop = { src: string; filename: string; aspect: number; target: 'avatar' | 'banner' };

function FormField({ label, error, children }: { label: string; error?: string; children: ReactNode }) {
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

const selectStyle = (hasError = false) => ({
  width: '100%',
  padding: '10px 38px 10px 14px',
  background: 'var(--bg-subtle)',
  border: `1.5px solid ${hasError ? 'var(--coral)' : 'var(--border)'}`,
  borderRadius: 10,
  fontSize: 14,
  color: 'var(--text)',
  outline: 'none',
  fontFamily: 'inherit',
  cursor: 'pointer' as const,
  // Kill the native arrow so it doesn't collide with our custom chevron.
  appearance: 'none' as const,
  WebkitAppearance: 'none' as const,
  MozAppearance: 'none' as const,
});

// Wraps a native <select> with a custom chevron (native arrow is reset above).
function SelectShell({ children }: { children: ReactNode }) {
  return (
    <div style={{ position: 'relative' }}>
      {children}
      <ChevronDown
        size={18}
        style={{ position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', pointerEvents: 'none' }}
      />
    </div>
  );
}

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

const today = new Date().toISOString().split('T')[0];

export function RegisterForm({
  register,
  errors,
  serverError,
  loading,
  showPass,
  showConfirm,
  faculties,
  facultiesLoading = false,
  usernameStatus,
  currentStep,
  passwordValue = '',
  onSubmit,
  onBack,
  onTogglePass,
  onToggleConfirm,
  onAvatarChange,
  onBannerChange,
}: RegisterFormProps) {
  const avatarInputRef = useRef<HTMLInputElement>(null);
  const bannerInputRef = useRef<HTMLInputElement>(null);
  const [pendingCrop, setPendingCrop] = useState<PendingCrop | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [bannerPreview, setBannerPreview] = useState<string | null>(null);

  function pickFile(e: React.ChangeEvent<HTMLInputElement>, target: 'avatar' | 'banner') {
    const file = e.target.files?.[0];
    if (!file) return;
    setPendingCrop({ src: URL.createObjectURL(file), filename: file.name, aspect: target === 'avatar' ? 1 : 3, target });
    e.target.value = '';
  }

  function onCropComplete(file: File, preview: string) {
    if (pendingCrop?.target === 'avatar') {
      setAvatarPreview(preview);
      onAvatarChange(file);
    } else {
      setBannerPreview(preview);
      onBannerChange(file);
    }
    setPendingCrop(null);
  }

  return (
    <>
      <ErrorAlert message={serverError} />
      <form onSubmit={onSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>

        {currentStep === 1 && (
          <>
            <FormField label="Full Name" error={errors.name?.message}>
              <input {...register('name')} type="text" placeholder="Your full name" autoComplete="name" style={inputStyle(!!errors.name)} />
            </FormField>

            <FormField label="Username" error={errors.username?.message}>
              <div style={{ position: 'relative' }}>
                <span style={{ position: 'absolute', left: 14, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', fontSize: 14, pointerEvents: 'none' }}>@</span>
                <input {...register('username')} type="text" placeholder="username" autoComplete="username" style={{ ...inputStyle(!!errors.username), paddingLeft: 28, paddingRight: 36 }} />
                {usernameStatus === 'checking' && (
                  <span style={{ position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)', width: 14, height: 14, borderRadius: '50%', border: '2px solid var(--border)', borderTopColor: 'var(--accent)', animation: 'spin 0.7s linear infinite', display: 'inline-block' }} />
                )}
                {usernameStatus === 'available' && (
                  <span style={{ position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)', color: '#22c55e', fontSize: 16, lineHeight: 1 }}>✓</span>
                )}
                {usernameStatus === 'taken' && (
                  <span style={{ position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--coral)', fontSize: 16, lineHeight: 1 }}>✗</span>
                )}
              </div>
              {usernameStatus === 'available' && !errors.username && (
                <p style={{ fontSize: 12, color: '#22c55e', marginTop: 5 }}>Username available</p>
              )}
              {usernameStatus === 'taken' && !errors.username && (
                <p style={{ fontSize: 12, color: 'var(--coral)', marginTop: 5 }}>Username already taken</p>
              )}
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
              <PasswordHints value={passwordValue} />
            </FormField>

            <FormField label="Confirm Password" error={errors.confirmPassword?.message}>
              <div style={{ position: 'relative' }}>
                <input {...register('confirmPassword')} type={showConfirm ? 'text' : 'password'} placeholder="••••••••" autoComplete="new-password" style={{ ...inputStyle(!!errors.confirmPassword), paddingRight: 42 }} />
                <button type="button" onClick={onToggleConfirm} style={toggleStyle}>
                  {showConfirm ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
            </FormField>
          </>
        )}

        {currentStep === 2 && (
          <>
            <FormField label="Gender" error={errors.gender?.message}>
              <SelectShell>
                <select {...register('gender')} style={selectStyle(!!errors.gender)}>
                  <option value="">Select gender</option>
                  <option value="male">Male</option>
                  <option value="female">Female</option>
                  <option value="other">Other</option>
                  <option value="prefer_not_to_say">Prefer not to say</option>
                </select>
              </SelectShell>
            </FormField>

            <FormField label="Birthday (optional)" error={errors.birthdayDate?.message}>
              <input
                {...register('birthdayDate')}
                type="date"
                max={today}
                style={inputStyle(!!errors.birthdayDate)}
              />
            </FormField>

            <FormField label="Faculty" error={errors.facultyId?.message}>
              <SelectShell>
                <select {...register('facultyId')} disabled={facultiesLoading} style={{ ...selectStyle(!!errors.facultyId), opacity: facultiesLoading ? 0.6 : 1 }}>
                  <option value="">{facultiesLoading ? 'Loading faculties…' : 'Select faculty'}</option>
                  {faculties.map((f) => (
                    <option key={f.faculty_id} value={String(f.faculty_id)}>{f.faculty_name}</option>
                  ))}
                </select>
              </SelectShell>
              <p style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 6, lineHeight: 1.4 }}>
                Please choose the correct faculty of your study in Binus University. This <span style={{ color: 'var(--coral)', fontWeight: 500 }}>cannot be changed</span> later.
              </p>
            </FormField>

            <FormField label="Bio (optional)" error={errors.bio?.message}>
              <textarea
                {...register('bio')}
                placeholder="Tell us about yourself…"
                rows={3}
                style={{ ...inputStyle(!!errors.bio), resize: 'vertical' as const }}
              />
            </FormField>

            {/* Banner + Avatar image pickers */}
            <div>
              {/* Banner */}
              <div
                role="button"
                tabIndex={0}
                onClick={() => bannerInputRef.current?.click()}
                onKeyDown={(e) => e.key === 'Enter' && bannerInputRef.current?.click()}
                style={{
                  position: 'relative',
                  height: 110,
                  borderRadius: 10,
                  overflow: 'hidden',
                  cursor: 'pointer',
                  background: bannerPreview ? undefined : 'var(--bg-subtle)',
                  border: '1.5px solid var(--border)',
                }}
              >
                {bannerPreview && (
                  <img src={bannerPreview} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
                )}
                <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.35)' }}>
                  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 5, color: '#fff' }}>
                    <Camera size={20} strokeWidth={1.8} />
                    <span style={{ fontSize: 12, fontWeight: 500 }}>{bannerPreview ? 'Change banner' : 'Add banner'}</span>
                  </div>
                </div>
                {bannerPreview && (
                  <button
                    type="button"
                    onClick={(e) => { e.stopPropagation(); setBannerPreview(null); onBannerChange(null); }}
                    style={{ position: 'absolute', top: 8, right: 8, background: 'rgba(0,0,0,0.65)', border: 'none', borderRadius: '50%', width: 28, height: 28, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#fff' }}
                  >
                    <X size={14} strokeWidth={2} />
                  </button>
                )}
                <input ref={bannerInputRef} type="file" accept="image/*" onChange={(e) => pickFile(e, 'banner')} style={{ display: 'none' }} />
              </div>

              {/* Avatar — overlaps banner */}
              <div style={{ display: 'flex', alignItems: 'flex-end', gap: 12, marginTop: -24, paddingLeft: 16 }}>
                <div style={{ position: 'relative', flexShrink: 0 }}>
                  <div
                    role="button"
                    tabIndex={0}
                    onClick={() => avatarInputRef.current?.click()}
                    onKeyDown={(e) => e.key === 'Enter' && avatarInputRef.current?.click()}
                    style={{
                      width: 72,
                      height: 72,
                      borderRadius: '50%',
                      overflow: 'hidden',
                      cursor: 'pointer',
                      border: '3px solid var(--bg-card)', outline: '1.5px solid var(--border)',
                      background: avatarPreview ? undefined : 'var(--bg-subtle)',
                      position: 'relative',
                      flexShrink: 0,
                    }}
                  >
                    {avatarPreview && (
                      <img src={avatarPreview} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
                    )}
                    <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.4)', borderRadius: '50%' }}>
                      <Camera size={18} strokeWidth={1.8} color="#fff" />
                    </div>
                  </div>
                  <input ref={avatarInputRef} type="file" accept="image/*" onChange={(e) => pickFile(e, 'avatar')} style={{ display: 'none' }} />
                </div>
                <div style={{ paddingBottom: 6 }}>
                  <p style={{ fontSize: 13, color: 'var(--text-tertiary)', margin: 0 }}>Profile photo</p>
                  <div style={{ display: 'flex', gap: 10, marginTop: 2 }}>
                    <button type="button" onClick={() => avatarInputRef.current?.click()} style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: 13, color: 'var(--accent)', padding: 0, fontFamily: 'inherit' }}>
                      {avatarPreview ? 'Change' : 'Add photo'}
                    </button>
                    {avatarPreview && (
                      <button type="button" onClick={() => { setAvatarPreview(null); onAvatarChange(null); }} style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: 13, color: 'var(--coral)', padding: 0, fontFamily: 'inherit' }}>
                        Remove
                      </button>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </>
        )}

        <div style={{ display: 'flex', gap: 12, marginTop: 4 }}>
          {currentStep === 2 && onBack && (
            <button
              type="button"
              onClick={onBack}
              disabled={loading}
              className="btn-ghost"
              style={{ flex: '0 0 auto', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6, padding: '11px 18px', border: '1px solid var(--border)', borderRadius: 10, background: 'transparent', color: 'var(--text)', fontWeight: 600, fontSize: 15, cursor: loading ? 'not-allowed' : 'pointer', fontFamily: 'inherit' }}
            >
              <ArrowLeft size={16} /> Back
            </button>
          )}
          <button
            type="submit"
            className="btn-solid"
            disabled={loading}
            style={{ flex: 1, padding: '11px 0', background: loading ? 'var(--accent-soft)' : 'var(--accent)', border: 'none', borderRadius: 10, color: '#fff', fontWeight: 600, fontSize: 15, cursor: loading ? 'not-allowed' : 'pointer', letterSpacing: '-0.01em', fontFamily: 'inherit', transition: 'filter 0.12s ease', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}
          >
            {loading ? (
              <>
                <span style={{ width: 14, height: 14, borderRadius: '50%', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', animation: 'spin 0.7s linear infinite', display: 'inline-block' }} />
                Please wait…
              </>
            ) : currentStep === 1 ? 'Continue' : 'Send verification code'}
          </button>
        </div>
      </form>

      {pendingCrop && (
        <ImageCropModal
          imageSrc={pendingCrop.src}
          filename={pendingCrop.filename}
          aspect={pendingCrop.aspect}
          onComplete={onCropComplete}
          onClose={() => setPendingCrop(null)}
        />
      )}
    </>
  );
}
