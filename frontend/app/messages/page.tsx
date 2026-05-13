'use client';

import { useEffect, useMemo } from 'react';
import { MessageSquare } from 'lucide-react';
import { ConversationList } from '@/components/chat/ConversationList';
import { ChatHeader } from '@/components/chat/ChatHeader';
import { MessageList } from '@/components/chat/MessageList';
import { ChatInput } from '@/components/chat/ChatInput';
import { useChatSocket } from '@/hooks/useChatSocket';
import { useChatMessages } from '@/hooks/chat/useChatMessages';
import { useChatRoom } from '@/hooks/chat/useChatRooms';
import { useMarkMessagesRead } from '@/hooks/chat/useMarkMessagesRead';
import { useChatStore } from '@/stores/chatStore';
import { useAuthStore } from '@/stores/authStore';

export default function MessagesPage() {
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const { activeRoomId, setActiveRoomId } = useChatStore();

  // Single WebSocket connection for the entire messages page
  const { joinRoom, leaveRoom, startTyping, stopTyping, typingUsersByRoom } = useChatSocket({
    accessToken: token,
    currentUserId: user?.id,
  });

  const messagesQuery = useChatMessages(activeRoomId);
  const activeRoom = useChatRoom(activeRoomId);
  const messages = messagesQuery.data?.data ?? [];
  const { mutate: markRead } = useMarkMessagesRead();

  // Join the active room when it changes; leave on cleanup
  useEffect(() => {
    if (!activeRoomId) return;
    joinRoom(activeRoomId);
    return () => {
      leaveRoom(activeRoomId);
    };
  }, [activeRoomId, joinRoom, leaveRoom]);

  // Mark unread messages from others as read whenever the message list updates
  useEffect(() => {
    if (!activeRoomId || !user) return;
    const unread = messages
      .filter((m) => !m.is_read && m.sender.id !== user.id && !m.deleted_at)
      .map((m) => m.chat_id);
    if (unread.length === 0) return;
    markRead({ roomId: activeRoomId, chatIds: unread });
  }, [messages, activeRoomId, user, markRead]);

  // Determine the "other" user in private rooms from message history
  const otherUsername = useMemo(() => {
    if (!activeRoom || activeRoom.type !== 'private' || !user) return undefined;
    const otherMsg = messages.find((m) => m.sender.id !== user.id);
    return otherMsg?.sender.username;
  }, [activeRoom, messages, user]);

  // Resolve typing user IDs → display names (uses message history as name lookup)
  const typingUserNames = useMemo(() => {
    if (!activeRoomId) return [];
    const typingIds = typingUsersByRoom[activeRoomId] ?? [];
    if (typingIds.length === 0) return [];
    const nameMap = new Map<number, string>();
    messages.forEach((m) => nameMap.set(m.sender.id, m.sender.username));
    return typingIds.map((uid) => nameMap.get(uid) ?? `User ${uid}`);
  }, [typingUsersByRoom, activeRoomId, messages]);

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
              typingUsers={typingUserNames}
            />
            <MessageList
              messages={messages}
              currentUserId={user?.id ?? 0}
              isLoading={messagesQuery.isLoading}
              roomId={activeRoomId!}
              typingUsers={typingUserNames}
            />
            <ChatInput
              roomId={activeRoomId!}
              onStartTyping={startTyping}
              onStopTyping={stopTyping}
            />
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
