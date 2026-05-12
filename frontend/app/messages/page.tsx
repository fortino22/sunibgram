'use client';

import { useEffect, useMemo, useState } from 'react';
import { MessageSquare } from 'lucide-react';
import { ConversationList } from '@/components/chat/ConversationList';
import { ChatHeader } from '@/components/chat/ChatHeader';
import { MessageList } from '@/components/chat/MessageList';
import { ChatInput } from '@/components/chat/ChatInput';
import { useSocket } from '@/hooks/useSocket';
import { useChatMessages } from '@/hooks/chat/useChatMessages';
import { useChatRoom } from '@/hooks/chat/useChatRooms';
import { useChatStore } from '@/stores/chatStore';
import { useAuthStore } from '@/stores/authStore';
import { socketManager } from '@/lib/socket';
import type { TypingPayload } from '@/lib/services/chat';

export default function MessagesPage() {
  const user = useAuthStore((s) => s.user);
  const { activeRoomId, setActiveRoomId } = useChatStore();
  const [typingUsers, setTypingUsers] = useState<string[]>([]);

  useSocket();

  const messagesQuery = useChatMessages(activeRoomId);
  const activeRoom = useChatRoom(activeRoomId);
  const messages = messagesQuery.data?.data ?? [];

  // Determine the "other" user in private rooms from message history
  const otherUsername = useMemo(() => {
    if (!activeRoom || activeRoom.type !== 'private' || !user) return undefined;
    const otherMsg = messages.find((m) => m.sender.id !== user.id);
    return otherMsg?.sender.username;
  }, [activeRoom, messages, user]);

  // Typing indicators for the active room
  useEffect(() => {
    if (!activeRoomId || !user) return;

    const usernames = new Map<number, string>();

    const handleStart = (data: unknown) => {
      const payload = data as TypingPayload;
      if (payload.room_id !== activeRoomId || payload.user_id === user.id) return;
      usernames.set(payload.user_id, `User ${payload.user_id}`);
      setTypingUsers([...usernames.values()]);
    };

    const handleStop = (data: unknown) => {
      const payload = data as TypingPayload;
      if (payload.room_id !== activeRoomId) return;
      usernames.delete(payload.user_id);
      setTypingUsers([...usernames.values()]);
    };

    const unsubStart = socketManager.on('typing:start', handleStart);
    const unsubStop = socketManager.on('typing:stop', handleStop);
    return () => {
      unsubStart();
      unsubStop();
    };
  }, [activeRoomId, user]);

  return (
    <div
      style={{
        display: 'flex',
        height: '100vh',
        overflow: 'hidden',
      }}
    >
      {/* Left — Conversation list */}
      <div style={{ width: 280, flexShrink: 0 }}>
        <ConversationList
          activeRoomId={activeRoomId}
          onRoomSelect={setActiveRoomId}
        />
      </div>

      {/* Right — Active conversation */}
      <div
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
          minWidth: 0,
        }}
      >
        {activeRoom ? (
          <>
            <ChatHeader
              room={activeRoom}
              otherUsername={otherUsername}
              typingUsers={typingUsers}
            />
            <MessageList
              messages={messages}
              currentUserId={user?.id ?? 0}
              isLoading={messagesQuery.isLoading}
              roomId={activeRoomId!}
            />
            <ChatInput roomId={activeRoomId!} />
          </>
        ) : (
          <div
            style={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'var(--text-muted)',
              gap: 12,
            }}
          >
            <MessageSquare size={40} color="var(--text-muted)" strokeWidth={1.2} />
            <div style={{ fontSize: 16, fontWeight: 500, color: 'var(--text)' }}>
              Your messages
            </div>
            <div style={{ fontSize: 13, textAlign: 'center', maxWidth: 240 }}>
              Select a conversation on the left, or start a new one with the + button.
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
