# WebSocket API

The backend WebSocket endpoint is:

```text
ws://localhost:8080/ws
```

The frontend uses the native browser `WebSocket` API. Do **not** use `socket.io-client`; this endpoint speaks plain WebSocket JSON messages, not the Socket.IO protocol.

## 1. Connection Setup

Authentication is required when opening the socket. The backend accepts the JWT access token in any of these locations:

1. Query parameter named `token`
2. Query parameter named `access_token`
3. `Authorization: Bearer <access_token>` header on the upgrade request

In browser code, custom headers cannot be set on the native `WebSocket` constructor. For Next.js client components, pass the token in the URL query string:

```ts
const socket = new WebSocket(
  `ws://localhost:8080/ws?token=${encodeURIComponent(accessToken)}`,
);
```

Use `access_token` instead if that is the convention in the calling code:

```ts
const socket = new WebSocket(
  `ws://localhost:8080/ws?access_token=${encodeURIComponent(accessToken)}`,
);
```

The `Authorization: Bearer` upgrade header is useful for non-browser clients or backend-to-backend clients that can control WebSocket upgrade headers.

### Reconnection Strategy

Recommended client behavior:

- Connect only when an access token is available.
- Reconnect after unexpected disconnects.
- Use exponential backoff with a maximum delay.
- Add a small jitter to avoid multiple browser tabs reconnecting at the same time.
- Reset the retry counter after a successful `open`.
- Do not reconnect after an intentional close, such as component unmount or user logout.
- Refresh or replace the access token before reconnecting if the current token may have expired.

Example backoff calculation:

```ts
const baseDelayMs = 500;
const maxDelayMs = 10_000;
const attempt = 3;
const jitterMs = Math.floor(Math.random() * 250);

const reconnectDelayMs = Math.min(maxDelayMs, baseDelayMs * 2 ** attempt) + jitterMs;
```

## 2. Message Format

All messages are JSON envelopes.

Client to server:

```json
{ "event": "<event_name>", "...otherFields": "..." }
```

Server to client:

```json
{ "event": "<event_name>", "data": "<payload>" }
```

Error:

```json
{ "event": "socket:error", "data": { "message": "..." } }
```

### Shared TypeScript Types

```ts
export type ClientEventName =
  | 'join_room'
  | 'leave_room'
  | 'typing:start'
  | 'typing:stop'
  | 'notification:get'
  | 'notification:delete';

export type ServerEventName =
  | 'room:joined'
  | 'room:left'
  | 'typing:start'
  | 'typing:stop'
  | 'chat:new'
  | 'notification:new'
  | 'notification:list'
  | 'notification:deleted'
  | 'socket:error';

export interface ClientEnvelope<EventName extends string = string> {
  event: EventName;
  [key: string]: unknown;
}

export interface ServerEnvelope<EventName extends string = string, Data = unknown> {
  event: EventName;
  data: Data;
}

export interface SocketErrorPayload {
  message: string;
}

export type SocketErrorMessage = ServerEnvelope<'socket:error', SocketErrorPayload>;
```

## 3. Chat Events

### Chat TypeScript Types

```ts
export interface JoinRoomMessage {
  event: 'join_room';
  room_id: number;
}

export interface LeaveRoomMessage {
  event: 'leave_room';
  room_id: number;
}

export interface TypingStartMessage {
  event: 'typing:start';
  room_id: number;
}

export interface TypingStopMessage {
  event: 'typing:stop';
  room_id: number;
}

export type ChatClientMessage =
  | JoinRoomMessage
  | LeaveRoomMessage
  | TypingStartMessage
  | TypingStopMessage;

export interface RoomJoinedPayload {
  room_id: number;
}

export interface RoomLeftPayload {
  room_id: number;
}

export interface TypingPayload {
  room_id: number;
  user_id: number;
}

export interface ChatNewPayload {
  id: number;
  room_id: number;
  sender_id: number;
  content: string;
  created_at: string;
}

