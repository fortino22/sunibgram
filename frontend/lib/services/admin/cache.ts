import { api } from '../../api';
import type { MessageResponse } from './types';

export function clearCache(token: string): Promise<MessageResponse> {
  return api.post<MessageResponse>('/api/admin/cache/clear', {}, token);
}
