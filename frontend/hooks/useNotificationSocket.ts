'use client';

import { useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import {
  useWebSocket,
  type ServerEnvelope,
  type WebSocketStatus,
} from './useWebSocket';
import type { NotificationItem, PaginatedNotificationResponse } from '@/lib/services/notifications';
import { useNotificationStore } from '@/stores/notificationStore';

interface NotificationNewPayload {
  Notification: {
    ID: number;
    Type: string;
    Message: string;
    SenderUserID: number;
    RecipientUserID: number;
    ReferenceID: number | null;
    CreatedAt: string;
  };
  Actor: NotificationItem['actor'];
  data: unknown;
}

interface NotificationDeletedPayload {
  id: number;
}

export interface UseNotificationSocketOptions {
  accessToken: string | null;
}

export interface UseNotificationSocketResult {
  status: WebSocketStatus;
}

export function useNotificationSocket({
  accessToken,
}: UseNotificationSocketOptions): UseNotificationSocketResult {
  const queryClient = useQueryClient();
  const increment  = useNotificationStore((s) => s.increment);
  const addToast   = useNotificationStore((s) => s.addToast);

  const handleMessage = useCallback(
    (message: ServerEnvelope): void => {
      switch (message.event) {
        case 'notification:new': {
          const payload = message.data as NotificationNewPayload;
          const item: NotificationItem = {
            id:           payload.Notification.ID,
            type:         payload.Notification.Type,
            message:      payload.Notification.Message,
            reference_id: payload.Notification.ReferenceID ?? null,
            created_at:   payload.Notification.CreatedAt,
            actor:        payload.Actor,
          };

          queryClient.setQueryData<{ pages: PaginatedNotificationResponse[]; pageParams: unknown[] }>(
            ['notifications'],
            (old) => {
              if (!old) return old;
              const firstPage = old.pages[0];
              if (!firstPage) return old;
              if (firstPage.data.some((n) => n.id === item.id)) return old;
              return {
                ...old,
                pages: [
                  { ...firstPage, data: [item, ...firstPage.data] },
                  ...old.pages.slice(1),
                ],
              };
            },
          );

          increment();
          addToast(item);
          break;
        }

        case 'notification:deleted': {
          const payload = message.data as NotificationDeletedPayload;
          queryClient.setQueryData<{ pages: PaginatedNotificationResponse[]; pageParams: unknown[] }>(
            ['notifications'],
            (old) => {
              if (!old) return old;
              return {
                ...old,
                pages: old.pages.map((page) => ({
                  ...page,
                  data: page.data.filter((n) => n.id !== payload.id),
                })),
              };
            },
          );
          break;
        }
      }
    },
    [queryClient, increment, addToast],
  );

  const socket = useWebSocket({
    accessToken,
    onMessage: handleMessage,
  });

  return useMemo(() => ({ status: socket.status }), [socket.status]);
}
