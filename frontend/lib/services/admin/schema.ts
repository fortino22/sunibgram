import { api } from '../../api';
import type { SchemaResponse } from './types';

export function getSchema(token: string): Promise<SchemaResponse> {
  return api.get<SchemaResponse>('/api/admin/schema', token);
}
