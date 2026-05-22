import { api } from '@/lib/api';
import type { ThreadPollInput, ThreadResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';

export interface CreateThreadParams {
  content: string;
  files?: File[];
  poll?: ThreadPollInput;
}

function buildCreateThreadFormData(params: CreateThreadParams) {
  const formData = new FormData();
  formData.append('content', params.content);

  params.files?.forEach((file) => {
    formData.append('files', file);
  });

  if (params.poll) {
    formData.append('poll_question', params.poll.question?.trim() ?? '');
    formData.append('poll_duration_seconds', String(params.poll.durationSeconds));
    params.poll.options.forEach((option) => {
      formData.append('poll_options', option.label);
    });
    const optionImages = params.poll.options.map((option) => option.image).filter((image): image is File => image instanceof File);
    if (optionImages.length === params.poll.options.length) {
      optionImages.forEach((image) => {
        formData.append('poll_option_images', image);
      });
    }
  }

  return formData;
}

export function createThread(params: CreateThreadParams): Promise<ThreadResponse> {
  const token = useAuthStore.getState().token;

  if (token === null) {
    throw new Error('Authentication token is required to create a thread.');
  }

  return api.post<ThreadResponse>('/api/threads', buildCreateThreadFormData(params), token);
}
