# Chat Socket Documentation

Backend: `github.com/gofiber/contrib/socketio` (Fiber-native Socket.io)  
Endpoint: `ws://localhost:8080/ws`  
Protocol: Socket.io over WebSocket

---

## Authentication

JWT token required on connect. Pass via one of:

| Method | Example |
|---|---|
| Authorization header | `Authorization: Bearer <token>` |
| Query param | `?token=<token>` |
| Query param (alt) | `?access_token=<token>` |

Token must be an **access token** (not refresh). Connection rejected with HTTP 401 if invalid/missing.

---

## Connection (Frontend)

```ts
import { io, Socket } from "socket.io-client";

const socket: Socket = io(process.env.NEXT_PUBLIC_SOCKET_URL, {
  path: "/ws",
  auth: { token: accessToken },        // preferred
  // OR: transports: ["websocket"], query: { token: accessToken }
  transports: ["websocket"],
  autoConnect: false,
});

socket.connect();

socket.on("connect", () => console.log("socket connected"));
socket.on("disconnect", () => console.log("socket disconnected"));
socket.on("socket:error", (data) => console.error("socket error:", data.message));
```

---

## Message Format

All messages (send and receive) use this envelope:

```json
{
  "event": "event_name",
  "data": { ... }
}
```

---

## Client → Server Events

Events sent FROM frontend TO backend.

### `join_room`
Join a chat room to start receiving its broadcasts.  
Must call before sending/receiving messages in a room.

```ts
socket.emit("join_room", {
  event: "join_room",
  room_id: 42
});
```

Response — server emits back to you:
```json
{
  "event": "room:joined",
  "data": { "room_id": 42 }
}
```

---

### `leave_room`
Leave a room. Stop receiving its broadcasts.

```ts
socket.emit("leave_room", {
  event: "leave_room",
  room_id: 42
});
```

Response:
```json
{
  "event": "room:left",
  "data": { "room_id": 42 }
}
```

---

### `typing:start`
User started typing. Must have joined the room first.

```ts
socket.emit("typing:start", {
  event: "typing:start",
  room_id: 42
});
```

No direct response — server broadcasts to all other room members (see Server → Client below).

---

### `typing:stop`
User stopped typing.

```ts
socket.emit("typing:stop", {
  event: "typing:stop",
  room_id: 42
});
```

---

## Server → Client Events

Events received FROM backend TO frontend. Register listeners after connecting.

### `message:new`
Triggered when: `POST /chats/rooms/:room_id/messages`  
Broadcast to: all members currently in the room

```ts
socket.on("message:new", (data: ChatResponse) => {
  // append to message list
});
```

Payload shape:
```ts
{
  chat_id: number;
  room_id: number;
  sender: {
    id: number;
    username: string;
  };
  message: string | null;
  media: {
    media_id: number;
    url: string;
    media_type: string;
    file_size: number;
  }[];
  is_read: boolean;
  deleted_at: string | null;
  created_at: string;
}
```

---

### `message:edited`
Triggered when: `PATCH /messages/:id`  
Broadcast to: all members currently in the room

```ts
socket.on("message:edited", (data) => {
  // update message in list by chat_id
});
```

Payload:
```ts
{
  chat_id: number;
  room_id: number;
  sender_id: number;
  message: string;
}
```

---

### `message:deleted`
Triggered when: `DELETE /messages/:id` (unsend)  
Broadcast to: all members currently in the room

```ts
socket.on("message:deleted", (data) => {
  // remove or mark as deleted by chat_id
});
```

Payload:
```ts
{
  chat_id: number;
  room_id: number;
  sender_id: number;
  deleted_by: number;
}
```

---

### `message:read`
Triggered when: `POST /messages/read`  
Emitted to: **original sender only** (not broadcast to room)

```ts
socket.on("message:read", (data) => {
  // update read receipts for these chat_ids
});
```

Payload:
```ts
{
  room_id: number;
  chat_ids: number[];
  read_by: number;  // user ID who read the messages
}
```

---

### `room:member_added`
Triggered when: `POST /chats/rooms/:id/members`  
Broadcast to: all existing members currently in the room

```ts
socket.on("room:member_added", (data) => {
  // update member list UI
});
```

Payload:
```ts
{
  room_id: number;
  added_user_ids: number[];
  added_by: number;
  already_members: number[];
  room: {
    room_id: number;
    name: string;
    description: string;
    type: "private" | "group";
    created_at: string;
    updated_at: string;
  };
}
```

---

