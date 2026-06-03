'use client';

import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { AlertTriangle, Check, Loader2, Lock, X } from 'lucide-react';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useUpdateProfile } from '@/hooks/useUpdateProfile';
import { useDebouncedValue } from '@/hooks/useSearchShareUsers';
import { checkUsernameAvailable } from '@/lib/services/auth';
import { ErrorAlert } from '@/components/ui/ErrorAlert';
import {
  PrimaryButton,
  SettingsField,
  SettingsPanel,
  SuccessBanner,
  settingsInputStyle,
} from '@/components/settings/SettingsPanel';

const USERNAME_REGEX = /^[a-zA-Z0-9._]+$/;

const schema = z.object({
  name: z.string().min(1, 'Name is required').max(50, 'Max 50 characters'),
  username: z
    .string()
    .min(3, 'Min 3 characters')
    .max(30, 'Max 30 characters')
    .regex(USERNAME_REGEX, 'Only letters, numbers, dots, underscores'),
});

type FormValues = z.infer<typeof schema>;

type UsernameStatus =
  | { kind: 'idle' }
  | { kind: 'unchanged' }
  | { kind: 'invalid'; message: string }
  | { kind: 'checking' }
  | { kind: 'available' }
  | { kind: 'taken' }
  | { kind: 'error'; message: string };

export default function AccountSettingsPage() {
  const { data: profile, isLoading } = useMyProfile();
  const { mutate, isPending, error } = useUpdateProfile();
  const [success, setSuccess] = useState<string | null>(null);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [pendingValues, setPendingValues] = useState<FormValues | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors, isDirty },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', username: '' },
    mode: 'onChange',
  });

  useEffect(() => {
    if (profile) {
      reset({ name: profile.name ?? '', username: profile.username ?? '' });
    }
  }, [profile, reset]);

  const [now] = useState(() => Date.now());
  const nameLockedUntil = profile?.name_change_available_at
    ? new Date(profile.name_change_available_at)
    : null;
  const usernameLockedUntil = profile?.username_change_available_at
    ? new Date(profile.username_change_available_at)
    : null;
  const nameLocked = !!nameLockedUntil && nameLockedUntil.getTime() > now;
  const usernameLocked = !!usernameLockedUntil && usernameLockedUntil.getTime() > now;

  const usernameValue = watch('username') ?? '';
  const debouncedUsername = useDebouncedValue(usernameValue.trim(), 400);
  const currentUsername = profile?.username ?? '';

  const usernameIsValidShape =
    debouncedUsername.length >= 3 &&
    debouncedUsername.length <= 30 &&
    USERNAME_REGEX.test(debouncedUsername);
  const isUnchanged = debouncedUsername.toLowerCase() === currentUsername.toLowerCase();
  const shouldCheck =
    !!profile && !usernameLocked && usernameIsValidShape && !isUnchanged;

  const availabilityQuery = useQuery({
    queryKey: ['username-available', debouncedUsername],
    queryFn: () => checkUsernameAvailable(debouncedUsername).then((r) => r.available),
    enabled: shouldCheck,
    staleTime: 30_000,
    retry: false,
  });

  const usernameStatus: UsernameStatus = useMemo(() => {
    if (!profile) return { kind: 'idle' };
    if (usernameLocked) return { kind: 'idle' };
    const trimmed = usernameValue.trim();
    if (trimmed.length === 0) return { kind: 'idle' };
    if (trimmed.toLowerCase() === currentUsername.toLowerCase()) {
      return { kind: 'unchanged' };
    }
    if (trimmed.length < 3) return { kind: 'invalid', message: 'Min 3 characters' };
    if (trimmed.length > 30) return { kind: 'invalid', message: 'Max 30 characters' };
    if (!USERNAME_REGEX.test(trimmed)) {
      return { kind: 'invalid', message: 'Only letters, numbers, dots, underscores' };
    }
    if (trimmed !== debouncedUsername || availabilityQuery.isFetching) {
      return { kind: 'checking' };
    }
    if (availabilityQuery.isError) {
      return { kind: 'error', message: 'Could not check availability' };
    }
    if (availabilityQuery.data === true) return { kind: 'available' };
    if (availabilityQuery.data === false) return { kind: 'taken' };
    return { kind: 'checking' };
  }, [
    profile,
    usernameLocked,
    usernameValue,
    debouncedUsername,
    currentUsername,
    availabilityQuery.isFetching,
    availabilityQuery.isError,
    availabilityQuery.data,
  ]);

  const usernameBlocksSave =
    usernameStatus.kind === 'taken' ||
    usernameStatus.kind === 'invalid' ||
    usernameStatus.kind === 'checking' ||
    usernameStatus.kind === 'error';

  function onSubmit(values: FormValues) {
    setSuccess(null);
    const nameChanged = !nameLocked && values.name !== (profile?.name ?? '');
    const usernameChanged = !usernameLocked && values.username !== currentUsername;
    if (nameChanged || usernameChanged) {
      setPendingValues(values);
      setConfirmOpen(true);
      return;
    }
    submit(values);
  }

  function submit(values: FormValues) {
    const payload: { name?: string; username?: string } = {};
    if (!nameLocked) payload.name = values.name;
    if (!usernameLocked) payload.username = values.username;
    mutate(payload, {
      onSuccess: () => {
        setSuccess('Account updated.');
        reset(values);
        setConfirmOpen(false);
        setPendingValues(null);
      },
      onError: () => {
        setConfirmOpen(false);
        setPendingValues(null);
      },
    });
  }

  return (
    <SettingsPanel
      title="Your account"
      description="See information about your account, edit display name and username."
    >
      {isLoading || !profile ? (
        <SkeletonForm />
      ) : (
        <form onSubmit={handleSubmit(onSubmit)} style={{ maxWidth: 560 }}>
          <SuccessBanner message={success} />
          <ErrorAlert message={error ? (error as Error).message : null} />

          <CooldownInfoBanner />

          <SettingsField
            label="Display name"
            htmlFor="name"
            error={errors.name?.message}
            hint={
              nameLocked && nameLockedUntil
                ? `Locked — can change ${formatUnlock(nameLockedUntil)}`
                : undefined
            }
          >
            <div style={inputWrapStyle}>
              <input
                id="name"
                type="text"
                {...register('name')}
                disabled={nameLocked}
                style={{
                  ...settingsInputStyle(!!errors.name),
                  paddingRight: nameLocked ? 38 : undefined,
                  opacity: nameLocked ? 0.6 : 1,
                  cursor: nameLocked ? 'not-allowed' : 'text',
                }}
              />
              {nameLocked && (
                <span style={inputAdornmentStyle} aria-hidden>
                  <Lock size={16} />
                </span>
              )}
            </div>
          </SettingsField>

          <SettingsField
            label="Username"
            htmlFor="username"
            error={errors.username?.message}
            hint={
              usernameLocked && usernameLockedUntil
                ? `Locked — can change ${formatUnlock(usernameLockedUntil)}`
                : undefined
            }
          >
            <div style={inputWrapStyle}>
              <input
                id="username"
                type="text"
                autoComplete="off"
                spellCheck={false}
                {...register('username')}
                disabled={usernameLocked}
                style={{
                  ...settingsInputStyle(
                    !!errors.username ||
                      usernameStatus.kind === 'taken' ||
                      usernameStatus.kind === 'invalid'
                  ),
                  paddingRight: 38,
                  opacity: usernameLocked ? 0.6 : 1,
                  cursor: usernameLocked ? 'not-allowed' : 'text',
                }}
              />
              <UsernameStatusIcon status={usernameStatus} locked={usernameLocked} />
            </div>
            {!usernameLocked && <UsernameStatusText status={usernameStatus} />}
          </SettingsField>

          <SettingsField label="Email" hint="Email is verified and cannot be changed.">
            <input
              type="text"
              value={maskEmail(profile.email)}
              disabled
              style={{ ...settingsInputStyle(), opacity: 0.6, cursor: 'not-allowed' }}
            />
          </SettingsField>

          <div style={{ marginTop: 24, display: 'flex', justifyContent: 'flex-end' }}>
            <PrimaryButton disabled={isPending || !isDirty || usernameBlocksSave}>
              {isPending ? 'Saving…' : 'Save'}
            </PrimaryButton>
          </div>
        </form>
      )}

      {confirmOpen && pendingValues && profile && (
        <ConfirmChangeModal
          changingName={!nameLocked && pendingValues.name !== (profile.name ?? '')}
          changingUsername={!usernameLocked && pendingValues.username !== (profile.username ?? '')}
          isPending={isPending}
          onCancel={() => {
            setConfirmOpen(false);
            setPendingValues(null);
          }}
          onConfirm={() => submit(pendingValues)}
        />
      )}
    </SettingsPanel>
  );
}

