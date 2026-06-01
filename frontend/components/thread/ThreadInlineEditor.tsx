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

// Mirrors backend MaxThreadMediaItems (thread_media.go).
const MAX_MEDIA = 4;

type PreviewKind = 'image' | 'video' | 'file';

interface FilePreview {
  id: string;
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
  const previewIdRef = useRef(0);
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

  // Backend supports granular media edits: keep some existing (keep_media_ids),
  // delete the rest, append new files — up to MAX_MEDIA total.
  const totalMedia = keptExistingMedia.length + selectedFiles.length;
  const atMediaLimit = totalMedia >= MAX_MEDIA;

  const canSave = normalizedContent.length > 0 && !pending && !gifFetching && totalMedia <= MAX_MEDIA;

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
    if (atMediaLimit) {
      return;
    }
    setGifFetching(true);
    try {
      const res = await fetch(gifUrl);
      const blob = await res.blob();
      const file = new File([blob], `gif-${gifUrl.length}-${blob.size}.gif`, { type: blob.type || 'image/gif' });
      const url = URL.createObjectURL(file);
      setPreviews((cur) => [...cur, { id: `pv-${previewIdRef.current++}`, file, kind: 'image', url }]);
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

    // Cap total media at MAX_MEDIA — drop overflow so multi-select can't overshoot.
    const remaining = MAX_MEDIA - keptExistingMedia.length - previews.length;
    const accepted = remaining > 0 ? incoming.slice(0, remaining) : [];
    if (accepted.length > 0) {
      setPreviews((current) => [
        ...current,
        ...accepted.map((file) => ({
          id: `pv-${previewIdRef.current++}`,
          file,
          kind: getPreviewKind(file),
          url: URL.createObjectURL(file),
        })),
      ]);
    }
    event.target.value = '';
  }

