import { api } from '../api';

export interface OrderMessageResponse {
  message: string;
}

export interface Order {
  id: number | string;
  status?: string;
  user_id?: number;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
}

export interface OrderResponse {
  data: Order;
  message?: string;
  [key: string]: unknown;
}

export interface OrderListResponse {
  data: Order[];
  [key: string]: unknown;
}

async function patch<T>(path: string, token: string): Promise<T> {
  const baseUrl = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  const res = await fetch(`${baseUrl}${normalizedPath}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
  });

  if (!res.ok) {
    const error: unknown = await res.json().catch(() => ({ message: 'Request failed' }));
    const message =
      typeof error === 'object' &&
      error !== null &&
      'message' in error &&
      typeof error.message === 'string'
        ? error.message
        : 'Request failed';
    throw new Error(message);
  }

  return res.json() as Promise<T>;
}

export function createPremiumOrder(token: string): Promise<OrderResponse> {
  return api.post<OrderResponse>('/api/orders/premium', undefined, token);
}

export function createOrderNotification(token: string): Promise<OrderMessageResponse> {
  return api.post<OrderMessageResponse>('/api/orders/notification', undefined, token);
}

export function getMyOrders(token: string): Promise<OrderListResponse> {
  return api.get<OrderListResponse>('/api/orders/me', token);
}

export function getOrder(id: number | string, token: string): Promise<OrderResponse> {
  return api.get<OrderResponse>(`/api/orders/${id}`, token);
}

export function getOrders(token: string): Promise<OrderListResponse> {
  return api.get<OrderListResponse>('/api/orders', token);
}

export function cancelOrder(id: number | string, token: string): Promise<OrderResponse> {
  return patch<OrderResponse>(`/api/orders/${id}/cancel`, token);
}
