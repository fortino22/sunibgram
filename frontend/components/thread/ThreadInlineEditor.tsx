'use client';

import { ImagePlus, LoaderCircle, X } from 'lucide-react';
import {
  type ChangeEvent,
  type MouseEvent,
  useEffect,
  useId,
  useRef,
  useState,
} from 'react';
import type { ThreadMediaResponse, ThreadMutationParams } from '@/lib/services/threads';

type PreviewKind = 'image' | 'video' | 'file';

interface FilePreview {
  file: File;
  kind: PreviewKind;
  url: string;
}

function getPreviewKind(file: File): PreviewKind {
  if (file.type.startsWith('image/')) return 'image';
  if (file.type.startsWith('video/')) return 'video';
  return 'file';
}

function getRemoteMediaKind(url: string): PreviewKind {
  const normalized = url.split('?')[0].toLowerCase();

  if (/\.(mp4|mov|webm|ogg|m4v)$/i.test(normalized)) {
    return 'video';
  }

  if (/\.(png|jpe?g|gif|webp|bmp|svg|avif|heic)$/i.test(normalized)) {
    return 'image';
  }

  return 'image';
}

function isSameFile(left: File, right: File): boolean {
  return (
    left.name === right.name &&
    left.lastModified === right.lastModified &&
    left.size === right.size
  );
}

interface ThreadInlineEditorProps {
  initialContent: string;
  initialMedia: ThreadMediaResponse[];
  onCancel: () => void;
  onSave: (params: ThreadMutationParams) => Promise<void>;
  pending: boolean;
  errorMessage?: string;
  compact?: boolean;
}

