'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { CalendarDays, Clock, ImagePlus, LoaderCircle, MapPin, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import { z } from 'zod';
import { useCreateEvent } from '@/hooks/useCreateEvent';
import { toast } from 'sonner';
import { IMAGE_MAX_BYTES } from '@/lib/compressImage';

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
    banner: z.custom<File | undefined>((v) => v instanceof File, {
      message: 'Banner image is required.',
    }),
    document: z.custom<File | undefined>((v) => v === undefined || v instanceof File),
  })
  .superRefine((v, ctx) => {
    if (v.type === 'official' && v.document === undefined) {
      ctx.addIssue({
        code: 'custom',
        message: 'Supporting document is required for official events.',
        path: ['document'],
      });
    }
    if (v.start_date && v.start_time && v.end_date && v.end_time) {
      const start = new Date(`${v.start_date}T${v.start_time}`);
      const end = new Date(`${v.end_date}T${v.end_time}`);
      if (end <= start) {
        ctx.addIssue({
          code: 'custom',
          message: 'End must be after start.',
          path: ['end_time'],
        });
      }
    }
  });

type FormValues = z.infer<typeof schema>;

function toBackendDatetime(date: string, time: string): string {
  return `${date} ${time}:00`;
}

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateEventDialog({ open, onOpenChange }: Props) {
  const mutation = useCreateEvent();
  const bannerInputRef = useRef<HTMLInputElement>(null);
  const docInputRef = useRef<HTMLInputElement>(null);
  const [bannerPreview, setBannerPreview] = useState<string | null>(null);

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      type: 'unofficial',
      title: '',
      description: '',
      location: '',
      start_date: '',
      start_time: '',
      end_date: '',
      end_time: '',
      banner: undefined,
      document: undefined,
    },
  });

  const type = useWatch({ control: form.control, name: 'type', defaultValue: 'unofficial' });
  const startDate = useWatch({ control: form.control, name: 'start_date', defaultValue: '' });
  const docFile = useWatch({ control: form.control, name: 'document', defaultValue: undefined });

  // Auto-fill end_date when start_date changes and end_date is empty
  useEffect(() => {
    if (startDate && !form.getValues('end_date')) {
      form.setValue('end_date', startDate);
    }
  }, [startDate, form]);

  // Revoke banner preview URL on unmount
  useEffect(() => {
    return () => {
      if (bannerPreview) URL.revokeObjectURL(bannerPreview);
    };
  }, [bannerPreview]);

  function handleBannerChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > IMAGE_MAX_BYTES.raw) {
      toast.error('Banner image exceeds the 5 MB limit. Please use a smaller image.');
      e.target.value = '';
      return;
    }
    if (bannerPreview) URL.revokeObjectURL(bannerPreview);
    setBannerPreview(URL.createObjectURL(file));
    form.setValue('banner', file, { shouldValidate: true });
  }

  function removeBanner() {
    if (bannerPreview) URL.revokeObjectURL(bannerPreview);
    setBannerPreview(null);
    form.setValue('banner', undefined, { shouldValidate: true });
    if (bannerInputRef.current) bannerInputRef.current.value = '';
  }

  async function onSubmit(values: FormValues) {
    if (!values.banner) return;
    await mutation.mutateAsync({
      title: values.title,
      description: values.description,
      location: values.location,
      startTime: toBackendDatetime(values.start_date, values.start_time),
      endTime: toBackendDatetime(values.end_date, values.end_time),
      type: values.type,
      banner: values.banner,
      document: values.document,
    });
    onOpenChange(false);
  }

  if (!open) return null;

  const inputStyle: React.CSSProperties = {
    width: '100%',
    background: 'transparent',
    border: 'none',
    borderBottom: '1px solid var(--border)',
    outline: 'none',
    color: 'var(--text)',
    fontSize: 15,
    lineHeight: 1.6,
    fontFamily: 'inherit',
    padding: '8px 0',
  };

  const fieldWrap: React.CSSProperties = { marginBottom: 18 };
  const labelStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: 6,
    fontSize: 12,
    fontWeight: 600,
    letterSpacing: '0.06em',
    textTransform: 'uppercase',
    color: 'var(--text-tertiary)',
    marginBottom: 6,
  };
  const errorStyle: React.CSSProperties = { fontSize: 12, color: 'var(--coral)', marginTop: 4 };

  return (
    <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center">
      {/* Backdrop */}
      <button
        type="button"
        aria-label="Close"
        className="absolute inset-0"
        onClick={() => onOpenChange(false)}
        style={{ background: 'rgba(2,6,23,0.72)' }}
      />

      {/* Sheet / modal */}
      <div
        className="relative z-10 w-full sm:max-w-2xl overflow-y-auto"
        style={{
          background: 'var(--bg-card)',
          borderTop: '1px solid var(--border)',
          borderRadius: '28px 28px 0 0',
          maxHeight: '92dvh',
          padding: '0 0 env(safe-area-inset-bottom, 16px)',
        }}
        // sm: proper modal
      >
        {/* Drag handle (mobile) */}
        <div className="sm:hidden flex justify-center pt-3 pb-1">
          <div className="h-1 w-10 rounded-full" style={{ background: 'var(--border)' }} />
        </div>

        {/* Header */}
        <div
          className="flex items-center justify-between px-5 py-4"
          style={{ borderBottom: '1px solid var(--border)' }}
        >
          <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)' }}>
            Create Event
          </span>
          <button
            type="button"
            onClick={() => onOpenChange(false)}
            className="flex items-center justify-center rounded-full"
            style={{
              width: 32,
              height: 32,
              background: 'var(--bg-subtle)',
              color: 'var(--text-secondary)',
            }}
          >
            <X size={16} />
          </button>
        </div>

        <form
          onSubmit={form.handleSubmit((v) => void onSubmit(v))}
          style={{ padding: '20px 20px 8px' }}
        >
          {/* Type toggle */}
          <div style={fieldWrap}>
            <p style={labelStyle}>Event type</p>
            <div style={{ display: 'flex', gap: 8 }}>
              {(['unofficial', 'official'] as const).map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => form.setValue('type', t)}
                  style={{
                    padding: '7px 18px',
                    borderRadius: 999,
                    fontSize: 14,
                    fontWeight: 600,
                    cursor: 'pointer',
                    transition: 'all 0.15s',
                    background: type === t ? 'var(--accent)' : 'transparent',
                    color: type === t ? '#fff' : 'var(--text-secondary)',
                    border: type === t ? '1.5px solid var(--accent)' : '1.5px solid var(--border)',
                  }}
                >
                  {t === 'official' ? 'Official' : 'Unofficial'}
                </button>
              ))}
            </div>
            {type === 'official' && (
              <p style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 8 }}>
                Official events require an approval document from the university.
              </p>
            )}
          </div>

          {/* Title */}
          <div style={fieldWrap}>
            <input
              style={{ ...inputStyle, fontSize: 20, fontWeight: 700 }}
              placeholder="Event title..."
              {...form.register('title')}
            />
            {form.formState.errors.title && (
              <p style={errorStyle}>{form.formState.errors.title.message}</p>
            )}
          </div>

          {/* Description */}
          <div style={fieldWrap}>
            <textarea
              rows={4}
              style={{ ...inputStyle, resize: 'vertical', minHeight: 88 }}
              placeholder="What should people know before they show up?"
              {...form.register('description')}
            />
            {form.formState.errors.description && (
              <p style={errorStyle}>{form.formState.errors.description.message}</p>
            )}
          </div>

          {/* Location */}
          <div style={fieldWrap}>
            <p style={labelStyle}><MapPin size={12} /> Location</p>
            <input
              style={inputStyle}
              placeholder="Binus Anggrek, Room 412"
              {...form.register('location')}
            />
            {form.formState.errors.location && (
              <p style={errorStyle}>{form.formState.errors.location.message}</p>
            )}
          </div>

          {/* Date + Time */}
          <div style={{ ...fieldWrap, display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <div>
              <p style={labelStyle}><CalendarDays size={12} /> Start date</p>
              <input type="date" style={inputStyle} {...form.register('start_date')} />
              {form.formState.errors.start_date && (
                <p style={errorStyle}>{form.formState.errors.start_date.message}</p>
              )}
            </div>
            <div>
              <p style={labelStyle}><Clock size={12} /> Start time</p>
              <input type="time" style={inputStyle} {...form.register('start_time')} />
              {form.formState.errors.start_time && (
                <p style={errorStyle}>{form.formState.errors.start_time.message}</p>
              )}
            </div>
            <div>
              <p style={labelStyle}><CalendarDays size={12} /> End date</p>
              <input type="date" style={inputStyle} {...form.register('end_date')} />
              {form.formState.errors.end_date && (
                <p style={errorStyle}>{form.formState.errors.end_date.message}</p>
              )}
            </div>
            <div>
              <p style={labelStyle}><Clock size={12} /> End time</p>
              <input type="time" style={inputStyle} {...form.register('end_time')} />
              {form.formState.errors.end_time && (
                <p style={errorStyle}>{form.formState.errors.end_time.message}</p>
              )}
            </div>
          </div>

          {/* Banner */}
          <div style={fieldWrap}>
            <p style={labelStyle}><ImagePlus size={12} /> Event banner</p>
            <input
              ref={bannerInputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={handleBannerChange}
            />
            {bannerPreview ? (
              <div
                style={{
                  position: 'relative',
                  borderRadius: 16,
                  overflow: 'hidden',
                  border: '1px solid var(--border)',
                }}
              >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={bannerPreview}
                  alt="Banner preview"
                  style={{ width: '100%', maxHeight: 220, objectFit: 'cover', display: 'block' }}
                />
                <button
                  type="button"
                  onClick={removeBanner}
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
              </div>
            ) : (
              <button
                type="button"
                onClick={() => bannerInputRef.current?.click()}
                className="flex w-full items-center justify-center gap-2"
                style={{
                  height: 120,
                  borderRadius: 16,
                  border: '1.5px dashed var(--border)',
                  background: 'var(--bg-subtle)',
                  color: 'var(--text-tertiary)',
                  fontSize: 14,
                  cursor: 'pointer',
                }}
              >
                <ImagePlus size={20} />
                Upload banner image
              </button>
            )}
            {form.formState.errors.banner && (
              <p style={errorStyle}>{form.formState.errors.banner.message}</p>
            )}
          </div>

          {/* Supporting document — official only */}
          {type === 'official' && (
            <div style={fieldWrap}>
              <p style={labelStyle}>📄 Supporting document (PDF)</p>
              <input
                ref={docInputRef}
                type="file"
                accept="application/pdf"
                className="hidden"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  form.setValue('document', file, { shouldValidate: true });
                }}
              />
              <button
                type="button"
                onClick={() => docInputRef.current?.click()}
                className="flex w-full items-center justify-center gap-2"
                style={{
                  height: 80,
                  borderRadius: 16,
                  border: '1.5px dashed var(--border)',
                  background: 'var(--bg-subtle)',
                  color: docFile ? 'var(--accent)' : 'var(--text-tertiary)',
                  fontSize: 14,
                  cursor: 'pointer',
                }}
              >
                {docFile ? `✓ ${docFile.name}` : 'Upload approval document'}
              </button>
              {form.formState.errors.document && (
                <p style={errorStyle}>{form.formState.errors.document.message}</p>
              )}
            </div>
          )}

          {mutation.isError && (
            <div
              style={{
                marginBottom: 16,
                padding: '10px 14px',
                borderRadius: 12,
                background: 'var(--coral-soft)',
                color: 'var(--coral)',
                fontSize: 14,
              }}
            >
              {mutation.error.message}
            </div>
          )}

          {/* Footer actions */}
          <div
            style={{
              display: 'flex',
              justifyContent: 'flex-end',
              gap: 10,
              paddingTop: 8,
              paddingBottom: 8,
            }}
          >
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              style={{
                padding: '9px 20px',
                borderRadius: 999,
                border: '1.5px solid var(--border)',
                background: 'transparent',
                color: 'var(--text-secondary)',
                fontSize: 14,
                fontWeight: 600,
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={mutation.isPending}
              style={{
                padding: '9px 24px',
                borderRadius: 999,
                border: 'none',
                background: 'var(--accent)',
                color: '#fff',
                fontSize: 14,
                fontWeight: 700,
                cursor: mutation.isPending ? 'not-allowed' : 'pointer',
                opacity: mutation.isPending ? 0.7 : 1,
                display: 'flex',
                alignItems: 'center',
                gap: 8,
              }}
            >
              {mutation.isPending && <LoaderCircle size={15} className="animate-spin" />}
              {mutation.isPending ? 'Creating...' : 'Create Event'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
