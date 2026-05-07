import { api } from '@/lib/api';
import type { ThreadResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export interface CreateThreadParams {
  content: string;
  files?: File[];
}

function buildCreateThreadFormData(params: CreateThreadParams) {
  const formData = new FormData();
  formData.append('content', params.content);

  params.files?.forEach((file) => {
    formData.append('files', file);
  });

  return formData;
}

export function createThread(params: CreateThreadParams): Promise<ThreadResponse> {
  const token = useAuthStore.getState().token;

  if (token === null) {
    throw new Error('Authentication token is required to create a thread.');
  }

  return api.post<ThreadResponse>('/api/threads', buildCreateThreadFormData(params), token);
}
