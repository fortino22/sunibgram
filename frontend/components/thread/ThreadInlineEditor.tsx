'use client';

import { ImagePlus, LoaderCircle, X, Smile, Film, Loader2 } from 'lucide-react';
import dynamic from 'next/dynamic';
import {
  type ChangeEvent,
  type MouseEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from 'react';
import type { EmojiClickData } from 'emoji-picker-react';
import type { ThreadMediaResponse, ThreadMutationParams, MentionUser } from '@/lib/services/threads';
import { decodeMentionsForEditing, encodeMentions } from '@/lib/mentions';
import { GifPicker } from '@/components/chat/GifPicker';
import { MentionAutocomplete } from '@/components/thread/MentionAutocomplete';

const EmojiPicker = dynamic(() => import('emoji-picker-react'), { ssr: false });

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
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);
  const [textareaEl, setTextareaEl] = useState<HTMLTextAreaElement | null>(null);
  const decoded = useMemo(() => decodeMentionsForEditing(initialContent), [initialContent]);
  const mentionMapRef = useRef<Map<string, string | number>>(decoded.mentionMap);
  const [content, setContent] = useState(decoded.text);
  const [removedMediaIds, setRemovedMediaIds] = useState<number[]>([]);
  const [previews, setPreviews] = useState<FilePreview[]>([]);

  // Mention autocomplete state
  const [mentionState, setMentionState] = useState<{ query: string; start: number; end: number } | null>(null);
  const [mentionIndex, setMentionIndex] = useState(0);
  const mentionResultsRef = useRef<MentionUser[]>([]);

  // Emoji / GIF pickers
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [showGifPicker, setShowGifPicker] = useState(false);
  const [gifFetching, setGifFetching] = useState(false);
  const emojiPickerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    previewsRef.current = previews;
  }, [previews]);

  useEffect(() => {
    return () => {
      previewsRef.current.forEach((preview) => URL.revokeObjectURL(preview.url));
    };
  }, []);

  useEffect(() => {
    if (!showEmojiPicker) return;
    const handler = (e: globalThis.MouseEvent) => {
      if (emojiPickerRef.current && !emojiPickerRef.current.contains(e.target as Node)) {
        setShowEmojiPicker(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [showEmojiPicker]);

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

  const canSave = normalizedContent.length > 0 && !pending && !gifFetching && !hasUnsupportedMediaState;

  function stopEvent(event: MouseEvent<HTMLElement>) {
    event.preventDefault();
    event.stopPropagation();
  }

  // --- Mention autocomplete --------------------------------------------------
  function detectMentionAt(text: string, caret: number): { query: string; start: number; end: number } | null {
    let i = caret - 1;
    while (i >= 0 && /[a-zA-Z0-9_.]/.test(text[i])) i--;
    if (i < 0 || text[i] !== '@') return null;
    const before = i === 0 ? '' : text[i - 1];
    if (before && !/\s/.test(before)) return null;
    const query = text.slice(i + 1, caret);
    if (query.length > 30) return null;
    return { query, start: i, end: caret };
  }

  function handleContentChange(e: ChangeEvent<HTMLTextAreaElement>) {
    const value = e.target.value;
    setContent(value);
    const caret = e.target.selectionStart ?? value.length;
    const detected = detectMentionAt(value, caret);
    if (detected) {
      setMentionState(detected);
      setMentionIndex(0);
    } else if (mentionState) {
      setMentionState(null);
    }
  }

  const insertMention = useCallback((user: MentionUser) => {
    const el = textareaRef.current;
    const state = mentionState;
    if (!state) return;
    const insertion = `@${user.username} `;
    setContent((prev) => prev.slice(0, state.start) + insertion + prev.slice(state.end));
    mentionMapRef.current.set(user.username, user.user_id);
    setMentionState(null);
    setMentionIndex(0);
    requestAnimationFrame(() => {
      if (!el) return;
      el.focus();
      const pos = state.start + insertion.length;
      el.setSelectionRange(pos, pos);
    });
  }, [mentionState]);

  function handleTextareaKeyDown(e: ReactKeyboardEvent<HTMLTextAreaElement>) {
    if (!mentionState) return;
    const results = mentionResultsRef.current;
    if (e.key === 'Escape') {
      e.preventDefault();
      setMentionState(null);
      return;
    }
    if (results.length === 0) return;
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setMentionIndex((i) => (i + 1) % results.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setMentionIndex((i) => (i - 1 + results.length) % results.length);
    } else if (e.key === 'Enter' || e.key === 'Tab') {
      e.preventDefault();
      const target = results[Math.min(mentionIndex, results.length - 1)];
      if (target) insertMention(target);
    }
  }

  const handleMentionResults = useCallback((results: MentionUser[]) => {
    mentionResultsRef.current = results;
    setMentionIndex((i) => (results.length === 0 ? 0 : Math.min(i, results.length - 1)));
  }, []);

  function handleTextareaSelect(e: React.SyntheticEvent<HTMLTextAreaElement>) {
    const el = e.currentTarget;
    const detected = detectMentionAt(el.value, el.selectionStart ?? 0);
    if (detected) {
      if (
        !mentionState ||
        mentionState.query !== detected.query ||
        mentionState.start !== detected.start
      ) {
        setMentionState(detected);
        setMentionIndex(0);
      }
    } else if (mentionState) {
      setMentionState(null);
    }
  }

  // --- Emoji / GIF -----------------------------------------------------------
  function handleEmojiClick(emojiData: EmojiClickData) {
    const el = textareaRef.current;
    if (!el) {
      setContent((prev) => prev + emojiData.emoji);
    } else {
      const start = el.selectionStart ?? content.length;
      const end = el.selectionEnd ?? content.length;
      setContent((prev) => prev.slice(0, start) + emojiData.emoji + prev.slice(end));
      requestAnimationFrame(() => {
        el.focus();
        const pos = start + emojiData.emoji.length;
        el.setSelectionRange(pos, pos);
      });
    }
    setShowEmojiPicker(false);
  }

  async function handleGifSelect(gifUrl: string) {
    setShowGifPicker(false);
    setGifFetching(true);
    try {
      const res = await fetch(gifUrl);
      const blob = await res.blob();
      const file = new File([blob], `gif-${gifUrl.length}-${blob.size}.gif`, { type: blob.type || 'image/gif' });
      const url = URL.createObjectURL(file);
      setPreviews((cur) => [...cur, { file, kind: 'image', url }]);
    } catch {
      // silently ignore if CORS blocks
    } finally {
      setGifFetching(false);
    }
  }

  // --- Media -----------------------------------------------------------------
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
        content: encodeMentions(normalizedContent, mentionMapRef.current),
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
    setContent(decoded.text);
    setRemovedMediaIds([]);
    previewsRef.current.forEach((preview) => URL.revokeObjectURL(preview.url));
    previewsRef.current = [];
    setPreviews([]);
    setShowEmojiPicker(false);
    setShowGifPicker(false);
    setMentionState(null);
    if (inputRef.current !== null) {
      inputRef.current.value = '';
    }
    onCancel();
  }

  const iconBtnStyle = (active: boolean, busy = false) => ({
    display: 'inline-flex' as const,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
    width: 36,
    height: 36,
    borderRadius: 8,
    background: 'none',
    border: 'none',
    color: active ? 'var(--accent)' : 'var(--text-secondary)',
    cursor: busy ? 'not-allowed' : 'pointer',
    opacity: busy ? 0.6 : 1,
  });

  return (
    <div
      className="mt-3 overflow-hidden rounded-2xl border"
      style={{ borderColor: 'var(--border)', background: 'var(--bg-card)' }}
      onClick={(event) => event.stopPropagation()}
    >
      {/* Content area */}
      <div style={{ padding: '12px 16px 0' }}>
        <textarea
          ref={(el) => {
            textareaRef.current = el;
            setTextareaEl(el);
          }}
          value={content}
          onChange={handleContentChange}
          onKeyDown={handleTextareaKeyDown}
          onSelect={handleTextareaSelect}
          rows={compact ? 2 : 3}
          autoFocus
          className="w-full resize-none bg-transparent text-base outline-none"
          style={{
            color: 'var(--text)',
            lineHeight: 1.5,
            minHeight: compact ? 56 : 72,
            border: 'none',
            padding: '2px 2px',
          }}
        />

        {mentionState && (
          <MentionAutocomplete
            query={mentionState.query}
            anchorEl={textareaEl}
            mentionStart={mentionState.start}
            caretIndex={mentionState.end}
            activeIndex={mentionIndex}
            onActiveIndexChange={setMentionIndex}
            onSelect={insertMention}
            onClose={() => setMentionState(null)}
            onResults={handleMentionResults}
          />
        )}

        {initialMedia.length > 0 && (
          <div className="mt-3">
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
          <div className="mt-3">
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

        {/* Emoji / GIF inline popovers */}
        {showEmojiPicker && (
          <div ref={emojiPickerRef} style={{ position: 'relative', zIndex: 50, marginTop: 10 }}>
            <EmojiPicker onEmojiClick={handleEmojiClick} lazyLoadEmojis height={360} width="100%" />
          </div>
        )}
        {showGifPicker && (
          <div style={{ position: 'relative', marginTop: 10 }}>
            <GifPicker
              placement="inline"
              onSelect={(url) => void handleGifSelect(url)}
              onClose={() => setShowGifPicker(false)}
            />
          </div>
        )}
      </div>

      {/* Action footer */}
      <div
        className="flex flex-wrap items-center justify-between gap-3"
        style={{ borderTop: '1px solid var(--border)', padding: '10px 16px', marginTop: 6 }}
      >
        <div className="flex items-center" style={{ gap: 4 }}>
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
            title="Add image or video"
            className="icon-knock"
            style={{
              ...iconBtnStyle(false, pending),
              pointerEvents: pending ? 'none' : undefined,
            }}
          >
            <ImagePlus size={18} />
          </label>

          <button
            type="button"
            title="Add GIF"
            className="icon-knock"
            onClick={() => { setShowGifPicker((v) => !v); setShowEmojiPicker(false); }}
            style={iconBtnStyle(showGifPicker, gifFetching)}
          >
            {gifFetching ? <Loader2 size={18} className="animate-spin" /> : <Film size={18} />}
          </button>

          <button
            type="button"
            title="Add emoji"
            className="icon-knock"
            onClick={() => { setShowEmojiPicker((v) => !v); setShowGifPicker(false); }}
            style={iconBtnStyle(showEmojiPicker)}
          >
            <Smile size={18} />
          </button>

          {(keptExistingMedia.length > 0 || selectedFiles.length > 0) && (
            <span className="text-sm" style={{ color: 'var(--text-tertiary)', marginLeft: 4 }}>
              {keptExistingMedia.length} current · {selectedFiles.length} new
            </span>
          )}
          {gifFetching && (
            <span className="text-sm" style={{ color: 'var(--accent)', marginLeft: 4 }}>Adding GIF…</span>
          )}
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={handleCancel}
            disabled={pending}
            className="btn-ghost inline-flex min-w-20 items-center justify-center rounded-full text-sm font-bold"
            style={{
              border: '1px solid var(--border)',
              color: 'var(--text)',
              padding: '9px 22px',
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
            className="btn-solid inline-flex min-w-20 items-center justify-center gap-2 rounded-full text-sm font-bold"
            style={{
              background: 'var(--text)',
              border: 0,
              color: 'var(--bg)',
              padding: '9px 22px',
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
