'use client';

import { ImagePlus, LoaderCircle, X } from 'lucide-react';
import {
  type ChangeEvent,
  type FormEvent,
  useEffect,
  useRef,
  useState,
} from 'react';
import { useCreateThread } from '@/hooks/thread/useCreateThread';
import { useAuthStore } from '@/stores/authStore';
import { useMyProfile } from '@/hooks/useMyProfile';
import { Avatar } from './Avatar';

type PreviewKind = 'image' | 'video' | 'file';

interface FilePreview {
  file: File;
  kind: PreviewKind;
  url: string;
}

function getPreviewKind(file: File): PreviewKind {
  if (file.type.startsWith('image/')) {
    return 'image';
  }

  if (file.type.startsWith('video/')) {
    return 'video';
  }

  return 'file';
}

function isSameFile(left: File, right: File): boolean {
  return (
    left.name === right.name &&
    left.lastModified === right.lastModified &&
    left.size === right.size
  );
}

export function CreatePostBox() {
  const user = useAuthStore((state) => state.user);
  const token = useAuthStore((state) => state.token);
  const { data: profile } = useMyProfile();
  const createThreadMutation = useCreateThread();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const previewsRef = useRef<FilePreview[]>([]);
  const [content, setContent] = useState('');
  const [previews, setPreviews] = useState<FilePreview[]>([]);

  useEffect(() => {
    previewsRef.current = previews;
  }, [previews]);

  useEffect(() => {
    return () => {
      previewsRef.current.forEach((preview) => {
        URL.revokeObjectURL(preview.url);
      });
    };
  }, []);

  const avatarSrc = profile?.avatar ?? null;
  const selectedFiles = previews.map((preview) => preview.file);
  const isPending = createThreadMutation.isPending;
  const normalizedContent = content.trim();
  const canSubmit =
    token !== null &&
    !isPending &&
    (normalizedContent.length > 0 || selectedFiles.length > 0);

  function handleFilesChange(event: ChangeEvent<HTMLInputElement>) {
    const incomingFiles = Array.from(event.target.files ?? []);

    if (incomingFiles.length === 0) {
      return;
    }

    setPreviews((currentPreviews) => [
      ...currentPreviews,
      ...incomingFiles.map((file) => ({
        file,
        kind: getPreviewKind(file),
        url: URL.createObjectURL(file),
      })),
    ]);
    event.target.value = '';
  }

  function removeSelectedFile(targetFile: File) {
    setPreviews((currentPreviews) => {
      const previewToRemove = currentPreviews.find((preview) => isSameFile(preview.file, targetFile));
      if (previewToRemove) {
        URL.revokeObjectURL(previewToRemove.url);
      }

      return currentPreviews.filter((preview) => !isSameFile(preview.file, targetFile));
    });
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!canSubmit) {
      return;
    }

    try {
      await createThreadMutation.mutateAsync({
        content: normalizedContent,
        files: selectedFiles,
      });

      setContent('');
      previews.forEach((preview) => {
        URL.revokeObjectURL(preview.url);
      });
      setPreviews([]);

      if (inputRef.current !== null) {
        inputRef.current.value = '';
      }
    } catch {
      // Mutation state already exposes the error to the UI.
    }
  }

  return (
    <form
      id="create-post-box"
      onSubmit={(event) => {
        void handleSubmit(event);
      }}
      style={{
        padding: '16px 20px',
        background: 'var(--bg-card)',
        borderBottom: '1px solid var(--border)',
      }}
    >
      <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
        <Avatar
          username={user?.username ?? 'Current user'}
          avatarUrl={avatarSrc}
          size={40}
        />

        <div style={{ flex: 1, minWidth: 0 }}>
          <textarea
            value={content}
            onChange={(event) => setContent(event.target.value)}
            placeholder="What is happening?!"
            rows={4}
            style={{
              width: '100%',
              resize: 'vertical',
              minHeight: 108,
              background: 'transparent',
              border: 'none',
              outline: 'none',
              color: 'var(--text)',
              fontSize: 15,
              lineHeight: 1.6,
              fontFamily: 'inherit',
              padding: '2px 0 10px',
            }}
          />

          {previews.length > 0 && (
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
                gap: 10,
                marginBottom: 12,
              }}
            >
              {previews.map((preview) => (
                <div
                  key={`${preview.file.name}-${preview.file.lastModified}-${preview.file.size}`}
                  style={{
                    position: 'relative',
                    borderRadius: 12,
                    overflow: 'hidden',
                    border: '1px solid var(--border)',
                    background: 'var(--bg-subtle)',
                    minHeight: 140,
                  }}
                >
                  {preview.kind === 'video' ? (
                    <video
                      src={preview.url}
                      muted
                      playsInline
                      controls
                      style={{
                        width: '100%',
                        height: '100%',
                        maxHeight: 220,
                        display: 'block',
                        objectFit: 'cover',
                      }}
                    />
                  ) : preview.kind === 'image' ? (
                    <img
                      src={preview.url}
                      alt={preview.file.name}
                      style={{
                        width: '100%',
                        height: '100%',
                        maxHeight: 220,
                        display: 'block',
                        objectFit: 'cover',
                      }}
                    />
                  ) : (
                    <div
                      style={{
                        minHeight: 140,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: 'var(--text-tertiary)',
                        fontSize: 13,
                        padding: 16,
                        textAlign: 'center',
                        lineHeight: 1.5,
                      }}
                    >
                      {preview.file.name}
                    </div>
                  )}

                  <button
                    type="button"
                    onClick={() => removeSelectedFile(preview.file)}
                    aria-label={`Remove ${preview.file.name}`}
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
              ))}
            </div>
          )}

          <div style={{ height: 1, background: 'var(--border)', margin: '0 0 12px' }} />

          {createThreadMutation.isError && (
            <div
              style={{
                marginBottom: 12,
                padding: '10px 12px',
                borderRadius: 10,
                background: 'var(--coral-soft)',
                color: 'var(--coral)',
                fontSize: 13,
                lineHeight: 1.5,
              }}
            >
              {createThreadMutation.error.message}
            </div>
          )}

          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: 12,
              flexWrap: 'wrap',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <input
                ref={inputRef}
                type="file"
                accept="image/*,video/*"
                multiple
                onChange={handleFilesChange}
                style={{ display: 'none' }}
                id="create-post-files"
              />

              <label
                htmlFor="create-post-files"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 8,
                  padding: '8px 12px',
                  borderRadius: 10,
                  border: '1px solid var(--border)',
                  color: 'var(--text)',
                  fontSize: 13,
                  fontWeight: 500,
                  cursor: isPending ? 'not-allowed' : 'pointer',
                  opacity: isPending ? 0.6 : 1,
                }}
              >
                <ImagePlus size={16} />
                <span>Add media</span>
              </label>

              {selectedFiles.length > 0 && (
                <span style={{ color: 'var(--text-tertiary)', fontSize: 13 }}>
                  {selectedFiles.length} file{selectedFiles.length > 1 ? 's' : ''} selected
                </span>
              )}
            </div>

            <button
              type="submit"
              disabled={!canSubmit}
              style={{
                background: 'var(--accent)',
                border: 'none',
                borderRadius: 10,
                color: '#fff',
                fontWeight: 600,
                fontSize: 14,
                padding: '9px 16px',
                cursor: canSubmit ? 'pointer' : 'not-allowed',
                opacity: canSubmit ? 1 : 0.5,
                fontFamily: 'inherit',
                display: 'inline-flex',
                alignItems: 'center',
                gap: 8,
              }}
            >
              {isPending && <LoaderCircle size={16} />}
              <span>{isPending ? 'Posting...' : 'Post'}</span>
            </button>
          </div>
        </div>
      </div>
    </form>
  );
}