### `typing:start` / `typing:stop`
Triggered by: other users in the room emitting typing events  
Broadcast to: all other room members (NOT back to the emitter)

```ts
socket.on("typing:start", (data) => {
  // show "username is typing..."
});

socket.on("typing:stop", (data) => {
  // hide typing indicator
});
```

Payload:
```ts
{
  room_id: number;
  user_id: number;
}
```

---

### `socket:error`
Emitted on invalid events, bad payload, unauthorized actions, etc.

```ts
socket.on("socket:error", (data) => {
  console.error(data.message);
});
```

Payload:
```ts
{
  message: string;
}
```

---

## Frontend Implementation Pattern

### Socket service (singleton)

```ts
// lib/socket.ts
import { io, Socket } from "socket.io-client";

let socket: Socket | null = null;

export function getSocket(token: string): Socket {
  if (!socket) {
    socket = io(process.env.NEXT_PUBLIC_SOCKET_URL!, {
      path: "/ws",
      transports: ["websocket"],
      query: { token },
      autoConnect: false,
    });
  }
  return socket;
}

export function disconnectSocket() {
  socket?.disconnect();
  socket = null;
}
```

---

### Hook — connect on auth

```ts
// hooks/useSocket.ts
import { useEffect } from "react";
import { useSession } from "next-auth/react";
import { getSocket, disconnectSocket } from "@/lib/socket";

export function useSocket() {
  const { data: session } = useSession();

  useEffect(() => {
    if (!session?.accessToken) return;

    const socket = getSocket(session.accessToken);
    socket.connect();

    return () => {
      disconnectSocket();
    };
  }, [session?.accessToken]);
}
```

---

### Hook — chat room listener

```ts
// hooks/useChatRoom.ts
import { useEffect } from "react";
import { getSocket } from "@/lib/socket";
import { useSession } from "next-auth/react";

export function useChatRoom(roomId: number) {
  const { data: session } = useSession();

  useEffect(() => {
    if (!session?.accessToken || !roomId) return;

    const socket = getSocket(session.accessToken);

    // join room
    socket.emit("join_room", { event: "join_room", room_id: roomId });

    // listeners
    socket.on("message:new", (data) => {
      // add to message list (TanStack Query: queryClient.setQueryData or invalidate)
    });

    socket.on("message:edited", (data) => {
      // patch message in list
    });

    socket.on("message:deleted", (data) => {
      // remove/mark message
    });

    socket.on("message:read", (data) => {
      // update read status
    });

    socket.on("typing:start", (data) => {
      // show typing indicator
    });

    socket.on("typing:stop", (data) => {
      // hide typing indicator
    });

    return () => {
      socket.emit("leave_room", { event: "leave_room", room_id: roomId });
      socket.off("message:new");
      socket.off("message:edited");
      socket.off("message:deleted");
      socket.off("message:read");
      socket.off("typing:start");
      socket.off("typing:stop");
    };
  }, [roomId, session?.accessToken]);
}
```

---

### Typing indicator — debounced

```ts
// hooks/useTyping.ts
import { useCallback, useRef } from "react";
import { getSocket } from "@/lib/socket";
import { useSession } from "next-auth/react";

export function useTyping(roomId: number) {
  const { data: session } = useSession();
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const isTypingRef = useRef(false);

  const onType = useCallback(() => {
    if (!session?.accessToken) return;
    const socket = getSocket(session.accessToken);

    if (!isTypingRef.current) {
      isTypingRef.current = true;
      socket.emit("typing:start", { event: "typing:start", room_id: roomId });
    }

    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      isTypingRef.current = false;
      socket.emit("typing:stop", { event: "typing:stop", room_id: roomId });
    }, 2000); // stop after 2s idle
  }, [roomId, session?.accessToken]);

  return { onType };
}
```

Usage on input:
```tsx
<input onChange={onType} />
```

---

## REST APIs That Trigger Socket Events (summary)

| REST Endpoint | Socket Event Emitted | Target |
|---|---|---|
| `POST /chats/rooms/:room_id/messages` | `message:new` | all room members |
| `PATCH /messages/:id` | `message:edited` | all room members |
| `DELETE /messages/:id` | `message:deleted` | all room members |
| `POST /messages/read` | `message:read` | sender only |
| `POST /chats/rooms/:id/members` | `room:member_added` | existing room members |

REST still handles auth, validation, and DB write. Socket only broadcasts the result.  
`GET /chats/rooms/:room_id/messages` — REST only, used for initial load and pagination.
