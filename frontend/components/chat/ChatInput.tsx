'use client';

import { useRef, useState, useCallback, useEffect } from 'react';
import dynamic from 'next/dynamic';
import { Send, X, ImageIcon, Smile, Mic, Sticker, Video, Square, Check } from 'lucide-react';
import { toast } from 'sonner';
import { useSendMessage } from '@/hooks/chat/useSendMessage';
import { useEditMessage } from '@/hooks/chat/useEditMessage';
import { compressImages, IMAGE_MAX_BYTES } from '@/lib/compressImage';
import { GifPicker } from '@/components/chat/GifPicker';
import type { EmojiClickData } from 'emoji-picker-react';
import type { ChatMessage } from '@/lib/services/chat';

const EmojiPicker = dynamic(() => import('emoji-picker-react'), { ssr: false });

interface ChatInputProps {
  roomId: number;
  editingMessage?: ChatMessage | null;
  onCancelEdit?: () => void;
  replyingTo?: ChatMessage | null;
  onCancelReply?: () => void;
  onStartTyping?: (roomId: number) => void;
  onStopTyping?: (roomId: number) => void;
}

const ALLOWED_TYPES = [
  'image/jpeg', 'image/png', 'image/gif', 'image/webp',
  'video/mp4', 'video/quicktime', 'video/x-msvideo', 'video/x-matroska', 'video/webm',
];

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, '0')}`;
}

export function ChatInput({
  roomId,
  editingMessage,
  onCancelEdit,
  replyingTo,
  onCancelReply,
  onStartTyping,
  onStopTyping,
}: ChatInputProps) {
  const isEditMode = !!editingMessage;

  const [text, setText] = useState('');
  const [files, setFiles] = useState<File[]>([]);
  const [pendingGif, setPendingGif] = useState<string | null>(null);
  const [compressing, setCompressing] = useState(false);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [showGifPicker, setShowGifPicker] = useState(false);
  const [isRecording, setIsRecording] = useState(false);
  const [recordingSeconds, setRecordingSeconds] = useState(0);

  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const emojiPickerRef = useRef<HTMLDivElement>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const isCancelRef = useRef(false);
  const recordingTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const typingTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isTypingRef = useRef(false);

  const sendMessage = useSendMessage();
  const editMessageMutation = useEditMessage(roomId);

  // When editingMessage changes, pre-fill the textarea
  useEffect(() => {
    if (editingMessage) {
      const msg = editingMessage.message ?? '';
      setText(msg);
      requestAnimationFrame(() => {
        const el = textareaRef.current;
        if (!el) return;
        el.focus();
        el.style.height = 'auto';
        el.style.height = `${Math.min(el.scrollHeight, 120)}px`;
        el.setSelectionRange(msg.length, msg.length);
      });
    } else {
      // Only clear if exiting edit mode (not on initial mount)
      if (isEditMode) setText('');
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editingMessage?.chat_id]);

  useEffect(() => {
    if (!showEmojiPicker) return;
    const handler = (e: MouseEvent) => {
      if (emojiPickerRef.current && !emojiPickerRef.current.contains(e.target as Node)) {
        setShowEmojiPicker(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [showEmojiPicker]);

  const startTyping = useCallback(() => {
    if (!isTypingRef.current) {
      isTypingRef.current = true;
      onStartTyping?.(roomId);
    }
    if (typingTimer.current) clearTimeout(typingTimer.current);
    typingTimer.current = setTimeout(() => {
      isTypingRef.current = false;
      onStopTyping?.(roomId);
    }, 2000);
  }, [roomId, onStartTyping, onStopTyping]);

  const stopTyping = useCallback(() => {
    if (typingTimer.current) clearTimeout(typingTimer.current);
    if (isTypingRef.current) {
      isTypingRef.current = false;
      onStopTyping?.(roomId);
    }
  }, [roomId, onStopTyping]);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value);
    const el = e.target;
    el.style.height = 'auto';
    el.style.height = `${Math.min(el.scrollHeight, 120)}px`;
    if (!isEditMode) {
      if (e.target.value) startTyping();
      else stopTyping();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      isEditMode ? handleSaveEdit() : handleSend();
    }
    if (e.key === 'Escape' && isEditMode) {
      handleCancelEdit();
    }
  };

  const handleEmojiClick = (emojiData: EmojiClickData) => {
    const el = textareaRef.current;
    if (!el) { setText((prev) => prev + emojiData.emoji); return; }
    const start = el.selectionStart ?? text.length;
    const end = el.selectionEnd ?? text.length;
    const next = text.slice(0, start) + emojiData.emoji + text.slice(end);
    setText(next);
    requestAnimationFrame(() => {
      el.focus();
      const pos = start + emojiData.emoji.length;
      el.setSelectionRange(pos, pos);
      el.style.height = 'auto';
      el.style.height = `${Math.min(el.scrollHeight, 120)}px`;
    });
    if (!isEditMode) startTyping();
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const all = Array.from(e.target.files ?? []);
    e.target.value = '';

    const rejected = all.filter((f) => !ALLOWED_TYPES.includes(f.type));
    if (rejected.length > 0) {
      toast.error('Only JPEG, PNG, WebP, GIF, and video files are supported.');
    }

    const picked = all.filter((f) => ALLOWED_TYPES.includes(f.type));
    if (picked.length === 0) return;

    setCompressing(true);
    try {
      const ready = await compressImages(picked);
      const oversized = ready.filter((f) =>
        f.type.startsWith('image/')
          ? f.size > IMAGE_MAX_BYTES.media
          : f.size > 3.8 * 1024 * 1024,
      );
      if (oversized.length > 0) {
        toast.error(`${oversized.length} file(s) are still too large after compression and were removed.`);
      }
      const accepted = ready.filter((f) => !oversized.includes(f));
      if (accepted.length > 0) setFiles((prev) => [...prev, ...accepted]);
    } finally {
      setCompressing(false);
    }
  };

  const removeFile = (idx: number) => setFiles((prev) => prev.filter((_, i) => i !== idx));

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const preferredTypes = [
        'audio/webm;codecs=opus',
        'audio/webm',
        'audio/mp4',
        'audio/ogg;codecs=opus',
        'audio/ogg',
      ];
      const mimeType = preferredTypes.find((t) => MediaRecorder.isTypeSupported(t)) ?? '';
      const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined);
      audioChunksRef.current = [];
      isCancelRef.current = false;

      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) audioChunksRef.current.push(e.data);
      };

      recorder.onstop = () => {
        stream.getTracks().forEach((t) => t.stop());
        if (!isCancelRef.current) {
          const actualMime = recorder.mimeType || mimeType || 'audio/webm';
          const ext = actualMime.includes('mp4') ? '.m4a' : actualMime.includes('ogg') ? '.ogg' : '.webm';
          const blob = new Blob(audioChunksRef.current, { type: actualMime });
          if (blob.size > 0) {
            const file = new File([blob], `voice-note-${Date.now()}${ext}`, { type: actualMime });
            setFiles((prev) => [...prev, file]);
          }
        }
        isCancelRef.current = false;
        setIsRecording(false);
        setRecordingSeconds(0);
      };

      mediaRecorderRef.current = recorder;
      recorder.start(250);
      setIsRecording(true);
      setRecordingSeconds(0);
      recordingTimerRef.current = setInterval(() => setRecordingSeconds((s) => s + 1), 1000);
    } catch {
      // Microphone permission denied or unavailable
    }
  };

  const stopRecording = () => {
    if (recordingTimerRef.current) clearInterval(recordingTimerRef.current);
    mediaRecorderRef.current?.stop();
  };

  const cancelRecording = () => {
    isCancelRef.current = true;
    if (recordingTimerRef.current) clearInterval(recordingTimerRef.current);
    if (mediaRecorderRef.current?.state !== 'inactive') mediaRecorderRef.current?.stop();
    setIsRecording(false);
    setRecordingSeconds(0);
  };

  const handleSend = () => {
    const trimmed = text.trim();
    if (!trimmed && files.length === 0 && !pendingGif) return;
    stopTyping();
    setShowEmojiPicker(false);
    setShowGifPicker(false);
    sendMessage.mutate(
      { roomId, message: trimmed || null, files, gifUrl: pendingGif, replyToId: replyingTo?.chat_id ?? null },
      {
        onSuccess: () => {
          setText('');
          setFiles([]);
          setPendingGif(null);
          onCancelReply?.();
          if (textareaRef.current) textareaRef.current.style.height = 'auto';
        },
      },
    );
  };

  const handleSaveEdit = () => {
    if (!editingMessage) return;
    const trimmed = text.trim();
    if (!trimmed) return;
    if (trimmed === (editingMessage.message ?? '').trim()) {
      handleCancelEdit();
      return;
    }
    editMessageMutation.mutate(
      { chatId: editingMessage.chat_id, message: trimmed },
      { onSuccess: () => onCancelEdit?.() },
    );
  };

  const handleCancelEdit = () => {
    setText('');
    if (textareaRef.current) textareaRef.current.style.height = 'auto';
    onCancelEdit?.();
  };

  const canSend = isEditMode
    ? text.trim().length > 0 && !editMessageMutation.isPending
    : (text.trim().length > 0 || files.length > 0 || !!pendingGif) && !sendMessage.isPending && !compressing;

  return (
    <div
      style={{
        borderTop: '1px solid var(--border)',
        padding: '10px 16px',
        flexShrink: 0,
        background: 'var(--bg-card)',
        position: 'relative',
      }}
    >
      {showEmojiPicker && (
        <div
          ref={emojiPickerRef}
          style={{ position: 'absolute', bottom: 'calc(100% + 6px)', left: 16, zIndex: 50 }}
        >
          <EmojiPicker onEmojiClick={handleEmojiClick} lazyLoadEmojis height={380} width={320} />
        </div>
      )}

      {showGifPicker && (
        <GifPicker
          onSelect={(url) => { setPendingGif(url); setShowGifPicker(false); }}
          onClose={() => setShowGifPicker(false)}
        />
      )}

      {/* Reply banner */}
      {replyingTo && !isEditMode && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            marginBottom: 8,
            padding: '6px 10px',
            borderRadius: 8,
            background: 'var(--bg)',
            borderLeft: '3px solid var(--accent)',
          }}
        >
          <div style={{ flex: 1, minWidth: 0 }}>
            <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--accent)', display: 'block' }}>
              Replying to @{replyingTo.sender.username}
            </span>
            <span
              style={{
                fontSize: 11,
                color: 'var(--text-muted)',
                display: 'block',
                overflow: 'hidden',
                whiteSpace: 'nowrap',
                textOverflow: 'ellipsis',
              }}
            >
              {replyingTo.message
                ? replyingTo.message.slice(0, 60)
                : replyingTo.gif_url
                ? 'GIF'
                : 'Media'}
            </span>
          </div>
          <button
            className="hover-dim"
            onClick={onCancelReply}
            style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 2, color: 'var(--text-muted)', display: 'flex', alignItems: 'center', flexShrink: 0 }}
          >
            <X size={14} />
          </button>
        </div>
      )}

      {/* Edit mode banner */}
      {isEditMode && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            marginBottom: 8,
            padding: '6px 10px',
            borderRadius: 8,
            background: 'var(--bg)',
            border: '1px solid var(--accent)',
          }}
        >
          <span style={{ fontSize: 12, color: 'var(--accent)', fontWeight: 600, flex: 1 }}>
            Editing message
          </span>
          <button
            className="hover-dim"
            onClick={handleCancelEdit}
            style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 2, color: 'var(--text-muted)', display: 'flex', alignItems: 'center' }}
          >
            <X size={14} />
          </button>
        </div>
      )}

      {/* Pending GIF preview (only in non-edit mode) */}
      {pendingGif && !isEditMode && (
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: 8, marginBottom: 10 }}>
          <div style={{ position: 'relative', borderRadius: 8, overflow: 'hidden', border: '1px solid var(--border)' }}>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img src={pendingGif} alt="GIF" style={{ width: 96, height: 72, objectFit: 'cover', display: 'block' }} />
            <button
              className="hover-dim"
              onClick={() => setPendingGif(null)}
              style={{ position: 'absolute', top: 2, right: 2, width: 18, height: 18, borderRadius: '50%', background: 'rgba(0,0,0,0.6)', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 0 }}
            >
              <X size={10} color="#fff" />
            </button>
          </div>
        </div>
      )}

      {/* File previews (only in non-edit mode) */}
      {files.length > 0 && !isEditMode && (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 10 }}>
          {files.map((file, idx) => (
            <div
              key={idx}
              style={{ position: 'relative', borderRadius: 8, overflow: 'hidden', border: '1px solid var(--border)' }}
            >
              {file.type.startsWith('image/') ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={URL.createObjectURL(file)} alt={file.name} style={{ width: 72, height: 72, objectFit: 'cover', display: 'block' }} />
              ) : file.type.startsWith('audio/') ? (
                <div style={{ width: 72, height: 72, background: 'var(--bg)', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 4, color: 'var(--accent)' }}>
                  <Mic size={22} />
                  <span style={{ fontSize: 9, color: 'var(--text-muted)' }}>Voice note</span>
                </div>
              ) : (
                <div style={{ width: 72, height: 72, background: 'var(--bg)', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 4, color: 'var(--text-muted)' }}>
                  <Video size={22} />
                  <span style={{ fontSize: 9, textAlign: 'center', padding: '0 4px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', width: '100%' }}>{file.name}</span>
                </div>
              )}
              <button
                className="hover-dim"
                onClick={() => removeFile(idx)}
                style={{ position: 'absolute', top: 2, right: 2, width: 18, height: 18, borderRadius: '50%', background: 'rgba(0,0,0,0.6)', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 0 }}
              >
                <X size={10} color="#fff" />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Pill bar */}
      <div
        style={{
          display: 'flex',
          alignItems: isRecording ? 'center' : 'flex-end',
          border: `1px solid ${isRecording ? 'var(--accent)' : isEditMode ? 'var(--accent)' : 'var(--border)'}`,
          borderRadius: 24,
          background: 'var(--bg)',
          padding: '6px 10px',
          gap: 4,
          transition: 'border-color 0.15s',
        }}
      >
        {isRecording ? (
          <>
            <button
              onClick={cancelRecording}
              title="Cancel recording"
              style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '2px 6px', display: 'flex', alignItems: 'center', color: 'var(--text-muted)', flexShrink: 0 }}
            >
              <X size={18} />
            </button>

            <div style={{ flex: 1, display: 'flex', alignItems: 'center', gap: 8 }}>
              <div style={{
                width: 8, height: 8, borderRadius: '50%',
                background: 'oklch(60% 0.22 25)',
                flexShrink: 0,
                animation: 'pulse 1s ease-in-out infinite',
              }} />
              <span style={{ fontSize: 13, color: 'var(--text)', fontVariantNumeric: 'tabular-nums' }}>
                {formatDuration(recordingSeconds)}
              </span>
              <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>Recording…</span>
            </div>

            <button
              onClick={stopRecording}
              title="Stop and attach"
              style={{ flexShrink: 0, width: 32, height: 32, borderRadius: '50%', background: 'var(--accent)', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
            >
              <Square size={12} color="#fff" fill="#fff" />
            </button>
          </>
        ) : (
          <>
            {/* Emoji */}
            <button
              onClick={() => setShowEmojiPicker((v) => !v)}
              title="Emoji"
              style={{ flexShrink: 0, background: 'none', border: 'none', cursor: 'pointer', padding: '2px 4px', display: 'flex', alignItems: 'center', color: showEmojiPicker ? 'var(--accent)' : 'var(--text-muted)', transition: 'color 0.15s' }}
            >
              <Smile size={20} />
            </button>

            {/* Textarea */}
            <textarea
              ref={textareaRef}
              value={text}
              onChange={handleChange}
              onKeyDown={handleKeyDown}
              placeholder={isEditMode ? 'Edit message…' : 'Message...'}
              rows={1}
              style={{ flex: 1, resize: 'none', border: 'none', background: 'transparent', color: 'var(--text)', fontSize: 14, fontFamily: 'inherit', lineHeight: 1.45, outline: 'none', overflowY: 'auto', maxHeight: 120, paddingTop: 2, paddingBottom: 2 }}
            />

            {isEditMode ? (
              /* Edit mode: Cancel + Save */
              <div style={{ display: 'flex', alignItems: 'center', gap: 4, flexShrink: 0 }}>
                <button
                  onClick={handleCancelEdit}
                  title="Cancel"
                  style={{ width: 30, height: 30, borderRadius: '50%', background: 'var(--bg)', border: '1px solid var(--border)', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
                >
                  <X size={14} style={{ color: 'var(--text-muted)' }} />
                </button>
                <button
                  onClick={handleSaveEdit}
                  disabled={!canSend}
                  title="Save"
                  style={{ width: 30, height: 30, borderRadius: '50%', background: canSend ? 'var(--accent)' : 'var(--border)', border: 'none', cursor: canSend ? 'pointer' : 'default', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'background 0.15s' }}
                >
                  <Check size={14} color="#fff" />
                </button>
              </div>
            ) : canSend ? (
              <button
                onClick={handleSend}
                title="Send"
                style={{ flexShrink: 0, width: 32, height: 32, borderRadius: '50%', background: 'var(--accent)', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'opacity 0.15s' }}
              >
                <Send size={15} color="#fff" />
              </button>
            ) : (
              <div style={{ display: 'flex', alignItems: 'center', gap: 2, flexShrink: 0 }}>
                <button
                  onClick={startRecording}
                  title="Record voice note"
                  style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '2px 4px', display: 'flex', alignItems: 'center', color: 'var(--text-muted)' }}
                >
                  <Mic size={20} />
                </button>
                <button
                  onClick={() => fileInputRef.current?.click()}
                  disabled={compressing}
                  title={compressing ? 'Compressing…' : 'Send image, video or GIF'}
                  style={{ background: 'none', border: 'none', cursor: compressing ? 'not-allowed' : 'pointer', padding: '2px 4px', display: 'flex', alignItems: 'center', color: compressing ? 'var(--accent)' : 'var(--text-muted)', opacity: compressing ? 0.6 : 1 }}
                >
                  <ImageIcon size={20} />
                </button>
                <button
                  onClick={() => { setShowGifPicker((v) => !v); setShowEmojiPicker(false); }}
                  title="GIF"
                  style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '2px 4px', display: 'flex', alignItems: 'center', color: showGifPicker ? 'var(--accent)' : 'var(--text-muted)', transition: 'color 0.15s' }}
                >
                  <Sticker size={20} />
                </button>
              </div>
            )}
          </>
        )}
      </div>

      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp,video/mp4,video/quicktime,video/x-msvideo,video/x-matroska,video/webm"
        multiple
        style={{ display: 'none' }}
        onChange={handleFileChange}
      />
    </div>
  );
}
