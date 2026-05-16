'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { MessageCircle } from 'lucide-react';
import { ConversationList } from '@/components/chat/ConversationList';
import { ChatHeader } from '@/components/chat/ChatHeader';
import { MessageList } from '@/components/chat/MessageList';
import { ChatInput } from '@/components/chat/ChatInput';
import { NewChatModal } from '@/components/chat/NewChatModal';
import { useChatSocket } from '@/hooks/useChatSocket';
import { useChatMessages } from '@/hooks/chat/useChatMessages';
import { useChatRoom } from '@/hooks/chat/useChatRooms';
import { useMarkMessagesRead } from '@/hooks/chat/useMarkMessagesRead';
import { useProfile } from '@/hooks/useProfile';
import { useChatStore } from '@/stores/chatStore';
import { useAuthStore } from '@/stores/authStore';
import { useDeleteMessage } from '@/hooks/chat/useDeleteMessage';
import { useAddReaction, useRemoveReaction, useSwapReaction } from '@/hooks/chat/useReactToMessage';
import type { ChatMessage } from '@/lib/services/chat';

export default function MessagesPage() {
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const { activeRoomId, setActiveRoomId } = useChatStore();
  const [showNewChat, setShowNewChat] = useState(false);
  const [editingMessage, setEditingMessage] = useState<ChatMessage | null>(null);
  const [replyingTo, setReplyingTo] = useState<ChatMessage | null>(null);

  const deleteMessage = useDeleteMessage(activeRoomId ?? 0);
  const addReaction = useAddReaction(activeRoomId ?? 0);
  const removeReaction = useRemoveReaction(activeRoomId ?? 0);
  const swapReaction = useSwapReaction(activeRoomId ?? 0);

  const handleEditMessage = useCallback((message: ChatMessage) => {
    setEditingMessage(message);
  }, []);

  const handleDeleteMessage = useCallback((chatId: number) => {
    deleteMessage.mutate(chatId);
  }, [deleteMessage]);

  const handleCancelEdit = useCallback(() => {
    setEditingMessage(null);
  }, []);

  const handleReplyMessage = useCallback((message: ChatMessage) => {
    setReplyingTo(message);
    setEditingMessage(null);
  }, []);

  const handleCancelReply = useCallback(() => {
    setReplyingTo(null);
  }, []);

  const handleReact = useCallback((chatId: number, emoji: string) => {
    addReaction.mutate({ chatId, emoji });
  }, [addReaction]);

  const handleUnreact = useCallback((chatId: number, emoji: string) => {
    removeReaction.mutate({ chatId, emoji });
  }, [removeReaction]);

  const handleSwapReact = useCallback((chatId: number, oldEmoji: string, newEmoji: string) => {
    swapReaction.mutate({ chatId, oldEmoji, newEmoji });
  }, [swapReaction]);

  // Single WebSocket connection for the entire messages page
  const { joinRoom, leaveRoom, startTyping, stopTyping, typingUsersByRoom } = useChatSocket({
    accessToken: token,
    currentUserId: user?.id,
  });

  const messagesQuery = useChatMessages(activeRoomId);
  const activeRoom = useChatRoom(activeRoomId);
  const messages = messagesQuery.data?.data ?? [];
  const { mutate: markRead } = useMarkMessagesRead();

  // For private rooms, fetch the other person's full profile (avatar, name, faculty)
  const otherUserId = activeRoom?.type === 'private' ? activeRoom.other_user?.user_id : undefined;
  const { data: otherProfile } = useProfile(otherUserId ?? null);

  // Join the active room when it changes; leave on cleanup
  useEffect(() => {
    if (!activeRoomId) return;
    setEditingMessage(null);
    setReplyingTo(null);
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

  // Whether anyone else is currently typing in this room
  const isTyping = useMemo(() => {
    if (!activeRoomId || !user) return false;
    const typingIds = typingUsersByRoom[activeRoomId] ?? [];
    return typingIds.some((id) => id !== user.id);
  }, [typingUsersByRoom, activeRoomId, user]);

  // Typing user display names — used by MessageList for the bubble label (if any)
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
      <div style={{ width: 300, flexShrink: 0 }}>
        <ConversationList
          activeRoomId={activeRoomId}
          onRoomSelect={setActiveRoomId}
          onNewChat={() => setShowNewChat(true)}
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
              otherProfile={otherProfile}
              isTyping={isTyping}
            />
            <MessageList
              messages={messages}
              currentUserId={user?.id ?? 0}
              isLoading={messagesQuery.isLoading}
              roomId={activeRoomId!}
              typingUsers={typingUserNames}
              isPrivateRoom={activeRoom.type === 'private'}
              onEditMessage={handleEditMessage}
              onDeleteMessage={handleDeleteMessage}
              onReplyMessage={handleReplyMessage}
              onReact={handleReact}
              onUnreact={handleUnreact}
              onSwapReact={handleSwapReact}
            />
            <ChatInput
              roomId={activeRoomId!}
              editingMessage={editingMessage}
              onCancelEdit={handleCancelEdit}
              replyingTo={replyingTo}
              onCancelReply={handleCancelReply}
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
              gap: 16,
            }}
          >
            <div
              style={{
                width: 72,
                height: 72,
                borderRadius: '50%',
                background: 'var(--bg)',
                border: '1px solid var(--border)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <MessageCircle size={36} color="var(--text)" strokeWidth={1.4} />
            </div>
            <div style={{ textAlign: 'center', display: 'flex', flexDirection: 'column', gap: 6 }}>
              <div style={{ fontSize: 20, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.02em' }}>
                Start a Conversation
              </div>
              <div style={{ fontSize: 14, color: 'var(--text-muted)', maxWidth: 280 }}>
                Choose from your existing conversations, or start a new one.
              </div>
            </div>
            <button
              onClick={() => setShowNewChat(true)}
              style={{
                padding: '10px 24px',
                borderRadius: 9999,
                background: 'var(--text)',
                color: 'var(--bg)',
                border: 'none',
                cursor: 'pointer',
                fontWeight: 700,
                fontSize: 14,
                fontFamily: 'inherit',
              }}
            >
              New chat
            </button>
          </div>
        )}
      </div>
      {showNewChat && (
        <NewChatModal
          onClose={() => setShowNewChat(false)}
          onRoomCreated={(roomId) => {
            setShowNewChat(false);
            setActiveRoomId(roomId);
          }}
        />
      )}
    </div>
  );
}
