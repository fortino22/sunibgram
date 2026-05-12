'use client';

import { useEffect, useRef, useState } from 'react';
import { MessageBubble } from './MessageBubble';
import { socketManager } from '@/lib/socket';
import type { ChatMessage, TypingPayload } from '@/lib/services/chat';

interface MessageListProps {
  messages: ChatMessage[];
  currentUserId: number;
  isLoading: boolean;
  roomId: number;
}

export function MessageList({ messages, currentUserId, isLoading, roomId }: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null);
  const [typingUsers, setTypingUsers] = useState<Map<number, string>>(new Map());

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages.length]);

  // Listen to typing events
  useEffect(() => {
    const handleTypingStart = (data: unknown) => {
      const payload = data as TypingPayload;
      if (payload.room_id !== roomId || payload.user_id === currentUserId) return;
      setTypingUsers((prev) => {
        const next = new Map(prev);
        next.set(payload.user_id, `User ${payload.user_id}`);
        return next;
      });
    };

    const handleTypingStop = (data: unknown) => {
      const payload = data as TypingPayload;
      if (payload.room_id !== roomId) return;
      setTypingUsers((prev) => {
        const next = new Map(prev);
        next.delete(payload.user_id);
        return next;
      });
    };

    const unsubStart = socketManager.on('typing:start', handleTypingStart);
    const unsubStop = socketManager.on('typing:stop', handleTypingStop);
    return () => {
      unsubStart();
      unsubStop();
    };
  }, [roomId, currentUserId]);

  // Messages come newest-first from API; reverse for display (oldest at top)
  const displayMessages = [...messages].reverse();

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
        const showSender =
          !prevMsg ||
          prevMsg.sender.id !== msg.sender.id ||
          new Date(msg.created_at).getTime() - new Date(prevMsg.created_at).getTime() > 300_000;

        return (
          <MessageBubble
            key={msg.chat_id}
            message={msg}
            isOwn={msg.sender.id === currentUserId}
            showSender={showSender}
          />
        );
      })}

      {/* Typing indicator */}
      {typingUsers.size > 0 && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            padding: '6px 0',
          }}
        >
          <div
            style={{
              padding: '6px 12px',
              borderRadius: '14px 14px 14px 4px',
              background: 'var(--bg)',
              border: '1px solid var(--border)',
              display: 'flex',
              gap: 4,
              alignItems: 'center',
            }}
          >
            {[0, 1, 2].map((i) => (
              <span
                key={i}
                style={{
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  background: 'var(--text-muted)',
                  display: 'inline-block',
                  animation: `bounce 1s ease-in-out ${i * 0.15}s infinite`,
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
