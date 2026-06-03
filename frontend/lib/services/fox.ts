/** One prior turn replayed to the backend for short-term chat context. */
export interface FoxQAPair {
  question: string;
  answer: string;
}

/** Thrown when the daily Fox-chat token budget is exhausted (HTTP 429). */
export class FoxLimitError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'FoxLimitError';
  }
}

/**
 * Streams the Fox guide answer as plain text. Calls onChunk with each decoded
 * piece as it arrives so the UI can render the text live (typewriter effect).
 * POSTs the question + recent history; uses fetch (not EventSource) so the
 * Bearer token and JSON body can be sent.
 */
export async function streamFoxAsk(
  question: string,
  history: FoxQAPair[],
  token: string,
  options: { onChunk: (text: string) => void; signal?: AbortSignal },
): Promise<void> {
  const baseUrl = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
  const res = await fetch(`${baseUrl}/api/fox/ask`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ question, history }),
    signal: options.signal,
  });

  if (res.status === 429) {
    const fallback = "You've reached your daily Fox chat limit. It resets tomorrow.";
    const message = await res
      .json()
      .then((body: { message?: string }) => body.message ?? fallback)
      .catch(() => fallback);
    throw new FoxLimitError(message);
  }

  if (!res.ok || res.body === null) {
    throw new Error('Failed to reach Fox');
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    const text = decoder.decode(value, { stream: true });
    if (text) options.onChunk(text);
  }
}
