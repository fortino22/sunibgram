'use client';

import Image from 'next/image';
import { useState, useRef, useEffect } from 'react';
import { Check, CheckCheck, X, ChevronLeft, ChevronRight, Images, Play, Pause, Pencil, Trash2, MoreHorizontal, Reply, Smile } from 'lucide-react';
import EmojiPicker from 'emoji-picker-react';
import { Avatar } from '@/components/shared/Avatar';
import { SharedThreadEmbed } from '@/components/chat/SharedThreadEmbed';
import type { ChatMessage, ChatReactionSummary } from '@/lib/services/chat';

const QUICK_REACTIONS = ['👍', '❤️', '😂', '😮', '😢', '👎'];

interface MessageBubbleProps {
  message: ChatMessage;
  isOwn: boolean;
  showSender: boolean;
  isLastInGroup: boolean;
  isPrivateRoom?: boolean;
  currentUserId?: number;
  isHighlighted?: boolean;
  msgRef?: (el: HTMLDivElement | null) => void;
  onScrollToMessage?: (chatId: number) => void;
  onEdit?: (message: ChatMessage) => void;
  onDelete?: (chatId: number) => void;
  onReply?: (message: ChatMessage) => void;
  onReact?: (chatId: number, emoji: string) => void;
  onUnreact?: (chatId: number, emoji: string) => void;
  onSwapReact?: (chatId: number, oldEmoji: string, newEmoji: string) => void;
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

// ── Lightbox ──────────────────────────────────────────────────────────────────

interface LightboxProps {
  urls: string[];
  initialIndex: number;
  onClose: () => void;
}

function Lightbox({ urls, initialIndex, onClose }: LightboxProps) {
  const [index, setIndex] = useState(initialIndex);

  const prev = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIndex((i) => (i - 1 + urls.length) % urls.length);
  };
  const next = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIndex((i) => (i + 1) % urls.length);
  };

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 1000,
        background: 'rgba(0,0,0,0.88)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <button
        onClick={onClose}
        style={{
          position: 'absolute',
          top: 16,
          right: 16,
          background: 'rgba(255,255,255,0.15)',
          border: 'none',
          borderRadius: '50%',
          width: 36,
          height: 36,
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#fff',
        }}
      >
        <X size={18} />
      </button>

      {urls.length > 1 && (
        <div
          style={{
            position: 'absolute',
            top: 20,
            left: '50%',
            transform: 'translateX(-50%)',
            color: '#fff',
            fontSize: 13,
            fontWeight: 500,
            background: 'rgba(0,0,0,0.5)',
            padding: '3px 10px',
            borderRadius: 20,
          }}
        >
          {index + 1} / {urls.length}
        </div>
      )}

      {urls.length > 1 && (
        <button
          onClick={prev}
          style={{
            position: 'absolute',
            left: 16,
            background: 'rgba(255,255,255,0.15)',
            border: 'none',
            borderRadius: '50%',
            width: 40,
            height: 40,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
          }}
        >
          <ChevronLeft size={22} />
        </button>
      )}

      <div
        onClick={(e) => e.stopPropagation()}
        style={{ maxWidth: '90vw', maxHeight: '85vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      >
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={urls[index]}
          alt={`media ${index + 1}`}
          style={{ maxWidth: '90vw', maxHeight: '85vh', objectFit: 'contain', borderRadius: 8 }}
        />
      </div>

      {urls.length > 1 && (
        <button
          onClick={next}
          style={{
            position: 'absolute',
            right: 16,
            background: 'rgba(255,255,255,0.15)',
            border: 'none',
            borderRadius: '50%',
            width: 40,
            height: 40,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
          }}
        >
          <ChevronRight size={22} />
        </button>
      )}

      {urls.length > 1 && (
        <div
          style={{
            position: 'absolute',
            bottom: 20,
            left: '50%',
            transform: 'translateX(-50%)',
            display: 'flex',
            gap: 6,
          }}
        >
          {urls.map((_, i) => (
            <button
              key={i}
              onClick={(e) => { e.stopPropagation(); setIndex(i); }}
              style={{
                width: i === index ? 20 : 8,
                height: 8,
                borderRadius: 4,
                background: i === index ? '#fff' : 'rgba(255,255,255,0.4)',
                border: 'none',
                cursor: 'pointer',
                padding: 0,
                transition: 'width 0.15s',
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
}

// ── Voice note ────────────────────────────────────────────────────────────────

const WAVEFORM = [3, 6, 10, 14, 10, 6, 14, 18, 14, 10, 6, 10, 14, 18, 22, 18, 14, 10, 6, 3];

function formatDuration(seconds: number): string {
  if (!isFinite(seconds) || isNaN(seconds) || seconds < 0) return '0:00';
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, '0')}`;
}

interface VoiceNoteBubbleProps {
  url: string;
  isOwn: boolean;
}

function VoiceNoteBubble({ url, isOwn }: VoiceNoteBubbleProps) {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [playing, setPlaying] = useState(false);
  const [progress, setProgress] = useState(0);
  const [duration, setDuration] = useState(0);
  const [currentTime, setCurrentTime] = useState(0);
  const [hasError, setHasError] = useState(false);

  useEffect(() => {
    setPlaying(false);
    setProgress(0);
    setCurrentTime(0);
    setDuration(0);
    setHasError(false);
    const audio = audioRef.current;
    if (audio) audio.load();
  }, [url]);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;
    const trySetDuration = () => {
      const d = audio.duration;
      if (isFinite(d) && !isNaN(d) && d > 0) setDuration(d);
    };
    const onTime = () => {
      setCurrentTime(audio.currentTime);
      if (isFinite(audio.duration)) setProgress(audio.currentTime / audio.duration);
    };
    const onEnded = () => { setPlaying(false); setProgress(0); setCurrentTime(0); };
    const onEmptied = () => { setPlaying(false); setProgress(0); setCurrentTime(0); };
    const onError = () => { setPlaying(false); setHasError(true); };
    audio.addEventListener('loadedmetadata', trySetDuration);
    audio.addEventListener('durationchange', trySetDuration);
    audio.addEventListener('timeupdate', onTime);
    audio.addEventListener('ended', onEnded);
    audio.addEventListener('emptied', onEmptied);
    audio.addEventListener('error', onError);
    return () => {
      audio.removeEventListener('loadedmetadata', trySetDuration);
      audio.removeEventListener('durationchange', trySetDuration);
      audio.removeEventListener('timeupdate', onTime);
      audio.removeEventListener('ended', onEnded);
      audio.removeEventListener('emptied', onEmptied);
      audio.removeEventListener('error', onError);
    };
  }, []);

  const toggle = () => {
    const audio = audioRef.current;
    if (!audio || !url) return;
    if (playing) {
      audio.pause();
      setPlaying(false);
    } else {
      if (audio.error || audio.networkState === HTMLMediaElement.NETWORK_NO_SOURCE) {
        setHasError(false);
        audio.load();
      }
      audio.play()
        .then(() => setPlaying(true))
        .catch(() => { setPlaying(false); setHasError(true); });
    }
  };

  const displayTime = playing || currentTime > 0 ? formatDuration(currentTime) : formatDuration(duration);
  const playedBars = Math.round(progress * WAVEFORM.length);

  if (!url) return null;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4, alignItems: isOwn ? 'flex-end' : 'flex-start' }}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          padding: '10px 12px',
          borderRadius: 20,
          background: hasError ? 'var(--bg)' : 'var(--accent)',
          border: hasError ? '1px solid var(--border)' : 'none',
          minWidth: 200,
          maxWidth: 260,
        }}
      >
        <audio ref={audioRef} src={url} preload="none" style={{ display: 'none' }} />

        <button
          onClick={hasError ? () => { setHasError(false); toggle(); } : toggle}
          style={{
            width: 34,
            height: 34,
            borderRadius: '50%',
            background: hasError ? 'var(--border)' : 'rgba(255,255,255,0.18)',
            border: 'none',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}
        >
          {playing
            ? <Pause size={16} fill={hasError ? 'var(--text-muted)' : '#fff'} color={hasError ? 'var(--text-muted)' : '#fff'} />
            : <Play size={16} fill={hasError ? 'var(--text-muted)' : '#fff'} color={hasError ? 'var(--text-muted)' : '#fff'} style={{ marginLeft: 2 }} />}
        </button>

        <div style={{ flex: 1, display: 'flex', alignItems: 'center', gap: 2 }}>
          {hasError ? (
            <span style={{ fontSize: 11, color: 'var(--text-muted)', flex: 1 }}>Failed to load audio</span>
          ) : (
            <>
              {WAVEFORM.map((h, i) => (
                <div
                  key={i}
                  style={{
                    width: 3,
                    height: h,
                    borderRadius: 2,
                    flexShrink: 0,
                    background: i < playedBars ? 'rgba(255,255,255,0.95)' : 'rgba(255,255,255,0.35)',
                    transition: 'background 0.1s',
                  }}
                />
              ))}
              <div style={{ flex: 1, borderBottom: '2px dotted rgba(255,255,255,0.35)', marginBottom: 1, minWidth: 8 }} />
            </>
          )}
        </div>

        {!hasError && (
          <div style={{ background: 'rgba(255,255,255,0.18)', borderRadius: 12, padding: '2px 8px', flexShrink: 0 }}>
            <span style={{ color: '#fff', fontSize: 12, fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
              {displayTime || '0:00'}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}

// ── Quick-react button ────────────────────────────────────────────────────────

interface QuickReactButtonProps {
  chatId: number;
  isOwn: boolean;
  currentUserId?: number;
  reactions: ChatReactionSummary[];
  onReact?: (chatId: number, emoji: string) => void;
  onUnreact?: (chatId: number, emoji: string) => void;
  onSwapReact?: (chatId: number, oldEmoji: string, newEmoji: string) => void;
}

function QuickReactButton({ chatId, isOwn, currentUserId, reactions, onReact, onUnreact, onSwapReact }: QuickReactButtonProps) {
  const [open, setOpen] = useState(false);
  const [showPicker, setShowPicker] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open && !showPicker) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
        setShowPicker(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open, showPicker]);

  const myCurrentEmoji = reactions.find((r) => r.user_ids.includes(currentUserId ?? -1))?.emoji ?? null;

  const handleEmoji = (emoji: string) => {
    if (myCurrentEmoji === emoji) {
      onUnreact?.(chatId, emoji);
    } else if (myCurrentEmoji) {
      onSwapReact?.(chatId, myCurrentEmoji, emoji);
    } else {
      onReact?.(chatId, emoji);
    }
    setOpen(false);
    setShowPicker(false);
  };

  return (
    <div ref={ref} style={{ position: 'relative' }}>
      <button
        onClick={() => { setOpen((v) => !v); setShowPicker(false); }}
        title="React"
        style={{
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          padding: '2px 4px',
          borderRadius: 6,
          color: open ? 'var(--accent)' : 'var(--text-muted)',
          display: 'flex',
          alignItems: 'center',
          opacity: 0.8,
          transition: 'color 0.12s',
        }}
      >
        <Smile size={15} />
      </button>

      {open && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            [isOwn ? 'right' : 'left']: 0,
            zIndex: 100,
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 24,
            boxShadow: '0 4px 16px rgba(0,0,0,0.25)',
            padding: '6px 8px',
            display: 'flex',
            gap: 2,
            alignItems: 'center',
          }}
        >
          {QUICK_REACTIONS.map((emoji) => {
            const isSelected = myCurrentEmoji === emoji;
            return (
              <button
                key={emoji}
                onClick={() => handleEmoji(emoji)}
                style={{
                  background: isSelected ? 'color-mix(in oklch, var(--accent) 20%, transparent)' : 'none',
                  border: isSelected ? '1.5px solid var(--accent)' : '1.5px solid transparent',
                  borderRadius: '50%',
                  width: 32,
                  height: 32,
                  cursor: 'pointer',
                  fontSize: 18,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  transition: 'transform 0.1s, background 0.1s',
                }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1.25)'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1)'; }}
              >
                {emoji}
              </button>
            );
          })}

          {/* Custom emoji picker trigger */}
          <button
            onClick={() => setShowPicker((v) => !v)}
            title="More emojis"
            style={{
              background: showPicker ? 'color-mix(in oklch, var(--accent) 15%, transparent)' : 'none',
              border: '1.5px solid transparent',
              borderRadius: '50%',
              width: 32,
              height: 32,
              cursor: 'pointer',
              fontSize: 16,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'var(--text-muted)',
              transition: 'transform 0.1s, background 0.1s',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1.15)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.transform = 'scale(1)'; }}
          >
            ＋
          </button>
        </div>
      )}

      {showPicker && open && (
        <div
          style={{
            position: 'absolute',
            top: 'calc(100% + 40px)',
            [isOwn ? 'right' : 'left']: 0,
            zIndex: 101,
          }}
        >
          <EmojiPicker
            onEmojiClick={(data) => handleEmoji(data.emoji)}
            skinTonesDisabled
            searchDisabled={false}
            height={340}
            width={300}
          />
        </div>
      )}
    </div>
  );
}

// ── Context menu ──────────────────────────────────────────────────────────────

interface MessageMenuProps {
  isOwn: boolean;
  hasTextMessage: boolean;
  onEdit: () => void;
  onDelete: () => void;
  onReply: () => void;
}

function MessageMenu({ isOwn, hasTextMessage, onEdit, onDelete, onReply }: MessageMenuProps) {
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  return (
    <div ref={menuRef} style={{ position: 'relative', alignSelf: 'center' }}>
      <button
        onClick={() => setOpen((v) => !v)}
        style={{
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          padding: '2px 4px',
          borderRadius: 6,
          color: 'var(--text-muted)',
          display: 'flex',
          alignItems: 'center',
          opacity: 0.7,
        }}
      >
        <MoreHorizontal size={15} />
      </button>

      {open && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            [isOwn ? 'right' : 'left']: 0,
            zIndex: 100,
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 10,
            boxShadow: '0 4px 16px rgba(0,0,0,0.25)',
            minWidth: 130,
            overflow: 'hidden',
          }}
        >
          <button
            onClick={() => { setOpen(false); onReply(); }}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              width: '100%',
              padding: '9px 14px',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text)',
              fontSize: 13,
              fontFamily: 'inherit',
              textAlign: 'left',
            }}
          >
            <Reply size={13} />
            Reply
          </button>
          {isOwn && hasTextMessage && (
            <button
              onClick={() => { setOpen(false); onEdit(); }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                width: '100%',
                padding: '9px 14px',
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                color: 'var(--text)',
                fontSize: 13,
                fontFamily: 'inherit',
                textAlign: 'left',
              }}
            >
              <Pencil size={13} />
              Edit
            </button>
          )}
          <button
            onClick={() => { setOpen(false); onDelete(); }}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              width: '100%',
              padding: '9px 14px',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'oklch(60% 0.22 25)',
              fontSize: 13,
              fontFamily: 'inherit',
              textAlign: 'left',
            }}
          >
            <Trash2 size={13} />
            Delete
          </button>
        </div>
      )}
    </div>
  );
}

// ── Reply bubble (quoted message preview) ─────────────────────────────────────

interface ReplyBubbleProps {
  replyTo: NonNullable<ChatMessage['reply_to']>;
  isOwn: boolean;
  onScrollTo?: () => void;
}

function ReplyBubble({ replyTo, isOwn, onScrollTo }: ReplyBubbleProps) {
  const preview = replyTo.message
    ? replyTo.message.slice(0, 80) + (replyTo.message.length > 80 ? '…' : '')
    : replyTo.gif_url
    ? 'GIF'
    : 'Media';

  return (
    <div
      onClick={onScrollTo}
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 1,
        padding: '5px 10px',
        borderRadius: 10,
        borderLeft: `3px solid var(--accent)`,
        background: isOwn ? 'rgba(255,255,255,0.12)' : 'var(--bg)',
        border: isOwn ? undefined : '1px solid var(--border)',
        borderLeftColor: 'var(--accent)',
        borderLeftWidth: 3,
        borderLeftStyle: 'solid',
        marginBottom: 2,
        maxWidth: '100%',
        cursor: onScrollTo ? 'pointer' : 'default',
        transition: 'opacity 0.15s',
      }}
    >
      <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--accent)' }}>
        {replyTo.sender.username}
      </span>
      <span
        style={{
          fontSize: 12,
          color: isOwn ? 'rgba(255,255,255,0.7)' : 'var(--text-muted)',
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
        }}
      >
        {preview}
      </span>
    </div>
  );
}

// ── Reaction bar ──────────────────────────────────────────────────────────────

interface ReactionBarProps {
  reactions: ChatReactionSummary[];
  chatId: number;
  isOwn: boolean;
  currentUserId?: number;
  onReact?: (chatId: number, emoji: string) => void;
  onUnreact?: (chatId: number, emoji: string) => void;
  onSwapReact?: (chatId: number, oldEmoji: string, newEmoji: string) => void;
}

function ReactionBar({ reactions, chatId, isOwn, currentUserId, onReact, onUnreact, onSwapReact }: ReactionBarProps) {
  if (reactions.length === 0) return null;

  const myCurrentEmoji = reactions.find((r) => r.user_ids.includes(currentUserId ?? -1))?.emoji ?? null;

  const handleClick = (emoji: string) => {
    if (myCurrentEmoji === emoji) {
      onUnreact?.(chatId, emoji);
    } else if (myCurrentEmoji) {
      onSwapReact?.(chatId, myCurrentEmoji, emoji);
    } else {
      onReact?.(chatId, emoji);
    }
  };

  return (
    <div
      style={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: 4,
        marginTop: 4,
        justifyContent: isOwn ? 'flex-end' : 'flex-start',
      }}
    >
      {reactions.map((r) => {
        const isOwned = myCurrentEmoji === r.emoji;
        return (
          <button
            key={r.emoji}
            onClick={() => handleClick(r.emoji)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 3,
              padding: '2px 7px',
              borderRadius: 12,
              border: isOwned ? '1.5px solid var(--accent)' : '1px solid var(--border)',
              background: isOwned ? 'color-mix(in oklch, var(--accent) 15%, transparent)' : 'var(--bg)',
              cursor: 'pointer',
              fontSize: 13,
              fontFamily: 'inherit',
              color: isOwned ? 'var(--accent)' : 'var(--text)',
              fontWeight: isOwned ? 700 : 400,
              transition: 'border-color 0.12s, background 0.12s',
            }}
          >
            <span>{r.emoji}</span>
            <span style={{ fontSize: 11 }}>{r.count}</span>
          </button>
        );
      })}
    </div>
  );
}

// ── MessageBubble ─────────────────────────────────────────────────────────────

export function MessageBubble({
  message,
  isOwn,
  showSender,
  isLastInGroup,
  isPrivateRoom = false,
  currentUserId,
  isHighlighted = false,
  msgRef,
  onScrollToMessage,
  onEdit,
  onDelete,
  onReply,
  onReact,
  onUnreact,
  onSwapReact,
}: MessageBubbleProps) {
  const isDeleted = !!message.deleted_at;
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null);
  const [hovered, setHovered] = useState(false);

  const imageMedia = message.media.filter((m) => m.media_type.startsWith('image/'));
  const audioMedia = message.media.filter((m) => m.media_type.startsWith('audio/'));
  const videoMedia = message.media.filter((m) => m.media_type.startsWith('video/'));
  const imageUrls = imageMedia.map((m) => m.url);
  const extraCount = imageMedia.length - 1;

  const hasTextMessage = !!(message.message && message.message.trim().length > 0);
  const reactions = message.reactions ?? [];
  const showControls = !isDeleted && (onEdit || onDelete || onReply || onReact);

  // Bubble radius: tail only on the last message of a group
  const ownRadius = isLastInGroup ? '18px 18px 4px 18px' : '18px';
  const otherRadius = isLastInGroup ? '18px 18px 18px 4px' : '18px';

  return (
    <>
      {lightboxIndex !== null && (
        <Lightbox
          urls={imageUrls}
          initialIndex={lightboxIndex}
          onClose={() => setLightboxIndex(null)}
        />
      )}

      <div
        ref={msgRef}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
        style={{
          display: 'flex',
          alignItems: 'flex-end',
          justifyContent: isOwn ? 'flex-end' : 'flex-start',
          gap: 8,
          marginBottom: isLastInGroup ? 10 : 2,
          padding: isHighlighted ? '4px 8px' : '0 8px',
          animation: isHighlighted ? 'msg-highlight 1s ease-in-out' : undefined,
        }}
      >
        {/* Avatar slot — always reserves 28 px for non-own to keep bubbles aligned */}
        {!isOwn && (
          <div style={{ width: 28, flexShrink: 0, alignSelf: 'flex-end' }}>
            {isLastInGroup && (
              <Avatar username={message.sender.username} size={28} />
            )}
          </div>
        )}

        {/* Inner row: bubble content + hover controls */}
        {/* row-reverse for own → controls end up to the LEFT of bubble */}
        <div
          style={{
            display: 'flex',
            flexDirection: isOwn ? 'row-reverse' : 'row',
            alignItems: 'center',
            gap: 6,
            maxWidth: '68%',
          }}
        >
          {/* ── Message content column ── */}
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: isOwn ? 'flex-end' : 'flex-start',
            }}
          >
            {/* Reply preview */}
            {message.reply_to && !isDeleted && (
              <ReplyBubble
                replyTo={message.reply_to}
                isOwn={isOwn}
                onScrollTo={onScrollToMessage ? () => onScrollToMessage(message.reply_to!.chat_id) : undefined}
              />
            )}

            {/* Images */}
            {imageMedia.length > 0 && !isDeleted && (
              <div
                style={{
                  position: 'relative',
                  borderRadius: 12,
                  overflow: 'hidden',
                  marginBottom: message.message || audioMedia.length > 0 || videoMedia.length > 0 ? 4 : 0,
                  cursor: 'pointer',
                  flexShrink: 0,
                }}
                onClick={() => setLightboxIndex(0)}
              >
                <Image
                  src={imageMedia[0].url}
                  alt="media"
                  width={220}
                  height={160}
                  style={{ objectFit: 'cover', display: 'block' }}
                  unoptimized
                />
                {extraCount > 0 && (
                  <div
                    style={{
                      position: 'absolute',
                      bottom: 6,
                      left: 6,
                      display: 'flex',
                      alignItems: 'center',
                      gap: 3,
                      background: 'rgba(0,0,0,0.55)',
                      borderRadius: 6,
                      padding: '2px 6px 2px 4px',
                      backdropFilter: 'blur(4px)',
                    }}
                  >
                    <Images size={13} color="#fff" />
                    <span style={{ color: '#fff', fontSize: 11, fontWeight: 600, lineHeight: 1 }}>
                      +{extraCount}
                    </span>
                  </div>
                )}
              </div>
            )}

            {/* Shared thread embed */}
            {message.shared_thread && !isDeleted && (
              <div style={{ marginBottom: message.message || message.gif_url || audioMedia.length > 0 || videoMedia.length > 0 ? 4 : 0 }}>
                <SharedThreadEmbed embed={message.shared_thread} isOwn={isOwn} />
              </div>
            )}

            {/* GIF */}
            {message.gif_url && !isDeleted && (
              <div
                style={{
                  borderRadius: 12,
                  overflow: 'hidden',
                  marginBottom: message.message || audioMedia.length > 0 || videoMedia.length > 0 ? 4 : 0,
                  flexShrink: 0,
                  cursor: 'pointer',
                  maxWidth: 220,
                }}
                onClick={() => window.open(message.gif_url!, '_blank', 'noopener')}
              >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={message.gif_url}
                  alt="GIF"
                  style={{ width: '100%', maxWidth: 220, display: 'block' }}
                />
              </div>
            )}

            {/* Voice notes */}
            {audioMedia.length > 0 && !isDeleted && audioMedia.map((m) => (
              <VoiceNoteBubble key={m.media_id} url={m.url} isOwn={isOwn} />
            ))}

            {/* Videos */}
            {videoMedia.length > 0 && !isDeleted && (
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, marginBottom: message.message ? 4 : 0, justifyContent: isOwn ? 'flex-end' : 'flex-start' }}>
                {videoMedia.map((m) => (
                  <div key={m.media_id} style={{ borderRadius: 10, overflow: 'hidden', maxWidth: 220, background: 'var(--bg)' }}>
                    <video src={m.url} controls style={{ maxWidth: 220, maxHeight: 160, display: 'block' }} />
                  </div>
                ))}
              </div>
            )}

            {/* Text bubble */}
            {(message.message || isDeleted) && (
              <div
                style={{
                  padding: '9px 13px',
                  borderRadius: isOwn ? ownRadius : otherRadius,
                  background: isDeleted
                    ? 'transparent'
                    : isOwn
                    ? 'var(--accent)'
                    : 'oklch(28% 0.01 260)',
                  border: isDeleted ? '1px dashed var(--border)' : 'none',
                  fontSize: 14,
                  lineHeight: 1.5,
                  color: isDeleted ? 'var(--text-muted)' : 'var(--text)',
                  fontStyle: isDeleted ? 'italic' : 'normal',
                  wordBreak: 'break-word',
                  whiteSpace: 'pre-wrap',
                }}
              >
                {isDeleted ? 'This message was deleted' : message.message}
              </div>
            )}

            {/* Reaction bar */}
            {!isDeleted && (
              <ReactionBar
                reactions={reactions}
                chatId={message.chat_id}
                isOwn={isOwn}
                currentUserId={currentUserId}
                onReact={onReact}
                onUnreact={onUnreact}
                onSwapReact={onSwapReact}
              />
            )}

            {/* Timestamp + edited + read receipt — only on last bubble of group */}
            {isLastInGroup && (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 4,
                  marginTop: 4,
                  flexDirection: isOwn ? 'row-reverse' : 'row',
                }}
              >
                <span style={{ fontSize: 10, color: 'var(--text-muted)' }}>
                  {formatTime(message.created_at)}
                </span>
                {message.is_edited && !isDeleted && (
                  <span style={{ fontSize: 10, color: 'var(--text-muted)', fontStyle: 'italic' }}>
                    edited
                  </span>
                )}
                {isOwn && !isDeleted && (
                  message.is_read
                    ? <CheckCheck size={12} style={{ color: 'var(--accent)' }} />
                    : <Check size={12} style={{ color: 'var(--text-muted)' }} />
                )}
              </div>
            )}
          </div>

          {/* ── Hover controls: emoji | reply | dots ── */}
          {showControls && (
            <div
              style={{
                visibility: hovered ? 'visible' : 'hidden',
                display: 'flex',
                alignItems: 'center',
                gap: 2,
                flexShrink: 0,
              }}
            >
              {(onReact || onUnreact) && (
                <QuickReactButton
                  chatId={message.chat_id}
                  isOwn={isOwn}
                  currentUserId={currentUserId}
                  reactions={reactions}
                  onReact={onReact}
                  onUnreact={onUnreact}
                  onSwapReact={onSwapReact}
                />
              )}
              {onReply && (
                <button
                  onClick={() => onReply(message)}
                  title="Reply"
                  style={{
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    padding: '2px 4px',
                    borderRadius: 6,
                    color: 'var(--text-muted)',
                    display: 'flex',
                    alignItems: 'center',
                    opacity: 0.8,
                  }}
                >
                  <Reply size={15} />
                </button>
              )}
              {(onEdit || onDelete || onReply) && (
                <MessageMenu
                  isOwn={isOwn}
                  hasTextMessage={hasTextMessage}
                  onEdit={() => onEdit?.(message)}
                  onDelete={() => onDelete?.(message.chat_id)}
                  onReply={() => onReply?.(message)}
                />
              )}
            </div>
          )}
        </div>
      </div>
    </>
  );
}