export type RoomJoinedMessage = ServerEnvelope<'room:joined', RoomJoinedPayload>;
export type RoomLeftMessage = ServerEnvelope<'room:left', RoomLeftPayload>;
export type TypingStartBroadcast = ServerEnvelope<'typing:start', TypingPayload>;
export type TypingStopBroadcast = ServerEnvelope<'typing:stop', TypingPayload>;
export type ChatNewMessage = ServerEnvelope<'chat:new', ChatNewPayload>;

export type ChatServerMessage =
  | RoomJoinedMessage
  | RoomLeftMessage
  | TypingStartBroadcast
  | TypingStopBroadcast
  | ChatNewMessage
  | SocketErrorMessage;
```

### Join a Room

Client sends:

```json
{ "event": "join_room", "room_id": 42 }
```

Server responds:

```json
{ "event": "room:joined", "data": { "room_id": 42 } }
```

### Leave a Room

Client sends:

```json
{ "event": "leave_room", "room_id": 42 }
```

Server responds:

```json
{ "event": "room:left", "data": { "room_id": 42 } }
```

### Typing Events

Client sends:

```json
{ "event": "typing:start", "room_id": 42 }
```

or:

```json
{ "event": "typing:stop", "room_id": 42 }
```

Server broadcasts to all room members:

```json
{ "event": "typing:start", "data": { "room_id": 42, "user_id": 7 } }
```

or:

```json
{ "event": "typing:stop", "data": { "room_id": 42, "user_id": 7 } }
```

### New Chat Messages

`chat:new` is a server push event triggered by `POST /api/chats`.

Server sends:

```json
{
  "event": "chat:new",
  "data": {
    "id": 1001,
    "room_id": 42,
    "sender_id": 7,
    "content": "Hello",
    "created_at": "2026-05-11T10:00:00Z"
  }
}
```

### Chat Send Example

```ts
sendJson({ event: 'join_room', room_id: 42 });
sendJson({ event: 'typing:start', room_id: 42 });
sendJson({ event: 'typing:stop', room_id: 42 });
sendJson({ event: 'leave_room', room_id: 42 });
```

### Chat Receive Example

```ts
function handleChatMessage(message: ChatServerMessage): void {
  switch (message.event) {
    case 'room:joined':
      console.log('Joined room', message.data.room_id);
      break;
    case 'room:left':
      console.log('Left room', message.data.room_id);
      break;
    case 'typing:start':
      console.log('User started typing', message.data.user_id);
      break;
    case 'typing:stop':
      console.log('User stopped typing', message.data.user_id);
      break;
    case 'chat:new':
      console.log('New chat message', message.data.content);
      break;
    case 'socket:error':
      console.error(message.data.message);
      break;
  }
}
```

## 4. Notification Events

### Notification TypeScript Types

The notification payload intentionally mixes PascalCase `Notification` fields with snake_case `Actor` fields. Keep the frontend types aligned with the backend payload instead of renaming fields in the socket layer.

```ts
export interface NotificationPayloadNotification {
  ID: number;
  Type: string;
  Message: string;
  SenderUserID: number;
  RecipientUserID: number;
  ReferenceID: number;
  CreatedAt: string;
}

export interface NotificationActorFaculty {
  faculty_id: number;
  faculty_name: string;
  total_users: number;
}

export interface NotificationPayloadActor {
  user_id: number;
  username: string;
  name: string;
  avatar_url: string | null;
  is_premium: boolean;
  faculty: NotificationActorFaculty | null;
}

export interface NotificationPayload {
  Notification: NotificationPayloadNotification;
  Actor: NotificationPayloadActor;
}

export interface NotificationGetMessage {
  event: 'notification:get';
  limit: number;
  offset: number;
}

export interface NotificationDeleteMessage {
  event: 'notification:delete';
  id: number;
}

export type NotificationClientMessage =
  | NotificationGetMessage
  | NotificationDeleteMessage;

