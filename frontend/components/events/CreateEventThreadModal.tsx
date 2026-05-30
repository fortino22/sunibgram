'use client';

import { useState, useRef } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createEventThread } from '@/lib/services/threads';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useAuthStore } from '@/stores/authStore';
import { Avatar } from '@/components/shared/Avatar';
import { toast } from 'sonner';

interface CreateEventThreadModalProps {
  eventId: string;
  eventTitle: string;
  onClose: () => void;
  onSuccess: () => void;
}

export function CreateEventThreadModal({
  eventId,
  eventTitle,
  onClose,
  onSuccess,
}: CreateEventThreadModalProps) {
  const { data: profile } = useMyProfile();
  const token = useAuthStore((s) => s.token);
  const queryClient = useQueryClient();

  const [content, setContent] = useState('');
  const [files, setFiles] = useState<File[]>([]);
  const [previews, setPreviews] = useState<string[]>([]);
  const fileRef = useRef<HTMLInputElement>(null);

  const { mutateAsync, isPending, error } = useMutation({
    mutationFn: () => {
      if (!token) throw new Error('Not authenticated.');
      return createEventThread({ content: content.trim(), files: files.length > 0 ? files : undefined, eventId, isOfficial: true }, token);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['event', eventId] });
    },
  });

  function handleFilesChange(selected: FileList | null) {
    if (!selected) return;
    const arr = Array.from(selected).slice(0, 4);
    setFiles(arr);
    setPreviews(arr.map((f) => URL.createObjectURL(f)));
  }

  function removeFile(i: number) {
    setFiles((prev) => prev.filter((_, idx) => idx !== i));
    setPreviews((prev) => prev.filter((_, idx) => idx !== i));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!content.trim()) return;
    try {
      await mutateAsync();
      onSuccess();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Something went wrong.');
    }
  }

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
    resize: 'vertical',
  };

  return (
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
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
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
          <div>
            <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)' }}>Official Event Update</span>
            <div style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 2 }}>
              {eventTitle} · posted to Official tab
            </div>
          </div>
          <button
            onClick={onClose}
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

        <form onSubmit={handleSubmit} style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 14 }}>
          {/* Avatar + textarea */}
          <div style={{ display: 'flex', gap: 10 }}>
            <Avatar
              username={profile?.username ?? ''}
              avatarUrl={profile?.avatar ?? null}
              size={36}
              style={{ flexShrink: 0, marginTop: 2 }}
            />
            <textarea
              style={{ ...inputStyle, minHeight: 100 }}
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Write an official update for this event…"
              maxLength={500}
              autoFocus
            />
          </div>

          {/* Image previews */}
          {previews.length > 0 && (
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {previews.map((src, i) => (
                <div key={i} style={{ position: 'relative' }}>
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={src}
                    alt=""
                    style={{ width: 80, height: 80, objectFit: 'cover', borderRadius: 8, border: '1px solid var(--border)' }}
                  />
                  <button
                    type="button"
                    onClick={() => removeFile(i)}
                    style={{
                      position: 'absolute',
                      top: -6,
                      right: -6,
                      width: 20,
                      height: 20,
                      borderRadius: '50%',
                      background: 'var(--text)',
                      color: 'var(--bg)',
                      border: 'none',
                      cursor: 'pointer',
                      fontSize: 11,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          )}

          {/* Actions row */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <button
              type="button"
              onClick={() => fileRef.current?.click()}
              style={{
                background: 'none',
                border: '1px solid var(--border)',
                borderRadius: 8,
                padding: '6px 12px',
                fontSize: 13,
                color: 'var(--text-secondary)',
                cursor: 'pointer',
              }}
            >
              + Photo
            </button>
            <input
              ref={fileRef}
              type="file"
              accept="image/*"
              multiple
              style={{ display: 'none' }}
              onChange={(e) => handleFilesChange(e.target.files)}
            />
            <span style={{ fontSize: 11, color: content.length > 480 ? 'var(--error, #e53e3e)' : 'var(--text-tertiary)' }}>
              {content.length}/500
            </span>
          </div>

          {error && (
            <div style={{ fontSize: 13, color: 'var(--error, #e53e3e)' }}>
              {error instanceof Error ? error.message : 'Something went wrong'}
            </div>
          )}

          <button
            type="submit"
            className="btn-solid"
            disabled={isPending || !content.trim()}
            style={{
              width: '100%',
              padding: '12px',
              borderRadius: 9999,
              border: 'none',
              background: 'var(--text)',
              color: 'var(--bg)',
              fontSize: 15,
              fontWeight: 700,
              cursor: isPending || !content.trim() ? 'default' : 'pointer',
              opacity: isPending || !content.trim() ? 0.5 : 1,
            }}
          >
            {isPending ? 'Posting…' : 'Post'}
          </button>
        </form>
      </div>
    </div>
  );
}
