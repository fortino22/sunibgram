'use client';

import { Suspense, useCallback, useEffect, useMemo, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { MessageCircle } from 'lucide-react';
import { ConversationList } from '@/components/chat/ConversationList';
import { ChatHeader } from '@/components/chat/ChatHeader';
import { MessageList } from '@/components/chat/MessageList';
import { ChatInput } from '@/components/chat/ChatInput';
import { NewChatModal } from '@/components/chat/NewChatModal';
import { GroupSettingsModal } from '@/components/chat/GroupSettingsModal';
import { useChatSocket } from '@/hooks/useChatSocket';
import { useChatMessages } from '@/hooks/chat/useChatMessages';
import { useChatRoom, useChatRooms } from '@/hooks/chat/useChatRooms';
import { useMarkMessagesRead } from '@/hooks/chat/useMarkMessagesRead';
import { useProfile } from '@/hooks/useProfile';
import { useChatStore } from '@/stores/chatStore';
import { useAuthStore } from '@/stores/authStore';
import { useDeleteMessage } from '@/hooks/chat/useDeleteMessage';
import { useAddReaction, useRemoveReaction, useSwapReaction } from '@/hooks/chat/useReactToMessage';
import { useSearchParams } from 'next/navigation';
import type { ChatMessage, ChatRoomsResponse } from '@/lib/services/chat';

function RoomDeepLink({ setActiveRoomId }: { setActiveRoomId: (id: number) => void }) {
  const searchParams = useSearchParams();
  useEffect(() => {
    const roomParam = searchParams.get('room');
    if (roomParam) setActiveRoomId(Number(roomParam));
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  return null;
}

export default function MessagesPage() {
  const token = useAuthStore((s) => s.token);
  const user = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();
  const { activeRoomId, setActiveRoomId } = useChatStore();
  const [showNewChat, setShowNewChat] = useState(false);
  const [showGroupSettings, setShowGroupSettings] = useState(false);
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
  const { joinRoom, startTyping, stopTyping, typingUsersByRoom, status: socketStatus } = useChatSocket({
    accessToken: token,
    currentUserId: user?.id,
    activeRoomId,
  });

  const allRoomsQuery = useChatRooms();
  const allRooms = allRoomsQuery.data?.data ?? [];

  // Auto-join ALL rooms when the socket opens so we receive broadcasts for every room
  // (enables real-time updates in conversation list, avatar changes, etc.)
  useEffect(() => {
    if (socketStatus !== 'open' || allRooms.length === 0) return;
    allRooms.forEach((r) => joinRoom(r.room_id));
  }, [socketStatus, allRooms, joinRoom]);

  const messagesQuery = useChatMessages(activeRoomId);
  const activeRoom = useChatRoom(activeRoomId);
  const messages = messagesQuery.data?.data ?? [];
  const { mutate: markRead } = useMarkMessagesRead();

  // For private rooms, fetch the other person's full profile (avatar, name, faculty)
  const otherUserId = activeRoom?.type === 'private' ? activeRoom.other_user?.user_id : undefined;
  const { data: otherProfile } = useProfile(otherUserId ?? null);

  // Join the active room when it changes and zero out its unread count
  useEffect(() => {
    if (!activeRoomId) return;
    setEditingMessage(null);
    setReplyingTo(null);
    joinRoom(activeRoomId);
    queryClient.setQueryData<ChatRoomsResponse>(['chat', 'rooms'], (old) => {
      if (!old) return old;
      return {
        ...old,
        data: old.data.map((r) =>
          r.room_id === activeRoomId ? { ...r, unread_count: 0 } : r,
        ),
      };
    });
  }, [activeRoomId, joinRoom, queryClient]);

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
      <Suspense fallback={null}>
        <RoomDeepLink setActiveRoomId={setActiveRoomId} />
      </Suspense>

      {/* Left — Conversation list */}
      <div style={{ width: 300, flexShrink: 0 }}>
        <ConversationList
          activeRoomId={activeRoomId}
          onRoomSelect={setActiveRoomId}
          onNewChat={() => setShowNewChat(true)}
          typingUsersByRoom={typingUsersByRoom}
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
              onGroupSettings={
                activeRoom.type === 'group'
                  ? () => setShowGroupSettings(true)
                  : undefined
              }
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
              <MessageCircle size={36} strokeWidth={1.4} style={{ color: 'var(--text)' }} />
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
              className="btn-solid"
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
      {showGroupSettings && activeRoom?.type === 'group' && (
        <GroupSettingsModal
          room={activeRoom}
          onClose={() => setShowGroupSettings(false)}
          onLeave={() => {
            setShowGroupSettings(false);
            setActiveRoomId(null);
          }}
        />
      )}
    </div>
  );
}
