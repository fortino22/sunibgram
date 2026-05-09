'use client';

import { useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Camera, X } from 'lucide-react';
import { useUpdateProfile } from '@/hooks/useUpdateProfile';
import { Avatar } from '@/components/shared/Avatar';
import { ImageCropModal } from '@/components/profile/ImageCropModal';
import type { UserProfile } from '@/lib/services/profile';

const schema = z.object({
  name: z.string().min(1, 'Name is required').max(50, 'Max 50 characters'),
  username: z
    .string()
    .min(3, 'Min 3 characters')
    .max(30, 'Max 30 characters')
    .regex(/^[a-zA-Z0-9._]+$/, 'Only letters, numbers, dots, underscores'),
  bio: z.string().max(160, 'Max 160 characters').optional(),
});

type FormValues = z.infer<typeof schema>;
type PendingCrop = { src: string; filename: string; aspect: number; target: 'avatar' | 'banner' };

interface ProfileEditModalProps {
  profile: UserProfile;
  onClose: () => void;
}

export function ProfileEditModal({ profile, onClose }: ProfileEditModalProps) {
  const { mutate, isPending, error } = useUpdateProfile();
  const avatarInputRef = useRef<HTMLInputElement>(null);
  const bannerInputRef = useRef<HTMLInputElement>(null);

  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(profile.avatar ?? null);
  const [bannerFile, setBannerFile] = useState<File | null>(null);
  const [bannerPreview, setBannerPreview] = useState<string | null>(profile.banner ?? null);
  const [pendingCrop, setPendingCrop] = useState<PendingCrop | null>(null);

  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: profile.name ?? '', username: profile.username ?? '', bio: profile.bio ?? '' },
  });

  function pickFile(e: React.ChangeEvent<HTMLInputElement>, target: 'avatar' | 'banner') {
    const file = e.target.files?.[0];
    if (!file) return;
    setPendingCrop({ src: URL.createObjectURL(file), filename: file.name, aspect: target === 'avatar' ? 1 : 3, target });
    e.target.value = '';
  }

  function onCropComplete(file: File, preview: string) {
    if (pendingCrop?.target === 'avatar') { setAvatarFile(file); setAvatarPreview(preview); }
    else { setBannerFile(file); setBannerPreview(preview); }
    setPendingCrop(null);
  }

  function onSubmit(values: FormValues) {
    mutate({ ...values, avatar: avatarFile, banner: bannerFile }, { onSuccess: () => onClose() });
  }

  return (
    <>
      <div
        style={{ position: 'fixed', inset: 0, zIndex: 50, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(4px)' }}
        onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
      >
        <div style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 16, width: '100%', maxWidth: 480, maxHeight: '90vh', overflowY: 'auto' }}>
          {/* Header */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 20px', borderBottom: '1px solid var(--border)' }}>
            <span style={{ fontWeight: 700, fontSize: 17, color: 'var(--text)', letterSpacing: '-0.02em' }}>Edit profile</span>
            <button type="button" onClick={onClose} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text)', padding: 4, borderRadius: '50%', display: 'flex' }}>
              <X size={20} strokeWidth={2} />
            </button>
          </div>

          {/* Banner picker */}
          <div
            role="button"
            tabIndex={0}
            onClick={() => bannerInputRef.current?.click()}
            onKeyDown={(e) => e.key === 'Enter' && bannerInputRef.current?.click()}
            style={{ position: 'relative', height: 130, cursor: 'pointer', overflow: 'hidden', background: bannerPreview ? undefined : 'linear-gradient(135deg, oklch(28% 0.09 260), oklch(18% 0.05 290))' }}
          >
            {bannerPreview && <img src={bannerPreview} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />}
            <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.35)' }}>
              <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6, color: '#fff' }}>
                <Camera size={22} strokeWidth={1.8} />
                <span style={{ fontSize: 12, fontWeight: 500 }}>Change banner</span>
              </div>
            </div>
            <input ref={bannerInputRef} type="file" accept="image/*" onChange={(e) => pickFile(e, 'banner')} style={{ display: 'none' }} />
          </div>

          <form onSubmit={handleSubmit(onSubmit)} style={{ padding: '0 20px 20px' }}>
            {/* Avatar picker — overlaps banner */}
            <div style={{ display: 'flex', alignItems: 'flex-end', gap: 14, marginBottom: 20, marginTop: -28 }}>
              <div style={{ position: 'relative', flexShrink: 0 }}>
                <Avatar username={profile.name ?? profile.username} avatarUrl={avatarPreview} size="lg" style={{ border: '3px solid var(--bg-card)' }} />
                <button
                  type="button"
                  onClick={() => avatarInputRef.current?.click()}
                  style={{ position: 'absolute', inset: 0, borderRadius: '50%', background: 'rgba(0,0,0,0.4)', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#fff' }}
                >
                  <Camera size={18} strokeWidth={1.8} />
                </button>
                <input ref={avatarInputRef} type="file" accept="image/*" onChange={(e) => pickFile(e, 'avatar')} style={{ display: 'none' }} />
              </div>
              <div style={{ paddingBottom: 4 }}>
                <div style={{ fontSize: 13, color: 'var(--text-tertiary)', fontWeight: 500 }}>Profile photo</div>
                <button type="button" onClick={() => avatarInputRef.current?.click()} style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: 13, color: 'var(--accent)', padding: 0, fontFamily: 'inherit', marginTop: 2 }}>
                  Change photo
                </button>
              </div>
            </div>

            {/* Text fields */}
            {([
              { id: 'name', label: 'Display name', error: errors.name?.message },
              { id: 'username', label: 'Username', error: errors.username?.message },
            ] as const).map(({ id, label, error: fe }) => (
              <div key={id} style={{ marginBottom: 16 }}>
                <label htmlFor={id} style={{ display: 'block', fontSize: 13, color: 'var(--text-tertiary)', marginBottom: 6, fontWeight: 500 }}>{label}</label>
                <input id={id} type="text" {...register(id)} style={{ width: '100%', background: 'var(--bg-subtle)', border: `1px solid ${fe ? 'var(--coral)' : 'var(--border)'}`, borderRadius: 8, padding: '10px 12px', color: 'var(--text)', fontSize: 15, fontFamily: 'inherit', outline: 'none' }} />
                {fe && <div style={{ fontSize: 12, color: 'var(--coral)', marginTop: 4 }}>{fe}</div>}
              </div>
            ))}

            <div style={{ marginBottom: 20 }}>
              <label htmlFor="bio" style={{ display: 'block', fontSize: 13, color: 'var(--text-tertiary)', marginBottom: 6, fontWeight: 500 }}>Bio</label>
              <textarea id="bio" rows={3} {...register('bio')} style={{ width: '100%', background: 'var(--bg-subtle)', border: `1px solid ${errors.bio ? 'var(--coral)' : 'var(--border)'}`, borderRadius: 8, padding: '10px 12px', color: 'var(--text)', fontSize: 15, fontFamily: 'inherit', outline: 'none', resize: 'vertical' }} />
              {errors.bio && <div style={{ fontSize: 12, color: 'var(--coral)', marginTop: 4 }}>{errors.bio.message}</div>}
            </div>

            {error && <div style={{ fontSize: 13, color: 'var(--coral)', marginBottom: 12 }}>{(error as Error).message}</div>}

            <button type="submit" disabled={isPending} style={{ width: '100%', background: 'var(--text)', border: 'none', borderRadius: 10, padding: '11px 0', color: 'var(--bg)', fontWeight: 700, fontSize: 15, cursor: isPending ? 'not-allowed' : 'pointer', fontFamily: 'inherit', opacity: isPending ? 0.6 : 1 }}>
              {isPending ? 'Saving…' : 'Save'}
            </button>
          </form>
        </div>
      </div>

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
