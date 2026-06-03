'use client';

import { useEffect, useRef, useState } from 'react';
import { LoaderCircle, X, MapPin, CalendarDays, Clock, Users } from 'lucide-react';
import { toast } from 'sonner';
import { IMAGE_MAX_BYTES } from '@/lib/compressImage';
import { ImageCropModal } from '@/components/profile/ImageCropModal';
import { InviteePicker } from './InviteePicker';
import { useCreateCommunityEvent, useUpdateCommunityEvent } from '@/hooks/useCommunityEventMutations';
import type { CommunityEventDetail, InviteeCandidate } from '@/lib/services/communityEvents';

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  communityId: number;
  /** Provided when editing an existing event. */
  event?: CommunityEventDetail;
}

// backend "2026-06-10 14:30:00" -> datetime-local "2026-06-10T14:30"
function toLocalInput(value?: string): string {
  if (!value) return '';
  return value.slice(0, 16).replace(' ', 'T');
}

function revokeBlob(url: string | null) {
  if (url && url.startsWith('blob:')) URL.revokeObjectURL(url);
}

export function CreateCommunityEventDialog({ open, onOpenChange, communityId, event }: Props) {
  const isEdit = !!event;
  const [step, setStep] = useState<1 | 2>(1);
  const [title, setTitle] = useState(event?.title ?? '');
  const [description, setDescription] = useState(event?.description ?? '');
  const [location, setLocation] = useState(event?.location ?? '');
  const [startTime, setStartTime] = useState(toLocalInput(event?.start_time));
  const [endTime, setEndTime] = useState(toLocalInput(event?.end_time));
  const [selected, setSelected] = useState<InviteeCandidate[]>(
    event?.invitees.map((i) => ({ user_id: i.user_id, username: i.username, name: i.name, avatar_url: i.avatar_url })) ?? []
  );
  const [banner, setBanner] = useState<File | null>(null);
  const [bannerPreview, setBannerPreview] = useState<string | null>(event?.banner_url ?? null);
  const [bannerCleared, setBannerCleared] = useState(false);
  const [pendingCrop, setPendingCrop] = useState<{ src: string; filename: string } | null>(null);
  const bannerInputRef = useRef<HTMLInputElement>(null);

  const createMutation = useCreateCommunityEvent(communityId);
  const updateMutation = useUpdateCommunityEvent(event?.id ?? 0, communityId);
  const isPending = createMutation.isPending || updateMutation.isPending;

  useEffect(() => () => revokeBlob(bannerPreview), [bannerPreview]);

  function handleClose() {
    onOpenChange(false);
  }

  function handleBannerChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > IMAGE_MAX_BYTES.raw) {
      toast.error('Banner image exceeds the 5 MB limit. Please use a smaller image.');
      e.target.value = '';
      return;
    }
    setPendingCrop({ src: URL.createObjectURL(file), filename: file.name });
    e.target.value = '';
  }

  function handleCropComplete(file: File, preview: string) {
    revokeBlob(bannerPreview);
    setBannerPreview(preview);
    setBanner(file);
    setBannerCleared(false);
    if (pendingCrop) URL.revokeObjectURL(pendingCrop.src);
    setPendingCrop(null);
  }

  function handleCropClose() {
    if (pendingCrop) URL.revokeObjectURL(pendingCrop.src);
    setPendingCrop(null);
  }

  function clearBanner() {
    revokeBlob(bannerPreview);
    setBanner(null);
    setBannerPreview(null);
    setBannerCleared(true);
  }

  function goToStep2() {
    if (!title.trim()) return toast.error('Title is required.');
    setStep(2);
  }

  async function handleSave() {
    if (!title.trim()) return toast.error('Title is required.');
    if (!startTime || !endTime) return toast.error('Start and end time are required.');
    if (new Date(endTime) <= new Date(startTime)) return toast.error('End must be after start.');

    const params = {
      title: title.trim(),
      description: description.trim(),
      location: location.trim(),
      startTime,
      endTime,
      banner,
      clearBanner: isEdit ? bannerCleared : false,
      inviteeUserIds: selected.map((s) => s.user_id),
    };

    try {
      if (isEdit) {
        await updateMutation.mutateAsync(params);
        toast.success('Event updated.');
      } else {
        await createMutation.mutateAsync(params);
        toast.success('Event created.');
      }
      onOpenChange(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Something went wrong.');
    }
  }

  if (!open) return null;

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '10px 12px',
    borderRadius: 8,
    border: '1px solid var(--border)',
    background: 'var(--bg-subtle)',
    color: 'var(--text)',
    fontSize: 14,
    boxSizing: 'border-box',
    fontFamily: 'inherit',
    outline: 'none',
  };

  const labelStyle: React.CSSProperties = {
    fontSize: 12,
    fontWeight: 600,
    color: 'var(--text-secondary)',
    display: 'flex',
    alignItems: 'center',
    gap: 5,
    marginBottom: 6,
  };

  return (
    <>
      <div
        style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0,0,0,0.6)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 100,
          padding: 16,
        }}
        onClick={(e) => { if (e.target === e.currentTarget) handleClose(); }}
      >
        <div
          style={{
            background: 'var(--bg-card)',
            borderRadius: 16,
            width: '100%',
            maxWidth: 480,
            maxHeight: '90vh',
            overflowY: 'auto',
            border: '1px solid var(--border)',
          }}
        >
          {/* Header */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '16px 20px',
              borderBottom: '1px solid var(--border)',
              position: 'sticky',
              top: 0,
              background: 'var(--bg-card)',
              zIndex: 1,
            }}
          >
            <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)' }}>
                {isEdit ? 'Edit event' : 'Create community event'}
              </span>
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>
                STEP {step} OF 2 · {step === 1 ? 'Details' : 'People'}
              </span>
            </div>
            <button
              type="button"
              onClick={handleClose}
              style={{ background: 'none', border: 'none', color: 'var(--text-tertiary)', fontSize: 20, cursor: 'pointer', lineHeight: 1 }}
            >
              ✕
            </button>
          </div>

          {/* Banner — step 1 */}
          {step === 1 && (
            <div
              onClick={() => bannerInputRef.current?.click()}
              style={{
                aspectRatio: '16 / 9',
                width: '100%',
                background: bannerPreview ? `url(${bannerPreview}) center/cover no-repeat` : 'var(--bg-subtle)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                position: 'relative',
              }}
            >
              {!bannerPreview && (
                <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Click to add a banner</span>
              )}
              {bannerPreview && (
                <button
                  type="button"
                  onClick={(e) => { e.stopPropagation(); clearBanner(); }}
                  style={{
                    position: 'absolute',
                    top: 8,
                    right: 8,
                    width: 28,
                    height: 28,
                    borderRadius: '50%',
                    border: '1px solid rgba(255,255,255,0.22)',
                    background: 'rgba(0,0,0,0.55)',
                    color: '#fff',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    cursor: 'pointer',
                  }}
                >
                  <X size={14} />
                </button>
              )}
              <input
                ref={bannerInputRef}
                type="file"
                accept="image/*"
                style={{ display: 'none' }}
                onChange={handleBannerChange}
              />
            </div>
          )}

          <div style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 14 }}>
            {step === 1 && (
              <>
                <div>
                  <label style={labelStyle}>Title *</label>
                  <input style={inputStyle} value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Event title" maxLength={255} />
                </div>
                <div>
                  <label style={labelStyle}>Description</label>
                  <textarea
                    rows={3}
                    style={{ ...inputStyle, resize: 'vertical', minHeight: 72 }}
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="What's this event about?"
                    maxLength={2000}
                  />
                </div>
              </>
            )}

            {step === 2 && (
              <>
                <div>
                  <label style={labelStyle}><MapPin size={12} /> Location</label>
                  <input style={inputStyle} value={location} onChange={(e) => setLocation(e.target.value)} placeholder="e.g. Auditorium, Room 301, Online" maxLength={255} />
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
                  <div>
                    <label style={labelStyle}><CalendarDays size={12} /> Starts *</label>
                    <input type="datetime-local" style={inputStyle} value={startTime} onChange={(e) => setStartTime(e.target.value)} />
                  </div>
                  <div>
                    <label style={labelStyle}><Clock size={12} /> Ends *</label>
                    <input type="datetime-local" style={inputStyle} value={endTime} onChange={(e) => setEndTime(e.target.value)} />
                  </div>
                </div>

                <div>
                  <label style={labelStyle}><Users size={12} /> Tag people</label>
                  <InviteePicker communityId={communityId} selected={selected} onChange={setSelected} />
                </div>
              </>
            )}

            {/* Footer nav */}
            <div style={{ display: 'flex', gap: 10 }}>
              {step === 2 && (
                <button
                  type="button"
                  onClick={() => setStep(1)}
                  style={{
                    flex: '0 0 auto',
                    padding: '12px 18px',
                    borderRadius: 999,
                    border: '1px solid var(--border)',
                    background: 'transparent',
                    color: 'var(--text)',
                    fontSize: 15,
                    fontWeight: 700,
                    cursor: 'pointer',
                  }}
                >
                  ← Back
                </button>
              )}
              {step === 1 ? (
                <button
                  type="button"
                  onClick={goToStep2}
                  style={{
                    width: '100%',
                    padding: '12px',
                    borderRadius: 999,
                    border: 'none',
                    background: 'var(--text)',
                    color: 'var(--bg)',
                    fontSize: 15,
                    fontWeight: 700,
                    cursor: 'pointer',
                  }}
                >
                  Next →
                </button>
              ) : (
                <button
                  type="button"
                  onClick={handleSave}
                  disabled={isPending}
                  style={{
                    flex: 1,
                    padding: '12px',
                    borderRadius: 999,
                    border: 'none',
                    background: 'var(--text)',
                    color: 'var(--bg)',
                    fontSize: 15,
                    fontWeight: 700,
                    cursor: isPending ? 'default' : 'pointer',
                    opacity: isPending ? 0.5 : 1,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: 8,
                  }}
                >
                  {isPending && <LoaderCircle size={15} className="animate-spin" />}
                  {isPending ? (isEdit ? 'Saving…' : 'Creating…') : isEdit ? 'Save changes' : 'Create event'}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {pendingCrop && (
        <ImageCropModal
          imageSrc={pendingCrop.src}
          filename={pendingCrop.filename}
          aspect={16 / 9}
          maxSizeBytes={IMAGE_MAX_BYTES.banner}
          onComplete={handleCropComplete}
          onClose={handleCropClose}
        />
      )}
    </>
  );
}
