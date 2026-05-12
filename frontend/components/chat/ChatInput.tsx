'use client';

import { useRef, useState, useCallback } from 'react';
import { Send, Paperclip, X } from 'lucide-react';
import { socketManager } from '@/lib/socket';
import { useSendMessage } from '@/hooks/chat/useSendMessage';

interface ChatInputProps {
  roomId: number;
}

const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'video/mp4', 'video/quicktime'];

export function ChatInput({ roomId }: ChatInputProps) {
  const [text, setText] = useState('');
  const [files, setFiles] = useState<File[]>([]);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const typingTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isTyping = useRef(false);
  const sendMessage = useSendMessage();

  const startTyping = useCallback(() => {
    if (!isTyping.current) {
      isTyping.current = true;
      socketManager.emit('typing:start', roomId);
    }
    if (typingTimer.current) clearTimeout(typingTimer.current);
    typingTimer.current = setTimeout(() => {
      isTyping.current = false;
      socketManager.emit('typing:stop', roomId);
    }, 2000);
  }, [roomId]);

  const stopTyping = useCallback(() => {
    if (typingTimer.current) clearTimeout(typingTimer.current);
    if (isTyping.current) {
      isTyping.current = false;
      socketManager.emit('typing:stop', roomId);
    }
  }, [roomId]);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value);
    // Auto-resize
    const el = e.target;
    el.style.height = 'auto';
    el.style.height = `${Math.min(el.scrollHeight, 120)}px`;
    if (e.target.value) {
      startTyping();
    } else {
      stopTyping();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const picked = Array.from(e.target.files ?? []).filter((f) =>
      ALLOWED_TYPES.includes(f.type)
    );
    setFiles((prev) => [...prev, ...picked]);
    e.target.value = '';
  };

  const removeFile = (idx: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== idx));
  };

  const handleSend = () => {
    const trimmed = text.trim();
    if (!trimmed && files.length === 0) return;
    stopTyping();

    sendMessage.mutate(
      { roomId, message: trimmed || null, files },
      {
        onSuccess: () => {
          setText('');
          setFiles([]);
          if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
          }
        },
      }
    );
  };

  const canSend = (text.trim().length > 0 || files.length > 0) && !sendMessage.isPending;

  return (
    <div
      style={{
        borderTop: '1px solid var(--border)',
        padding: '12px 16px',
        flexShrink: 0,
        background: 'var(--bg-card)',
      }}
    >
      {/* File previews */}
      {files.length > 0 && (
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: 8,
            marginBottom: 10,
          }}
        >
          {files.map((file, idx) => (
            <div
              key={idx}
              style={{
                position: 'relative',
                borderRadius: 8,
                overflow: 'hidden',
                border: '1px solid var(--border)',
              }}
            >
              {file.type.startsWith('image/') ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={URL.createObjectURL(file)}
                  alt={file.name}
                  style={{ width: 72, height: 72, objectFit: 'cover', display: 'block' }}
                />
              ) : (
                <div
                  style={{
                    width: 72,
                    height: 72,
                    background: 'var(--bg)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 10,
                    color: 'var(--text-muted)',
                    padding: 4,
                    textAlign: 'center',
                  }}
                >
                  {file.name}
                </div>
              )}
              <button
                onClick={() => removeFile(idx)}
                style={{
                  position: 'absolute',
                  top: 2,
                  right: 2,
                  width: 18,
                  height: 18,
                  borderRadius: '50%',
                  background: 'rgba(0,0,0,0.6)',
                  border: 'none',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  padding: 0,
                }}
              >
                <X size={10} color="#fff" />
              </button>
            </div>
          ))}
        </div>
      )}

      <div style={{ display: 'flex', alignItems: 'flex-end', gap: 8 }}>
        {/* File attach */}
        <button
          onClick={() => fileInputRef.current?.click()}
          title="Attach file"
          style={{
            flexShrink: 0,
            width: 36,
            height: 36,
            borderRadius: 10,
            background: 'var(--bg)',
            border: '1px solid var(--border)',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <Paperclip size={16} color="var(--text-muted)" />
        </button>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp,video/mp4,video/quicktime"
          multiple
          style={{ display: 'none' }}
          onChange={handleFileChange}
        />

        {/* Text input */}
        <textarea
          ref={textareaRef}
          value={text}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder="Message..."
          rows={1}
          style={{
            flex: 1,
            resize: 'none',
            border: '1px solid var(--border)',
            borderRadius: 12,
            padding: '8px 12px',
            background: 'var(--bg)',
            color: 'var(--text)',
            fontSize: 14,
            fontFamily: 'inherit',
            lineHeight: 1.45,
            outline: 'none',
            overflowY: 'auto',
            maxHeight: 120,
          }}
        />

        {/* Send button */}
        <button
          onClick={handleSend}
          disabled={!canSend}
          title="Send"
          style={{
            flexShrink: 0,
            width: 36,
            height: 36,
            borderRadius: 10,
            background: canSend ? 'var(--accent)' : 'var(--bg)',
            border: canSend ? 'none' : '1px solid var(--border)',
            cursor: canSend ? 'pointer' : 'not-allowed',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'background 0.15s ease',
          }}
        >
          <Send size={16} color={canSend ? '#fff' : 'var(--text-muted)'} />
        </button>
      </div>
    </div>
  );
}
