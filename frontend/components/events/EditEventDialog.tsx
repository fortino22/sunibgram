'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { CalendarDays, Clock, LoaderCircle, MapPin, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import { z } from 'zod';
import { useUpdateEvent } from '@/hooks/useUpdateEvent';
import { useFacultyLocations } from '@/hooks/useFacultyLocations';
import { toast } from 'sonner';
import { IMAGE_MAX_BYTES } from '@/lib/compressImage';
import type { EventDetail } from '@/lib/services/events';
import { ImageCropModal } from '@/components/profile/ImageCropModal';

const schema = z
  .object({
    type: z.enum(['official', 'unofficial']),
    title: z.string().trim().min(1, 'Title is required.'),
    description: z.string().trim().min(1, 'Description is required.'),
    location: z.string().trim().min(1, 'Location is required.'),
    start_date: z.string().min(1, 'Start date is required.'),
    start_time: z.string().min(1, 'Start time is required.'),
    end_date: z.string().min(1, 'End date is required.'),
    end_time: z.string().min(1, 'End time is required.'),
    banner: z.custom<File | undefined>((v) => v === undefined || v instanceof File),
    document: z.custom<File | undefined>((v) => v === undefined || v instanceof File),
    facultyLocationIds: z.array(z.number()).min(1, 'Pick at least one audience region.'),
  })
  .superRefine((v, ctx) => {
    if (v.start_date && v.start_time && v.end_date && v.end_time) {
      const start = new Date(`${v.start_date}T${v.start_time}`);
      const end = new Date(`${v.end_date}T${v.end_time}`);
      if (end <= start) {
        ctx.addIssue({ code: 'custom', message: 'End must be after start.', path: ['end_time'] });
      }
    }
  });

type FormValues = z.infer<typeof schema>;

function toBackendDatetime(date: string, time: string): string {
  return `${date} ${time}:00`;
}

function parseIsoToDateParts(iso: string | null): { date: string; time: string } {
  if (!iso) return { date: '', time: '' };
  const s = iso.replace(' ', 'T');
  return { date: s.slice(0, 10), time: s.slice(11, 16) };
}

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  event: EventDetail;
}