export type NotificationNewMessage = ServerEnvelope<'notification:new', NotificationPayload>;
export type NotificationListMessage = ServerEnvelope<'notification:list', NotificationPayload[]>;
export type NotificationDeletedMessage = ServerEnvelope<'notification:deleted', { id: number }>;

export type NotificationServerMessage =
  | NotificationNewMessage
  | NotificationListMessage
  | NotificationDeletedMessage
  | SocketErrorMessage;
```

### `notification:new`

`notification:new` is an automatic server push event.

Full payload:

```json
{
  "event": "notification:new",
  "data": {
    "Notification": {
      "ID": 55,
      "Type": "like",
      "Message": "Someone liked your post",
      "SenderUserID": 7,
      "RecipientUserID": 9,
      "ReferenceID": 123,
      "CreatedAt": "2026-05-11T10:00:00Z"
    },
    "Actor": {
      "user_id": 7,
      "username": "sender",
      "name": "Sender User",
      "avatar_url": "https://example.com/avatar.png",
      "is_premium": true,
      "faculty": {
        "faculty_id": 1,
        "faculty_name": "Engineering",
        "total_users": 240
      }
    }
  }
}
```

### Get Notifications

Client sends:

```json
{ "event": "notification:get", "limit": 20, "offset": 0 }
```

Server responds:

```json
{
  "event": "notification:list",
  "data": [
    {
      "Notification": {
        "ID": 55,
        "Type": "like",
        "Message": "Someone liked your post",
        "SenderUserID": 7,
        "RecipientUserID": 9,
        "ReferenceID": 123,
        "CreatedAt": "2026-05-11T10:00:00Z"
      },
      "Actor": {
        "user_id": 7,
        "username": "sender",
        "name": "Sender User",
        "avatar_url": null,
        "is_premium": false,
        "faculty": null
      }
    }
  ]
}
```

### Delete Notification

Client sends:

```json
{ "event": "notification:delete", "id": 55 }
```

Server responds:

```json
{ "event": "notification:deleted", "data": { "id": 55 } }
```

## 5. Error Handling

The backend can send socket errors as regular JSON messages:

```json
{ "event": "socket:error", "data": { "message": "Invalid payload" } }
```

Handle `socket:error` in every feature hook. It may indicate validation errors, authorization failures, missing rooms, or other backend-side failures.

Also handle WebSocket lifecycle failures:

- `onerror`: the browser reports that the connection failed or encountered a transport error. Browser WebSocket errors usually do not expose detailed backend messages.
- `onclose`: the socket disconnected. Inspect the close code and reason when available.
- Unexpected close: schedule reconnect using backoff.
- Intentional close: do not reconnect.

Common close handling:

```ts
socket.onclose = (event) => {
  console.log('Socket closed', event.code, event.reason);
};
```

## 6. Full Example Hooks

The following hooks are written for Next.js App Router client components and TypeScript strict mode. Place them in a client-side file, for example `frontend/src/hooks/useWebSocket.ts`, or split them by feature.

### `useWebSocket` Base Hook

```tsx
'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

export type WebSocketStatus =
  | 'idle'
  | 'connecting'
  | 'open'
  | 'closed'
  | 'reconnecting'
  | 'error';

export interface ClientEnvelope<EventName extends string = string> {
  event: EventName;
  [key: string]: unknown;
}

export interface ServerEnvelope<EventName extends string = string, Data = unknown> {
  event: EventName;
  data: Data;
}

export interface SocketErrorPayload {
  message: string;
}

export type SocketErrorMessage = ServerEnvelope<'socket:error', SocketErrorPayload>;

export interface UseWebSocketOptions {
  accessToken: string | null;
  url?: string;
  tokenParamName?: 'token' | 'access_token';
  reconnect?: boolean;
  maxReconnectAttempts?: number;
  baseReconnectDelayMs?: number;
  maxReconnectDelayMs?: number;
  onMessage?: (message: ServerEnvelope) => void;
  onErrorMessage?: (error: SocketErrorPayload) => void;
}

