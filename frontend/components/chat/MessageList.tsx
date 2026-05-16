'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { MessageBubble } from './MessageBubble';
import type { ChatMessage } from '@/lib/services/chat';

interface MessageListProps {
  messages: ChatMessage[];
  currentUserId: number;
  isLoading: boolean;
  roomId: number;
  typingUsers?: string[];
  isPrivateRoom?: boolean;
  onEditMessage?: (message: ChatMessage) => void;
  onDeleteMessage?: (chatId: number) => void;
  onReplyMessage?: (message: ChatMessage) => void;
  onReact?: (chatId: number, emoji: string) => void;
  onUnreact?: (chatId: number, emoji: string) => void;
  onSwapReact?: (chatId: number, oldEmoji: string, newEmoji: string) => void;
}

export function MessageList({
  messages,
  currentUserId,
  isLoading,
  typingUsers = [],
  isPrivateRoom = false,
  onEditMessage,
  onDeleteMessage,
  onReplyMessage,
  onReact,
  onUnreact,
  onSwapReact,
}: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null);
  const msgRefs = useRef<Map<number, HTMLDivElement>>(new Map());
  const [highlightedId, setHighlightedId] = useState<number | null>(null);
  const highlightTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages.length]);

  const scrollToMessage = useCallback((chatId: number) => {
    const el = msgRefs.current.get(chatId);
    if (!el) return;
    el.scrollIntoView({ behavior: 'smooth', block: 'center' });
    if (highlightTimer.current) clearTimeout(highlightTimer.current);
    setHighlightedId(chatId);
    highlightTimer.current = setTimeout(() => setHighlightedId(null), 1000);
  }, []);

  // Messages come newest-first from API; reverse for display (oldest at top)
  const displayMessages = [...messages].reverse();

  function sameMinute(a: string, b: string) {
    const da = new Date(a);
    const db = new Date(b);
    return (
      da.getFullYear() === db.getFullYear() &&
      da.getMonth() === db.getMonth() &&
      da.getDate() === db.getDate() &&
      da.getHours() === db.getHours() &&
      da.getMinutes() === db.getMinutes()
    );
  }

  if (isLoading) {
    return (
      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '16px 20px',
          display: 'flex',
          flexDirection: 'column',
          gap: 8,
        }}
      >
        {Array.from({ length: 6 }, (_, i) => (
          <div
            key={i}
            style={{
              height: 40,
              borderRadius: 14,
              width: `${40 + (i % 3) * 20}%`,
              alignSelf: i % 2 === 0 ? 'flex-start' : 'flex-end',
              background: 'var(--skeleton, oklch(16% 0.005 260))',
              animation: 'pulse 1.5s ease-in-out infinite',
            }}
          />
        ))}
      </div>
    );
  }

  if (displayMessages.length === 0) {
    return (
      <div
        style={{
          flex: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'var(--text-muted)',
          fontSize: 13,
        }}
      >
        No messages yet. Say hello!
      </div>
    );
  }

  return (
    <div
      style={{
        flex: 1,
        overflowY: 'auto',
        padding: '16px 20px',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {displayMessages.map((msg, idx) => {
        const prevMsg = displayMessages[idx - 1];
        const nextMsg = displayMessages[idx + 1];

        const showSender =
          !prevMsg ||
          prevMsg.sender.id !== msg.sender.id ||
          new Date(msg.created_at).getTime() - new Date(prevMsg.created_at).getTime() > 300_000;

        const isLastInGroup =
          !nextMsg ||
          nextMsg.sender.id !== msg.sender.id ||
          !sameMinute(msg.created_at, nextMsg.created_at);

        const isOwn = msg.sender.id === currentUserId;

        return (
          <MessageBubble
            key={msg.chat_id}
            message={msg}
            isOwn={isOwn}
            showSender={showSender}
            isLastInGroup={isLastInGroup}
            isPrivateRoom={isPrivateRoom}
            currentUserId={currentUserId}
            isHighlighted={highlightedId === msg.chat_id}
            msgRef={(el) => {
              if (el) msgRefs.current.set(msg.chat_id, el);
              else msgRefs.current.delete(msg.chat_id);
            }}
            onScrollToMessage={scrollToMessage}
            onEdit={onEditMessage}
            onDelete={onDeleteMessage}
            onReply={onReplyMessage}
            onReact={onReact}
            onUnreact={onUnreact}
            onSwapReact={onSwapReact}
          />
        );
      })}

      {/* Typing indicator */}
      {typingUsers.length > 0 && (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '4px 0 8px' }}>
          <div
            style={{
              padding: '10px 16px',
              borderRadius: '18px 18px 18px 4px',
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              boxShadow: '0 1px 4px rgba(0,0,0,0.18)',
              display: 'flex',
              gap: 5,
              alignItems: 'center',
            }}
          >
            {[0, 1, 2].map((i) => (
              <span
                key={i}
                style={{
                  width: 8,
                  height: 8,
                  borderRadius: '50%',
                  background: 'var(--accent)',
                  display: 'inline-block',
                  animation: `bounce 1.1s ease-in-out ${i * 0.18}s infinite`,
                }}
              />
            ))}
          </div>
        </div>
      )}

      <div ref={bottomRef} />
    </div>
  );
}