export function ThreadInlineEditor({
  initialContent,
  initialMedia,
  onCancel,
  onSave,
  pending,
  errorMessage,
  compact = false,
}: ThreadInlineEditorProps) {
  const inputId = useId();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const previewsRef = useRef<FilePreview[]>([]);
  const [content, setContent] = useState(initialContent);
  const [removedMediaIds, setRemovedMediaIds] = useState<number[]>([]);
  const [previews, setPreviews] = useState<FilePreview[]>([]);

  useEffect(() => {
    previewsRef.current = previews;
  }, [previews]);

  useEffect(() => {
    return () => {
      previewsRef.current.forEach((preview) => URL.revokeObjectURL(preview.url));
    };
  }, []);

  const normalizedContent = content.trim();
  const selectedFiles = previews.map((preview) => preview.file);
  const keptExistingMedia = initialMedia.filter((item) => !removedMediaIds.includes(item.media_id));
  const hasExistingMedia = initialMedia.length > 0;
  const removedAllExistingMedia = hasExistingMedia && keptExistingMedia.length === 0;
  const removedSomeExistingMedia = keptExistingMedia.length > 0 && keptExistingMedia.length < initialMedia.length;

  // The current API only supports either keeping the whole existing media set,
  // clearing it entirely via clear_media=true, or replacing it completely by
  // sending a new files[] set. It cannot persist a partial subset of existing media.
  const hasUnsupportedMediaState =
    removedSomeExistingMedia ||
    (selectedFiles.length > 0 && keptExistingMedia.length > 0);

  const mediaMessage =
    selectedFiles.length > 0 && keptExistingMedia.length > 0
      ? 'New uploads replace the full current media set. Remove the remaining current media first.'
      : removedSomeExistingMedia
        ? 'The current API cannot remove only some existing media. Restore all current media or remove the rest.'
        : selectedFiles.length > 0 && removedAllExistingMedia
          ? 'Saving will replace the current media with the new uploads.'
          : undefined;

  const canSave = normalizedContent.length > 0 && !pending && !hasUnsupportedMediaState;

  function stopEvent(event: MouseEvent<HTMLElement>) {
    event.preventDefault();
    event.stopPropagation();
  }

  function handleFilesChange(event: ChangeEvent<HTMLInputElement>) {
    const incoming = Array.from(event.target.files ?? []);
    if (incoming.length === 0) {
      return;
    }

    setPreviews((current) => [
      ...current,
      ...incoming.map((file) => ({
        file,
        kind: getPreviewKind(file),
        url: URL.createObjectURL(file),
      })),
    ]);
    event.target.value = '';
  }

  function removeSelectedFile(target: File) {
    setPreviews((current) => {
      const previewToRemove = current.find((preview) => isSameFile(preview.file, target));
      if (previewToRemove !== undefined) {
        URL.revokeObjectURL(previewToRemove.url);
      }

      return current.filter((preview) => !isSameFile(preview.file, target));
    });
  }

  function toggleExistingMedia(mediaId: number) {
    setRemovedMediaIds((current) =>
      current.includes(mediaId)
        ? current.filter((id) => id !== mediaId)
        : [...current, mediaId]
    );
  }

  async function handleSubmit(event: MouseEvent<HTMLButtonElement>) {
    stopEvent(event);
    if (!canSave) {
      return;
    }

    try {
      const params: ThreadMutationParams = {
        content: normalizedContent,
      };

      if (selectedFiles.length > 0) {
        params.files = selectedFiles;
      } else if (removedAllExistingMedia) {
        params.clearMedia = true;
      }

      await onSave(params);
    } catch {
      // Mutation state already exposes the error.
    }
  }

  function handleCancel(event: MouseEvent<HTMLButtonElement>) {
    stopEvent(event);
    setContent(initialContent);
    setRemovedMediaIds([]);
    previewsRef.current.forEach((preview) => URL.revokeObjectURL(preview.url));
    previewsRef.current = [];
    setPreviews([]);
    if (inputRef.current !== null) {
      inputRef.current.value = '';
    }
    onCancel();
  }

  return (
    <div
      className="mt-3 overflow-hidden rounded-2xl border"
      style={{ borderColor: 'var(--border)' }}
      onClick={(event) => event.stopPropagation()}
    >
      {/* Content area */}
      <div className="px-5 pt-5 pb-4">
        <textarea
          value={content}
          onChange={(event) => setContent(event.target.value)}
          rows={compact ? 5 : 6}
          autoFocus
          className="w-full resize-y bg-transparent text-base outline-none"
          style={{
            color: 'var(--text)',
            lineHeight: 1.7,
            minHeight: compact ? 148 : 172,
            border: 'none',
          }}
        />

        {initialMedia.length > 0 && (
          <div className="mt-4">
            <div
              className="mb-3 text-xs font-semibold uppercase tracking-[0.18em]"
              style={{ color: 'var(--text-tertiary)' }}
            >
              Current media
            </div>
            <div className="grid grid-cols-2 gap-3">
              {initialMedia.map((item) => {
                const removed = removedMediaIds.includes(item.media_id);
                const kind = getRemoteMediaKind(item.url);

                return (
                  <div
                    key={item.media_id}
                    className="relative overflow-hidden rounded-2xl border"
                    style={{
                      borderColor: 'var(--border)',
                      background: 'var(--bg-subtle)',
                      opacity: removed ? 0.55 : 1,
                    }}
                  >
                    {kind === 'video' ? (
                      <video
                        src={item.url}
                        controls
                        muted
                        playsInline
                        className="block max-h-55 min-h-35 w-full object-cover"
                      />
                    ) : (
                      <img
                        src={item.url}
                        alt=""
                        className="block max-h-55 min-h-35 w-full object-cover"
                      />
                    )}

                    {removed && (
                      <div
                        className="absolute inset-x-0 bottom-0 px-3 py-2 text-xs font-medium"
                        style={{
                          background: 'rgba(15, 20, 25, 0.72)',
                          color: '#fff',
                        }}
                      >
                        Removed from this edit
                      </div>
                    )}

                    <button
                      type="button"
                      onClick={(event) => {
                        stopEvent(event);
                        toggleExistingMedia(item.media_id);
                      }}
                      aria-label={removed ? 'Keep current media' : 'Remove current media'}
                      className="absolute right-2 top-2 inline-flex h-8 min-w-8 items-center justify-center rounded-full px-2"
                      onMouseEnter={(e) => { e.currentTarget.style.filter = 'brightness(0.85)'; }}
                      onMouseLeave={(e) => { e.currentTarget.style.filter = 'none'; }}
                      style={{
                        border: '1px solid rgba(255,255,255,0.22)',
                        background: removed ? 'rgba(15,20,25,0.78)' : 'rgba(0,0,0,0.55)',
                        color: '#fff',
                        cursor: 'pointer',
                        transition: 'filter 0.12s ease',
                      }}
                    >
                      {removed ? 'Undo' : <X size={14} />}
                    </button>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {previews.length > 0 && (
          <div className="mt-4">
            <div
              className="mb-3 text-xs font-semibold uppercase tracking-[0.18em]"
              style={{ color: 'var(--text-tertiary)' }}
            >
              New uploads
            </div>
            <div className="grid grid-cols-2 gap-3">
              {previews.map((preview) => (
                <div
                  key={`${preview.file.name}-${preview.file.lastModified}-${preview.file.size}`}
                  className="relative overflow-hidden rounded-2xl border"
                  style={{
                    borderColor: 'var(--border)',
                    background: 'var(--bg-subtle)',
                  }}
                >
                  {preview.kind === 'video' ? (
                    <video
                      src={preview.url}
                      controls
                      muted
                      playsInline
                      className="block max-h-55 min-h-35 w-full object-cover"
                    />
                  ) : preview.kind === 'image' ? (
                    <img
                      src={preview.url}
                      alt={preview.file.name}
                      className="block max-h-55 min-h-35 w-full object-cover"
                    />
                  ) : (
                    <div
                      className="flex min-h-35 items-center justify-center px-4 text-center text-sm"
                      style={{ color: 'var(--text-tertiary)' }}
                    >
                      {preview.file.name}
                    </div>
                  )}

                  <button
                    type="button"
                    onClick={(event) => {
                      stopEvent(event);
                      removeSelectedFile(preview.file);
                    }}
                    aria-label={`Remove ${preview.file.name}`}
                    className="absolute right-2 top-2 inline-flex h-8 w-8 items-center justify-center rounded-full"
                    onMouseEnter={(e) => { e.currentTarget.style.filter = 'brightness(0.85)'; }}
                    onMouseLeave={(e) => { e.currentTarget.style.filter = 'none'; }}
                    style={{
                      border: '1px solid rgba(255,255,255,0.22)',
                      background: 'rgba(0,0,0,0.55)',
                      color: '#fff',
                      cursor: 'pointer',
                      transition: 'filter 0.12s ease',
                    }}
                  >
                    <X size={14} />
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {mediaMessage && (
          <div
            className="mt-3 rounded-xl px-3 py-2 text-[13px]"
            style={{
              background: hasUnsupportedMediaState ? 'var(--amber-soft)' : 'var(--bg-subtle)',
              color: hasUnsupportedMediaState ? 'var(--amber)' : 'var(--text-tertiary)',
            }}
          >
            {mediaMessage}
          </div>
        )}

        {errorMessage && (
          <div
            className="mt-3 rounded-xl px-3 py-2 text-[13px]"
            style={{
              background: 'var(--coral-soft)',
              color: 'var(--coral)',
            }}
          >
            {errorMessage}
          </div>
        )}
      </div>

      {/* Action footer */}
      <div
        className="flex flex-wrap items-center justify-between gap-3 px-5 py-3.5"
        style={{ borderTop: '1px solid var(--border)' }}
      >
        <div className="flex flex-wrap items-center gap-3">
          <input
            ref={inputRef}
            id={inputId}
            type="file"
            accept="image/*,video/*"
            multiple
            onChange={handleFilesChange}
            style={{ display: 'none' }}
          />
          <label
            htmlFor={inputId}
            className="inline-flex min-h-10 cursor-pointer items-center gap-2 rounded-full px-4 py-2 text-sm font-semibold"
            style={{
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
              opacity: pending ? 0.6 : 1,
              pointerEvents: pending ? 'none' : 'auto',
            }}
          >
            <ImagePlus size={16} />
            <span>Add media</span>
          </label>

          {(keptExistingMedia.length > 0 || selectedFiles.length > 0) && (
            <span className="text-sm" style={{ color: 'var(--text-tertiary)' }}>
              {keptExistingMedia.length} current · {selectedFiles.length} new
            </span>
          )}
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={handleCancel}
            disabled={pending}
            className="btn-ghost inline-flex min-h-10 min-w-24 items-center justify-center rounded-full px-5 py-2.5 text-sm font-bold"
            style={{
              border: '1px solid var(--border)',
              color: 'var(--text)',
              cursor: pending ? 'not-allowed' : 'pointer',
              opacity: pending ? 0.5 : 1,
            }}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={(event) => void handleSubmit(event)}
            disabled={!canSave}
            className="btn-solid inline-flex min-h-10 min-w-20 items-center justify-center gap-2 rounded-full px-5 py-2.5 text-sm font-bold"
            style={{
              background: 'var(--accent)',
              border: 0,
              color: '#fff',
              cursor: canSave ? 'pointer' : 'not-allowed',
              opacity: canSave ? 1 : 0.5,
            }}
          >
            {pending && <LoaderCircle size={15} className="animate-spin" />}
            <span>{pending ? 'Saving' : 'Save'}</span>
          </button>
        </div>
      </div>
    </div>
  );
}