export interface UseWebSocketResult {
  status: WebSocketStatus;
  lastError: string | null;
  closeCode: number | null;
  sendJson: <Message extends ClientEnvelope>(message: Message) => boolean;
  connect: () => void;
  disconnect: () => void;
}

const DEFAULT_WS_URL = 'ws://localhost:8080/ws';

function buildSocketUrl(
  baseUrl: string,
  accessToken: string,
  tokenParamName: 'token' | 'access_token',
): string {
  const url = new URL(baseUrl);
  url.searchParams.set(tokenParamName, accessToken);
  return url.toString();
}

function getReconnectDelayMs(
  attempt: number,
  baseReconnectDelayMs: number,
  maxReconnectDelayMs: number,
): number {
  const jitterMs = Math.floor(Math.random() * 250);
  return Math.min(maxReconnectDelayMs, baseReconnectDelayMs * 2 ** attempt) + jitterMs;
}

function parseServerMessage(raw: string): ServerEnvelope | null {
  try {
    const parsed: unknown = JSON.parse(raw);

    if (
      typeof parsed !== 'object' ||
      parsed === null ||
      !('event' in parsed) ||
      typeof (parsed as { event: unknown }).event !== 'string'
    ) {
      return null;
    }

    return parsed as ServerEnvelope;
  } catch {
    return null;
  }
}

export function useWebSocket({
  accessToken,
  url = DEFAULT_WS_URL,
  tokenParamName = 'token',
  reconnect = true,
  maxReconnectAttempts = 8,
  baseReconnectDelayMs = 500,
  maxReconnectDelayMs = 10_000,
  onMessage,
  onErrorMessage,
}: UseWebSocketOptions): UseWebSocketResult {
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptRef = useRef(0);
  const intentionallyClosedRef = useRef(false);
  const onMessageRef = useRef(onMessage);
  const onErrorMessageRef = useRef(onErrorMessage);

  const [status, setStatus] = useState<WebSocketStatus>('idle');
  const [lastError, setLastError] = useState<string | null>(null);
  const [closeCode, setCloseCode] = useState<number | null>(null);

  useEffect(() => {
    onMessageRef.current = onMessage;
  }, [onMessage]);

  useEffect(() => {
    onErrorMessageRef.current = onErrorMessage;
  }, [onErrorMessage]);

  const clearReconnectTimer = useCallback((): void => {
    if (reconnectTimerRef.current !== null) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
  }, []);

  const disconnect = useCallback((): void => {
    intentionallyClosedRef.current = true;
    clearReconnectTimer();

    if (socketRef.current !== null) {
      socketRef.current.close(1000, 'Client disconnect');
      socketRef.current = null;
    }

    setStatus('closed');
  }, [clearReconnectTimer]);

  const connect = useCallback((): void => {
    if (!accessToken) {
      setStatus('idle');
      return;
    }

    const currentSocket = socketRef.current;
    if (
      currentSocket?.readyState === WebSocket.OPEN ||
      currentSocket?.readyState === WebSocket.CONNECTING
    ) {
      return;
    }

    intentionallyClosedRef.current = false;
    clearReconnectTimer();
    setStatus(reconnectAttemptRef.current > 0 ? 'reconnecting' : 'connecting');
    setLastError(null);

    const socket = new WebSocket(buildSocketUrl(url, accessToken, tokenParamName));
    socketRef.current = socket;

    socket.onopen = (): void => {
      reconnectAttemptRef.current = 0;
      setStatus('open');
      setCloseCode(null);
    };

    socket.onmessage = (event: MessageEvent<string>): void => {
      const message = parseServerMessage(event.data);

      if (message === null) {
        setLastError('Received invalid WebSocket message');
        return;
      }

      if (message.event === 'socket:error') {
        const errorMessage = message as SocketErrorMessage;
        setLastError(errorMessage.data.message);
        onErrorMessageRef.current?.(errorMessage.data);
        return;
      }

      onMessageRef.current?.(message);
    };

    socket.onerror = (): void => {
      setStatus('error');
      setLastError('WebSocket connection error');
    };

    socket.onclose = (event: CloseEvent): void => {
      if (socketRef.current === socket) {
        socketRef.current = null;
      }

      setCloseCode(event.code);

      if (intentionallyClosedRef.current || !reconnect || !accessToken) {
        setStatus('closed');
        return;
      }

      if (reconnectAttemptRef.current >= maxReconnectAttempts) {
        setStatus('closed');
        setLastError('WebSocket reconnect attempts exhausted');
        return;
      }

      const delayMs = getReconnectDelayMs(
        reconnectAttemptRef.current,
        baseReconnectDelayMs,
        maxReconnectDelayMs,
      );

      reconnectAttemptRef.current += 1;
      setStatus('reconnecting');

      reconnectTimerRef.current = setTimeout(() => {
        connect();
      }, delayMs);
    };
  }, [
    accessToken,
    baseReconnectDelayMs,
    clearReconnectTimer,
    maxReconnectAttempts,
    maxReconnectDelayMs,
    reconnect,
    tokenParamName,
    url,
  ]);

  const sendJson = useCallback(<Message extends ClientEnvelope>(message: Message): boolean => {
    const socket = socketRef.current;

    if (socket?.readyState !== WebSocket.OPEN) {
      setLastError('WebSocket is not open');
      return false;
    }

    socket.send(JSON.stringify(message));
    return true;
  }, []);

  useEffect(() => {
    connect();

    return () => {
      intentionallyClosedRef.current = true;
      clearReconnectTimer();

      if (socketRef.current !== null) {
        socketRef.current.close(1000, 'Component unmounted');
        socketRef.current = null;
      }
    };
  }, [clearReconnectTimer, connect]);

  return useMemo(
    () => ({
      status,
      lastError,
      closeCode,
      sendJson,
      connect,
      disconnect,
    }),
    [closeCode, connect, disconnect, lastError, sendJson, status],
  );
}
```

### `useChatSocket`

```tsx
'use client';

