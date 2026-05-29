import { api } from '../../api';
import type { Order, PaginatedResponse } from './types';

export function listOrders(token: string, page = 1, limit = 20): Promise<PaginatedResponse<Order>> {
  return api.get<PaginatedResponse<Order>>(`/api/orders?page=${page}&limit=${limit}`, token);
}