const inputWrapStyle: React.CSSProperties = { position: 'relative' };

const inputAdornmentStyle: React.CSSProperties = {
  position: 'absolute',
  right: 12,
  top: '50%',
  transform: 'translateY(-50%)',
  display: 'inline-flex',
  alignItems: 'center',
  justifyContent: 'center',
  color: 'var(--text-tertiary)',
  pointerEvents: 'none',
};

function UsernameStatusIcon({
  status,
  locked,
}: {
  status: UsernameStatus;
  locked: boolean;
}) {
  if (locked) {
    return (
      <span style={inputAdornmentStyle} aria-hidden>
        <Lock size={16} />
      </span>
    );
  }
  switch (status.kind) {
    case 'checking':
      return (
        <span style={inputAdornmentStyle} aria-label="Checking username">
          <Loader2 size={16} className="animate-spin" style={{ animation: 'spin 0.8s linear infinite' }} />
        </span>
      );
    case 'available':
      return (
        <span style={{ ...inputAdornmentStyle, color: 'oklch(70% 0.15 145)' }} aria-label="Username available">
          <Check size={16} strokeWidth={3} />
        </span>
      );
    case 'taken':
    case 'invalid':
    case 'error':
      return (
        <span style={{ ...inputAdornmentStyle, color: 'var(--coral)' }} aria-label="Username unavailable">
          <X size={16} strokeWidth={3} />
        </span>
      );
    default:
      return null;
  }
}