import { useCallback, useMemo, useState } from 'react';

import {
  ServerEnvelope,
  SocketErrorPayload,
  useWebSocket,
  WebSocketStatus,
} from './useWebSocket';

export interface JoinRoomMessage {
  event: 'join_room';
  room_id: number;
}

export interface LeaveRoomMessage {
  event: 'leave_room';
  room_id: number;
}

export interface TypingStartMessage {
  event: 'typing:start';
  room_id: number;
}

export interface TypingStopMessage {
  event: 'typing:stop';
  room_id: number;
}

export type ChatClientMessage =
  | JoinRoomMessage
  | LeaveRoomMessage
  | TypingStartMessage
  | TypingStopMessage;

export interface RoomJoinedPayload {
  room_id: number;
}

export interface RoomLeftPayload {
  room_id: number;
}

export interface TypingPayload {
  room_id: number;
  user_id: number;
}

export interface ChatNewPayload {
  id: number;
  room_id: number;
  sender_id: number;
  content: string;
  created_at: string;
}

export interface UseChatSocketOptions {
  accessToken: string | null;
  onRoomJoined?: (payload: RoomJoinedPayload) => void;
  onRoomLeft?: (payload: RoomLeftPayload) => void;
  onTypingStart?: (payload: TypingPayload) => void;
  onTypingStop?: (payload: TypingPayload) => void;
  onChatNew?: (payload: ChatNewPayload) => void;
  onSocketError?: (payload: SocketErrorPayload) => void;
}

export interface UseChatSocketResult {
  status: WebSocketStatus;
  lastError: string | null;
  closeCode: number | null;
  messages: ChatNewPayload[];
  typingUsersByRoom: Record<number, number[]>;
  joinRoom: (roomId: number) => boolean;
  leaveRoom: (roomId: number) => boolean;
  startTyping: (roomId: number) => boolean;
  stopTyping: (roomId: number) => boolean;
  connect: () => void;
  disconnect: () => void;
}