export function EditEventDialog({ open, onOpenChange, event }: Props) {
  const mutation = useUpdateEvent(event.id);
  const bannerInputRef = useRef<HTMLInputElement>(null);
  const docInputRef = useRef<HTMLInputElement>(null);

  const start = parseIsoToDateParts(event.startTime);
  const end = parseIsoToDateParts(event.endTime);

  const [bannerPreview, setBannerPreview] = useState<string | null>(event.banner ?? null);
  const [bannerIsNew, setBannerIsNew] = useState(false);
  const [pendingCrop, setPendingCrop] = useState<{ src: string; filename: string } | null>(null);
  const [step, setStep] = useState<1 | 2>(1);

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      type: event.type,
      title: event.title,
      description: event.description,
      location: event.location,
      start_date: start.date,
      start_time: start.time,
      end_date: end.date,
      end_time: end.time,
      banner: undefined,
      document: undefined,
      facultyLocationIds: (event.audience ?? []).map((a) => a.id),
    },
  });

  const { data: facultyLocations = [], isLoading: locationsLoading } = useFacultyLocations(true);
  const allRegionsId = facultyLocations.find((l) => l.isAllRegions)?.id ?? null;

  const type = useWatch({ control: form.control, name: 'type', defaultValue: event.type });
  const startDate = useWatch({ control: form.control, name: 'start_date', defaultValue: start.date });
  const docFile = useWatch({ control: form.control, name: 'document', defaultValue: undefined });
  const selectedAudience = useWatch({ control: form.control, name: 'facultyLocationIds', defaultValue: [] });

  function toggleAudience(id: number) {
    const current = form.getValues('facultyLocationIds') ?? [];
    if (allRegionsId !== null && id === allRegionsId) {
      form.setValue('facultyLocationIds', current.includes(id) ? [] : [id], { shouldValidate: true });
      return;
    }
    const withoutAll = current.filter((x) => x !== allRegionsId);
    const next = withoutAll.includes(id) ? withoutAll.filter((x) => x !== id) : [...withoutAll, id];
    form.setValue('facultyLocationIds', next, { shouldValidate: true });
  }

  useEffect(() => {
    if (startDate && !form.getValues('end_date')) {
      form.setValue('end_date', startDate);
    }
  }, [startDate, form]);

  useEffect(() => {
    return () => {
      if (bannerIsNew && bannerPreview) URL.revokeObjectURL(bannerPreview);
    };
  }, [bannerIsNew, bannerPreview]);

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
    if (bannerIsNew && bannerPreview) URL.revokeObjectURL(bannerPreview);
    setBannerPreview(preview);
    setBannerIsNew(true);
    form.setValue('banner', file, { shouldValidate: true });
    if (pendingCrop) URL.revokeObjectURL(pendingCrop.src);
    setPendingCrop(null);
  }

  function handleCropClose() {
    if (pendingCrop) URL.revokeObjectURL(pendingCrop.src);
    setPendingCrop(null);
  }

  async function onSubmit(values: FormValues) {
    await mutation.mutateAsync({
      title: values.title,
      description: values.description,
      location: values.location,
      startTime: toBackendDatetime(values.start_date, values.start_time),
      endTime: toBackendDatetime(values.end_date, values.end_time),
      type: values.type,
      banner: values.banner,
      document: values.document,
      facultyLocationIds: values.facultyLocationIds,
    });
    form.reset();
    if (bannerIsNew && bannerPreview) URL.revokeObjectURL(bannerPreview);
    setBannerPreview(event.banner ?? null);
    setBannerIsNew(false);
    setStep(1);
    onOpenChange(false);
  }

  const submit = form.handleSubmit((v) => void onSubmit(v));

  async function goToStep2() {
    const ok = await form.trigger(['type', 'title', 'description', 'banner']);
    if (!ok) return;
    form.clearErrors(['location', 'start_date', 'start_time', 'end_date', 'end_time', 'facultyLocationIds', 'document']);
    setStep(2);
  }

  if (!open) return null;
  if (event.status !== 'revision') return null;

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
    display: 'block',
    marginBottom: 6,
  };

  const errorStyle: React.CSSProperties = { fontSize: 12, color: 'var(--coral, #e53e3e)', marginTop: 4 };

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
              Edit Event
            </span>
            <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>
              STEP {step} OF 2 · {step === 1 ? 'Details' : 'Logistics'} · Saving resubmits for admin review.
            </span>
          </div>
          <button
            type="button"
            onClick={() => handleClose()}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--text-tertiary)',
              fontSize: 20,
              cursor: 'pointer',
              lineHeight: 1,
            }}
          >
            ✕
          </button>
        </div>

        {/* Banner */}
        {step === 1 && (
        <div
          onClick={() => bannerInputRef.current?.click()}
          style={{
            aspectRatio: '2.31',
            width: '100%',
            background: bannerPreview
              ? `url(${bannerPreview}) center/cover no-repeat`
              : 'var(--bg-subtle)',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            position: 'relative',
          }}
        >
          {!bannerPreview && (
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Click to change banner</span>
          )}
          {bannerPreview && (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                if (bannerIsNew && bannerPreview) URL.revokeObjectURL(bannerPreview);
                setBannerPreview(null);
                setBannerIsNew(false);
                form.setValue('banner', undefined, { shouldValidate: true });
              }}
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

        <form
          onSubmit={(e) => { e.preventDefault(); if (step === 2) submit(); }}
          style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 14 }}
        >
          {step === 1 && (
          <>
          {/* Type toggle */}
          <div>
            <label style={labelStyle}>Event type</label>
            <div style={{ display: 'flex', gap: 8 }}>
              {(['unofficial', 'official'] as const).map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => form.setValue('type', t)}
                  style={{
                    flex: 1,
                    padding: '8px',
                    borderRadius: 8,
                    border: `1px solid ${type === t ? 'var(--text)' : 'var(--border)'}`,
                    background: type === t ? 'var(--text)' : 'transparent',
                    color: type === t ? 'var(--bg)' : 'var(--text)',
                    fontSize: 13,
                    fontWeight: 600,
                    cursor: 'pointer',
                  }}
                >
                  {t === 'official' ? 'Official' : 'Unofficial'}
                </button>
              ))}
            </div>
          </div>

          {/* Title */}
          <div>
            <label style={labelStyle}>Title *</label>
            <input style={inputStyle} placeholder="Event title" {...form.register('title')} />
            {form.formState.errors.title && (
              <p style={errorStyle}>{form.formState.errors.title.message}</p>
            )}
          </div>

          {/* Description */}
          <div>
            <label style={labelStyle}>Description *</label>
            <textarea
              rows={3}
              style={{ ...inputStyle, resize: 'vertical', minHeight: 72 }}
              placeholder="What should people know before they show up?"
              {...form.register('description')}
            />
            {form.formState.errors.description && (
              <p style={errorStyle}>{form.formState.errors.description.message}</p>
            )}
          </div>
          </>
          )}

          {step === 2 && (
          <>
          {/* Location */}
          <div>
            <label style={{ ...labelStyle, display: 'flex', alignItems: 'center', gap: 5 }}>
              <MapPin size={12} /> Location *
            </label>
            <input style={inputStyle} placeholder="Binus Anggrek, Room 412" {...form.register('location')} />
            {form.formState.errors.location && (
              <p style={errorStyle}>{form.formState.errors.location.message}</p>
            )}
          </div>

          {/* Date + Time grid */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
            <div>
              <label style={{ ...labelStyle, display: 'flex', alignItems: 'center', gap: 5 }}>
                <CalendarDays size={12} /> Start date
              </label>
              <input type="date" style={inputStyle} {...form.register('start_date')} />
              {form.formState.errors.start_date && (
                <p style={errorStyle}>{form.formState.errors.start_date.message}</p>
              )}
            </div>
            <div>
              <label style={{ ...labelStyle, display: 'flex', alignItems: 'center', gap: 5 }}>
                <Clock size={12} /> Start time
              </label>
              <input type="time" style={inputStyle} {...form.register('start_time')} />
              {form.formState.errors.start_time && (
                <p style={errorStyle}>{form.formState.errors.start_time.message}</p>
              )}
            </div>
            <div>
              <label style={{ ...labelStyle, display: 'flex', alignItems: 'center', gap: 5 }}>
                <CalendarDays size={12} /> End date
              </label>
              <input type="date" style={inputStyle} {...form.register('end_date')} />
              {form.formState.errors.end_date && (
                <p style={errorStyle}>{form.formState.errors.end_date.message}</p>
              )}
            </div>
            <div>
              <label style={{ ...labelStyle, display: 'flex', alignItems: 'center', gap: 5 }}>
                <Clock size={12} /> End time
              </label>
              <input type="time" style={inputStyle} {...form.register('end_time')} />
              {form.formState.errors.end_time && (
                <p style={errorStyle}>{form.formState.errors.end_time.message}</p>
              )}
            </div>
          </div>

          {/* Audience regions */}
          <div>
            <label style={labelStyle}>Audience regions *</label>
            {locationsLoading ? (
              <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Loading campuses…</div>
            ) : (
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {facultyLocations.map((loc) => {
                  const selected = selectedAudience.includes(loc.id);
                  return (
                    <button
                      key={loc.id}
                      type="button"
                      onClick={() => toggleAudience(loc.id)}
                      style={{
                        padding: '7px 14px', borderRadius: 999,
                        border: `1px solid ${selected ? 'var(--text)' : 'var(--border)'}`,
                        background: selected ? 'var(--text)' : 'transparent',
                        color: selected ? 'var(--bg)' : 'var(--text)',
                        fontSize: 13, fontWeight: 600, cursor: 'pointer',
                      }}
                    >
                      {loc.name}
                    </button>
                  );
                })}
              </div>
            )}
            {form.formState.errors.facultyLocationIds && (
              <p style={errorStyle}>{form.formState.errors.facultyLocationIds.message}</p>
            )}
          </div>

          {/* Docs */}
          <div>
            <label style={labelStyle}>Docs <span style={{ fontWeight: 400, color: 'var(--text-tertiary)' }}>(.doc / .docx)</span></label>
            <input
              ref={docInputRef}
              type="file"
              accept=".doc,.docx,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
              style={{ display: 'none' }}
              onChange={(e) => {
                const file = e.target.files?.[0];
                form.setValue('document', file, { shouldValidate: true });
              }}
            />
            <button
              type="button"
              onClick={() => docInputRef.current?.click()}
              style={{
                width: '100%',
                padding: '10px 12px',
                borderRadius: 8,
                border: '1px dashed var(--border)',
                background: 'var(--bg-subtle)',
                color: docFile ? 'var(--text)' : 'var(--text-tertiary)',
                fontSize: 13,
                fontWeight: docFile ? 600 : 400,
                cursor: 'pointer',
                textAlign: 'center',
                boxSizing: 'border-box',
              }}
            >
              {docFile ? docFile.name : event.documentUrl ? 'Replace document (.doc / .docx)' : 'Upload document (.doc / .docx)'}
            </button>
            {form.formState.errors.document && (
              <p style={errorStyle}>{form.formState.errors.document.message}</p>
            )}
          </div>

          {mutation.isError && (
            <div style={{ fontSize: 13, color: 'var(--coral, #e53e3e)' }}>
              {mutation.error.message}
            </div>
          )}
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
                onClick={() => void goToStep2()}
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
                type="submit"
                disabled={mutation.isPending}
                style={{
                  flex: 1,
                  padding: '12px',
                  borderRadius: 999,
                  border: 'none',
                  background: 'var(--text)',
                  color: 'var(--bg)',
                  fontSize: 15,
                  fontWeight: 700,
                  cursor: mutation.isPending ? 'default' : 'pointer',
                  opacity: mutation.isPending ? 0.5 : 1,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 8,
                }}
              >
                {mutation.isPending && <LoaderCircle size={15} className="animate-spin" />}
                {mutation.isPending ? 'Saving…' : 'Save Changes'}
              </button>
            )}
          </div>
        </form>
      </div>
    </div>
    {pendingCrop && (
      <ImageCropModal
        imageSrc={pendingCrop.src}
        filename={pendingCrop.filename}
        aspect={2.31}
        maxSizeBytes={IMAGE_MAX_BYTES.banner}
        onComplete={handleCropComplete}
        onClose={handleCropClose}
      />
    )}
    </>
  );
}