function UsernameStatusText({ status }: { status: UsernameStatus }) {
  let text: string | null = null;
  let color = 'var(--text-tertiary)';
  switch (status.kind) {
    case 'checking':
      text = 'Checking availability…';
      break;
    case 'available':
      text = 'Username available';
      color = 'oklch(70% 0.15 145)';
      break;
    case 'taken':
      text = 'Username already taken';
      color = 'var(--coral)';
      break;
    case 'invalid':
      text = status.message;
      color = 'var(--coral)';
      break;
    case 'error':
      text = status.message;
      color = 'var(--coral)';
      break;
    case 'unchanged':
      text = 'This is your current username';
      break;
    default:
      text = null;
  }
  if (!text) return null;
  return (
    <div style={{ fontSize: 12, color, marginTop: 4 }}>
      <style>{`@keyframes spin { to { transform: translateY(-50%) rotate(360deg); } }`}</style>
      {text}
    </div>
  );
}

function CooldownInfoBanner() {
  return (
    <div
      style={{
        display: 'flex',
        gap: 10,
        alignItems: 'flex-start',
        padding: '10px 14px',
        background: 'oklch(75% 0.12 75 / 0.10)',
        border: '1px solid oklch(75% 0.12 75 / 0.30)',
        borderRadius: 10,
        marginBottom: 18,
        color: 'oklch(80% 0.12 75)',
      }}
      role="note"
    >
      <AlertTriangle size={16} style={{ marginTop: 2, flexShrink: 0 }} />
      <div style={{ fontSize: 13, lineHeight: 1.45 }}>
        Display name and username can only be changed{' '}
        <strong style={{ fontWeight: 700 }}>once every 7 days</strong>. Choose carefully.
      </div>
    </div>
  );
}

function ConfirmChangeModal({
  changingName,
  changingUsername,
  isPending,
  onCancel,
  onConfirm,
}: {
  changingName: boolean;
  changingUsername: boolean;
  isPending: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  const fields = [changingName ? 'display name' : null, changingUsername ? 'username' : null]
    .filter(Boolean)
    .join(' and ');
  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 50,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
        backdropFilter: 'blur(4px)',
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget && !isPending) onCancel();
      }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 14,
          width: '100%',
          maxWidth: 400,
          padding: 22,
        }}
      >
        <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start', marginBottom: 14 }}>
          <div
            style={{
              width: 36,
              height: 36,
              borderRadius: '50%',
              background: 'oklch(75% 0.12 75 / 0.18)',
              color: 'oklch(80% 0.12 75)',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
            }}
          >
            <AlertTriangle size={18} />
          </div>
          <div style={{ flex: 1 }}>
            <h3 style={{ margin: 0, fontSize: 16, fontWeight: 700, color: 'var(--text)' }}>
              Confirm change
            </h3>
            <p style={{ margin: '6px 0 0', fontSize: 13, color: 'var(--text-tertiary)', lineHeight: 1.5 }}>
              Your {fields} will be locked for <strong>7 days</strong> after saving. Continue?
            </p>
          </div>
        </div>
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, marginTop: 6 }}>
          <button
            type="button"
            onClick={onCancel}
            disabled={isPending}
            style={{
              background: 'transparent',
              color: 'var(--text)',
              border: '1px solid var(--border)',
              borderRadius: 10,
              padding: '9px 16px',
              fontWeight: 600,
              fontSize: 14,
              cursor: isPending ? 'not-allowed' : 'pointer',
              fontFamily: 'inherit',
            }}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isPending}
            style={{
              background: 'var(--text)',
              color: 'var(--bg)',
              border: 'none',
              borderRadius: 10,
              padding: '9px 18px',
              fontWeight: 700,
              fontSize: 14,
              cursor: isPending ? 'not-allowed' : 'pointer',
              opacity: isPending ? 0.6 : 1,
              fontFamily: 'inherit',
            }}
          >
            {isPending ? 'Saving…' : 'Confirm'}
          </button>
        </div>
      </div>
    </div>
  );
}

function formatUnlock(date: Date): string {
  const days = Math.ceil((date.getTime() - Date.now()) / (1000 * 60 * 60 * 24));
  const formatted = date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
  if (days <= 1) return `in less than a day (${formatted})`;
  return `in ${days} days (${formatted})`;
}

function maskEmail(email: string | null | undefined): string {
  if (!email) return '';
  const [local, domain] = email.split('@');
  if (!domain) return email;
  const visible = local.slice(0, 4);
  const hidden = '*'.repeat(Math.max(local.length - 4, 3));
  return `${visible}${hidden}@${domain}`;
}

function SkeletonForm() {
  return (
    <div style={{ maxWidth: 560 }}>
      {[0, 1, 2].map((i) => (
        <div key={i} style={{ marginBottom: 18 }}>
          <div style={{ width: 90, height: 12, background: 'var(--bg-subtle)', borderRadius: 4, marginBottom: 8 }} />
          <div style={{ width: '100%', height: 42, background: 'var(--bg-subtle)', borderRadius: 8 }} />
        </div>
      ))}
    </div>
  );
}
