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

function getDefaultWsUrl(): string {
  if (typeof window === 'undefined') return 'ws://localhost:8080/ws';
  const base =
    process.env.NEXT_PUBLIC_SOCKET_URL ||
    process.env.NEXT_PUBLIC_API_URL ||
    'http://localhost:8080';
  return `${base.replace(/^http/, 'ws').replace(/\/$/, '')}/ws`;
}

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
  url,
  tokenParamName = 'token',
  reconnect = true,
  maxReconnectAttempts = 8,
  baseReconnectDelayMs = 500,
  maxReconnectDelayMs = 10_000,
  onMessage,
  onErrorMessage,
}: UseWebSocketOptions): UseWebSocketResult {
  const resolvedUrl = url ?? getDefaultWsUrl();

  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptRef = useRef(0);
  const intentionallyClosedRef = useRef(false);
  const onMessageRef = useRef(onMessage);
  const onErrorMessageRef = useRef(onErrorMessage);

  const [status, setStatus] = useState<WebSocketStatus>('idle');
  const [lastError, setLastError] = useState<string | null>(null);
  const [closeCode, setCloseCode] = useState<number | null>(null);

  useEffect(() => { onMessageRef.current = onMessage; }, [onMessage]);
  useEffect(() => { onErrorMessageRef.current = onErrorMessage; }, [onErrorMessage]);

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

    const socket = new WebSocket(buildSocketUrl(resolvedUrl, accessToken, tokenParamName));
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
    resolvedUrl,
    tokenParamName,
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
      const socket = socketRef.current;
      if (socket !== null) {
        if (socket.readyState === WebSocket.CONNECTING) {
          socket.addEventListener('open', () => {
            socket.close(1000, 'Component unmounted');
          });
        } else if (socket.readyState === WebSocket.OPEN) {
          socket.close(1000, 'Component unmounted');
        }
        socketRef.current = null;
      }
    };
  }, [clearReconnectTimer, connect]);

  return useMemo(
    () => ({ status, lastError, closeCode, sendJson, connect, disconnect }),
    [closeCode, connect, disconnect, lastError, sendJson, status],
  );
}
