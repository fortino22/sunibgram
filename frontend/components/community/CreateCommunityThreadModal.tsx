'use client';

import { useState, useRef } from 'react';
import { useCreateCommunityThread } from '@/hooks/useCommunityMutations';
import { useMyProfile } from '@/hooks/useMyProfile';
import { Avatar } from '@/components/shared/Avatar';

interface CreateCommunityThreadModalProps {
  communityId: number;
  communityName: string;
  onClose: () => void;
  onSuccess: () => void;
}

export function CreateCommunityThreadModal({
  communityId,
  communityName,
  onClose,
  onSuccess,
}: CreateCommunityThreadModalProps) {
  const { data: profile } = useMyProfile();
  const { mutateAsync, isPending, error } = useCreateCommunityThread(communityId);

  const [content, setContent] = useState('');
  const [files, setFiles] = useState<File[]>([]);
  const [previews, setPreviews] = useState<string[]>([]);
  const fileRef = useRef<HTMLInputElement>(null);

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
      await mutateAsync({ content: content.trim(), files: files.length > 0 ? files : undefined });
      onSuccess();
    } catch {
      // error shown via mutation error state
    }
  }

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'center',
        zIndex: 100,
        padding: '60px 16px 16px',
      }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          borderRadius: 16,
          width: '100%',
          maxWidth: 540,
          border: '1px solid var(--border)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '14px 16px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)' }}>
            Post in {communityName}
          </span>
          <button
            onClick={onClose}
            style={{ background: 'none', border: 'none', color: 'var(--text-tertiary)', fontSize: 20, cursor: 'pointer' }}
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} style={{ padding: '14px 16px' }}>
          <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
            <Avatar
              username={profile?.username ?? 'me'}
              avatarUrl={profile?.avatar ?? null}
              size="md"
            />
            <div style={{ flex: 1 }}>
              <textarea
                autoFocus
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder={`Share something with ${communityName}…`}
                style={{
                  width: '100%',
                  minHeight: 100,
                  background: 'transparent',
                  border: 'none',
                  outline: 'none',
                  color: 'var(--text)',
                  fontSize: 15,
                  resize: 'none',
                  boxSizing: 'border-box',
                }}
              />

              {/* Image previews */}
              {previews.length > 0 && (
                <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', marginTop: 8 }}>
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
                          top: 2,
                          right: 2,
                          width: 20,
                          height: 20,
                          borderRadius: '50%',
                          background: 'rgba(0,0,0,0.6)',
                          border: 'none',
                          color: '#fff',
                          fontSize: 12,
                          cursor: 'pointer',
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
            </div>
          </div>

          {error && (
            <div style={{ fontSize: 13, color: 'var(--error, #e53e3e)', marginTop: 8, paddingLeft: 52 }}>
              {error instanceof Error ? error.message : 'Something went wrong'}
            </div>
          )}

          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              marginTop: 12,
              paddingTop: 12,
              borderTop: '1px solid var(--border)',
            }}
          >
            <button
              type="button"
              onClick={() => fileRef.current?.click()}
              style={{
                background: 'none',
                border: 'none',
                color: 'var(--accent)',
                fontSize: 20,
                cursor: 'pointer',
                padding: 4,
              }}
            >
              🖼
            </button>
            <input
              ref={fileRef}
              type="file"
              accept="image/*,video/*"
              multiple
              style={{ display: 'none' }}
              onChange={(e) => handleFilesChange(e.target.files)}
            />

            <button
              type="submit"
              disabled={isPending || !content.trim()}
              style={{
                padding: '8px 20px',
                borderRadius: 999,
                border: 'none',
                background: 'var(--text)',
                color: 'var(--bg)',
                fontSize: 14,
                fontWeight: 700,
                cursor: isPending || !content.trim() ? 'default' : 'pointer',
                opacity: isPending || !content.trim() ? 0.5 : 1,
              }}
            >
              {isPending ? 'Posting…' : 'Post'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