function addTypingUser(
  current: Record<number, number[]>,
  payload: TypingPayload,
): Record<number, number[]> {
  const existing = current[payload.room_id] ?? [];

  if (existing.includes(payload.user_id)) {
    return current;
  }

  return {
    ...current,
    [payload.room_id]: [...existing, payload.user_id],
  };
}

function removeTypingUser(
  current: Record<number, number[]>,
  payload: TypingPayload,
): Record<number, number[]> {
  const existing = current[payload.room_id] ?? [];
  const nextUsers = existing.filter((userId) => userId !== payload.user_id);

  return {
    ...current,
    [payload.room_id]: nextUsers,
  };
}

export function useChatSocket({
  accessToken,
  onRoomJoined,
  onRoomLeft,
  onTypingStart,
  onTypingStop,
  onChatNew,
  onSocketError,
}: UseChatSocketOptions): UseChatSocketResult {
  const [messages, setMessages] = useState<ChatNewPayload[]>([]);
  const [typingUsersByRoom, setTypingUsersByRoom] = useState<Record<number, number[]>>({});

  const handleMessage = useCallback(
    (message: ServerEnvelope): void => {
      switch (message.event) {
        case 'room:joined': {
          const payload = message.data as RoomJoinedPayload;
          onRoomJoined?.(payload);
          break;
        }
        case 'room:left': {
          const payload = message.data as RoomLeftPayload;
          onRoomLeft?.(payload);
          break;
        }
        case 'typing:start': {
          const payload = message.data as TypingPayload;
          setTypingUsersByRoom((current) => addTypingUser(current, payload));
          onTypingStart?.(payload);
          break;
        }
        case 'typing:stop': {
          const payload = message.data as TypingPayload;
          setTypingUsersByRoom((current) => removeTypingUser(current, payload));
          onTypingStop?.(payload);
          break;
        }
        case 'chat:new': {
          const payload = message.data as ChatNewPayload;
          setMessages((current) => [...current, payload]);
          onChatNew?.(payload);
          break;
        }
      }
    },
    [onChatNew, onRoomJoined, onRoomLeft, onTypingStart, onTypingStop],
  );

  const socket = useWebSocket({
    accessToken,
    onMessage: handleMessage,
    onErrorMessage: onSocketError,
  });

  const joinRoom = useCallback(
    (roomId: number): boolean => socket.sendJson<JoinRoomMessage>({ event: 'join_room', room_id: roomId }),
    [socket],
  );

  const leaveRoom = useCallback(
    (roomId: number): boolean => socket.sendJson<LeaveRoomMessage>({ event: 'leave_room', room_id: roomId }),
    [socket],
  );

  const startTyping = useCallback(
    (roomId: number): boolean =>
      socket.sendJson<TypingStartMessage>({ event: 'typing:start', room_id: roomId }),
    [socket],
  );

  const stopTyping = useCallback(
    (roomId: number): boolean =>
      socket.sendJson<TypingStopMessage>({ event: 'typing:stop', room_id: roomId }),
    [socket],
  );

  return useMemo(
    () => ({
      status: socket.status,
      lastError: socket.lastError,
      closeCode: socket.closeCode,
      messages,
      typingUsersByRoom,
      joinRoom,
      leaveRoom,
      startTyping,
      stopTyping,
      connect: socket.connect,
      disconnect: socket.disconnect,
    }),
    [
      joinRoom,
      leaveRoom,
      messages,
      socket.closeCode,
      socket.connect,
      socket.disconnect,
      socket.lastError,
      socket.status,
      startTyping,
      stopTyping,
      typingUsersByRoom,
    ],
  );
}
```

### `useNotificationSocket`

```tsx
'use client';

import { useCallback, useMemo, useState } from 'react';

import {
  ServerEnvelope,
  SocketErrorPayload,
  useWebSocket,
  WebSocketStatus,
} from './useWebSocket';

export interface NotificationPayloadNotification {
  ID: number;
  Type: string;
  Message: string;
  SenderUserID: number;
  RecipientUserID: number;
  ReferenceID: number;
  CreatedAt: string;
}