  function removeSelectedFile(id: string) {
    setPreviews((current) => {
      const previewToRemove = current.find((preview) => preview.id === id);
      if (previewToRemove !== undefined) {
        URL.revokeObjectURL(previewToRemove.url);
      }

      return current.filter((preview) => preview.id !== id);
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

      const mediaTouched = removedMediaIds.length > 0 || selectedFiles.length > 0;
      if (mediaTouched) {
        if (hasExistingMedia) {
          // Present even if empty → backend deletes everything not listed.
          params.keepMediaIds = keptExistingMedia.map((item) => item.media_id);
        }
        if (selectedFiles.length > 0) {
          params.files = selectedFiles;
        }
      }
      // Untouched media → send neither field → backend keeps all.

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

        {(initialMedia.length > 0 || previews.length > 0) && (
          <div
            className={`mt-3 grid gap-2 ${
              initialMedia.length + previews.length > 2 ? 'grid-cols-3' : 'grid-cols-2'
            }`}
          >
            {initialMedia.map((item) => {
              const removed = removedMediaIds.includes(item.media_id);
              const kind = getRemoteMediaKind(item.url);

              return (
                <div
                  key={`existing-${item.media_id}`}
                  className="relative aspect-square overflow-hidden rounded-2xl border"
                  style={{
                    borderColor: 'var(--border)',
                    background: 'var(--bg-subtle)',
                    opacity: removed ? 0.5 : 1,
                    transition: 'opacity 0.12s ease',
                  }}
                >
                  {kind === 'video' ? (
                    <video src={item.url} muted playsInline className="h-full w-full object-cover" />
                  ) : (
                    <img src={item.url} alt="" className="h-full w-full object-cover" />
                  )}

                  <button
                    type="button"
                    onClick={(event) => {
                      stopEvent(event);
                      toggleExistingMedia(item.media_id);
                    }}
                    aria-label={removed ? 'Keep this media' : 'Remove this media'}
                    className={`absolute right-1.5 top-1.5 inline-flex h-7 items-center justify-center rounded-full ${removed ? 'px-2.5 text-xs font-semibold' : 'w-7'}`}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'rgba(0,0,0,0.85)'; e.currentTarget.style.transform = 'scale(1.08)'; }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = removed ? 'rgba(15,20,25,0.7)' : 'rgba(0,0,0,0.65)'; e.currentTarget.style.transform = 'scale(1)'; }}
                    style={{
                      border: '1px solid rgba(255,255,255,0.18)',
                      background: removed ? 'rgba(15,20,25,0.7)' : 'rgba(0,0,0,0.65)',
                      color: '#fff',
                      cursor: 'pointer',
                      backdropFilter: 'blur(6px)',
                      WebkitBackdropFilter: 'blur(6px)',
                      boxShadow: '0 1px 4px rgba(0,0,0,0.35)',
                      transition: 'background 0.12s ease, transform 0.12s ease',
                    }}
                  >
                    {removed ? 'Undo' : <X size={15} strokeWidth={2.5} />}
                  </button>
                </div>
              );
            })}

            {previews.map((preview) => (
              <div
                key={preview.id}
                className="relative aspect-square overflow-hidden rounded-2xl border"
                style={{ borderColor: 'var(--border)', background: 'var(--bg-subtle)' }}
              >
                {preview.kind === 'video' ? (
                  <video src={preview.url} muted playsInline className="h-full w-full object-cover" />
                ) : preview.kind === 'image' ? (
                  <img src={preview.url} alt={preview.file.name} className="h-full w-full object-cover" />
                ) : (
                  <div
                    className="flex h-full items-center justify-center px-4 text-center text-sm"
                    style={{ color: 'var(--text-tertiary)' }}
                  >
                    {preview.file.name}
                  </div>
                )}

                <span
                  className="absolute left-2 top-2 rounded-full px-2 py-0.5 text-[11px] font-bold uppercase tracking-wide"
                  style={{ background: 'var(--text)', color: 'var(--bg)' }}
                >
                  New
                </span>

                <button
                  type="button"
                  onClick={(event) => {
                    stopEvent(event);
                    removeSelectedFile(preview.id);
                  }}
                  aria-label={`Remove ${preview.file.name}`}
                  className="absolute right-1.5 top-1.5 inline-flex h-7 w-7 items-center justify-center rounded-full"
                  onMouseEnter={(e) => { e.currentTarget.style.background = 'rgba(0,0,0,0.85)'; e.currentTarget.style.transform = 'scale(1.08)'; }}
                  onMouseLeave={(e) => { e.currentTarget.style.background = 'rgba(0,0,0,0.65)'; e.currentTarget.style.transform = 'scale(1)'; }}
                  style={{
                    border: '1px solid rgba(255,255,255,0.18)',
                    background: 'rgba(0,0,0,0.65)',
                    color: '#fff',
                    cursor: 'pointer',
                    backdropFilter: 'blur(6px)',
                    WebkitBackdropFilter: 'blur(6px)',
                    boxShadow: '0 1px 4px rgba(0,0,0,0.35)',
                    transition: 'background 0.12s ease, transform 0.12s ease',
                  }}
                >
                  <X size={15} strokeWidth={2.5} />
                </button>
              </div>
            ))}
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
            htmlFor={atMediaLimit || pending ? undefined : inputId}
            title={atMediaLimit ? `Maximum ${MAX_MEDIA} media` : 'Add image or video'}
            className="icon-knock"
            style={{
              ...iconBtnStyle(false, pending || atMediaLimit),
              pointerEvents: pending || atMediaLimit ? 'none' : undefined,
            }}
          >
            <ImagePlus size={18} />
          </label>

          <button
            type="button"
            title={atMediaLimit ? `Maximum ${MAX_MEDIA} media` : 'Add GIF'}
            className="icon-knock"
            disabled={atMediaLimit || gifFetching}
            onClick={() => { setShowGifPicker((v) => !v); setShowEmojiPicker(false); }}
            style={iconBtnStyle(showGifPicker, gifFetching || atMediaLimit)}
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

          {totalMedia > 0 && (
            <span
              className="text-sm"
              style={{ color: atMediaLimit ? 'var(--text-secondary)' : 'var(--text-tertiary)', marginLeft: 4 }}
            >
              {totalMedia} / {MAX_MEDIA}
              {atMediaLimit ? ' · Max' : ''}
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