export interface NotificationActorFaculty {
  faculty_id: number;
  faculty_name: string;
  total_users: number;
}

export interface NotificationPayloadActor {
  user_id: number;
  username: string;
  name: string;
  avatar_url: string | null;
  is_premium: boolean;
  faculty: NotificationActorFaculty | null;
}

export interface NotificationPayload {
  Notification: NotificationPayloadNotification;
  Actor: NotificationPayloadActor;
}

export interface NotificationGetMessage {
  event: 'notification:get';
  limit: number;
  offset: number;
}

export interface NotificationDeleteMessage {
  event: 'notification:delete';
  id: number;
}

export interface UseNotificationSocketOptions {
  accessToken: string | null;
  onNotificationNew?: (payload: NotificationPayload) => void;
  onNotificationList?: (payload: NotificationPayload[]) => void;
  onNotificationDeleted?: (payload: { id: number }) => void;
  onSocketError?: (payload: SocketErrorPayload) => void;
}

export interface UseNotificationSocketResult {
  status: WebSocketStatus;
  lastError: string | null;
  closeCode: number | null;
  notifications: NotificationPayload[];
  getNotifications: (limit?: number, offset?: number) => boolean;
  deleteNotification: (id: number) => boolean;
  connect: () => void;
  disconnect: () => void;
}

export function useNotificationSocket({
  accessToken,
  onNotificationNew,
  onNotificationList,
  onNotificationDeleted,
  onSocketError,
}: UseNotificationSocketOptions): UseNotificationSocketResult {
  const [notifications, setNotifications] = useState<NotificationPayload[]>([]);

  const handleMessage = useCallback(
    (message: ServerEnvelope): void => {
      switch (message.event) {
        case 'notification:new': {
          const payload = message.data as NotificationPayload;
          setNotifications((current) => [payload, ...current]);
          onNotificationNew?.(payload);
          break;
        }
        case 'notification:list': {
          const payload = message.data as NotificationPayload[];
          setNotifications(payload);
          onNotificationList?.(payload);
          break;
        }
        case 'notification:deleted': {
          const payload = message.data as { id: number };
          setNotifications((current) =>
            current.filter((item) => item.Notification.ID !== payload.id),
          );
          onNotificationDeleted?.(payload);
          break;
        }
      }
    },
    [onNotificationDeleted, onNotificationList, onNotificationNew],
  );

  const socket = useWebSocket({
    accessToken,
    onMessage: handleMessage,
    onErrorMessage: onSocketError,
  });

  const getNotifications = useCallback(
    (limit = 20, offset = 0): boolean =>
      socket.sendJson<NotificationGetMessage>({
        event: 'notification:get',
        limit,
        offset,
      }),
    [socket],
  );

  const deleteNotification = useCallback(
    (id: number): boolean =>
      socket.sendJson<NotificationDeleteMessage>({
        event: 'notification:delete',
        id,
      }),
    [socket],
  );

  return useMemo(
    () => ({
      status: socket.status,
      lastError: socket.lastError,
      closeCode: socket.closeCode,
      notifications,
      getNotifications,
      deleteNotification,
      connect: socket.connect,
      disconnect: socket.disconnect,
    }),
    [
      deleteNotification,
      getNotifications,
      notifications,
      socket.closeCode,
      socket.connect,
      socket.disconnect,
      socket.lastError,
      socket.status,
    ],
  );
}
```

## 7. Notes

- Do **not** use `socket.io-client`. The backend endpoint is a native WebSocket endpoint at `ws://localhost:8080/ws`.
- Native browser WebSocket clients cannot attach custom `Authorization` headers. Use `?token=<access_token>` or `?access_token=<access_token>` from client components.
- Keep REST message creation separate from WebSocket delivery. New chat messages are created through `POST /api/chats`; the socket receives the resulting `chat:new` push event.
- Treat socket payload field names as API contracts. Notification payloads use PascalCase under `Notification` and snake_case under `Actor`.
- Put hook files behind a `'use client';` directive because they use React state/effects and the browser `WebSocket` API.
