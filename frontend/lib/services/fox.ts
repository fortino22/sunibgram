import { api } from '../api';
import type { FeedMeta, PaginationParams, ThreadMediaResponse } from './threads';

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
 * POSTs the question + recent history + the chat session id (0 to start a new
 * session). Resolves with the session id the backend used (read from the
 * X-Fox-Session-Id header) so the client can attach the next turn to it. Uses
 * fetch (not EventSource) so the Bearer token and JSON body can be sent.
 */
export async function streamFoxAsk(
  question: string,
  sessionId: number,
  history: FoxQAPair[],
  token: string,
  options: { onChunk: (text: string) => void; signal?: AbortSignal },
): Promise<number | null> {
  const baseUrl = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
  const res = await fetch(`${baseUrl}/api/fox/ask`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ session_id: sessionId, question, history }),
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

  const headerId = res.headers.get('X-Fox-Session-Id');
  const resolvedSession = headerId ? Number.parseInt(headerId, 10) : NaN;

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    const text = decoder.decode(value, { stream: true });
    if (text) options.onChunk(text);
  }

  return Number.isFinite(resolvedSession) && resolvedSession > 0 ? resolvedSession : null;
}

/**
 * One row in the user's Fox history — a single session. Two kinds by `type`:
 *   - "chat":    `title` = first question, `message_count` = turn count,
 *     `preview` = latest answer snippet. Re-openable to continue.
 *   - "summary": `thread_id`/`content`/`summary`/`media` set — read-only.
 */
export interface FoxSessionItem {
  id: number;
  type: 'chat' | 'summary';
  thread_id: number;
  /** Chat sessions: the first question (list label). */
  title: string;
  /** Summary sessions: root post content captured when summarized. */
  content: string;
  /** Summary sessions: generated summary text. */
  summary: string;
  /** Chat sessions: latest answer snippet. */
  preview: string;
  /** Chat sessions: number of turns in the conversation. */
  message_count: number;
  /** Summary sessions: true when the thread changed after summarizing. */
  is_stale: boolean;
  /** "YYYY-MM-DD HH:mm" (WIB) — last activity. */
  updated_at: string;
  /** Summary sessions: current thread images (empty otherwise). */
  media: ThreadMediaResponse[];
}

export interface FoxSessionsResponse {
  data: FoxSessionItem[];
  meta: FeedMeta;
}

function sessionQuery(params: PaginationParams): string {
  const search = new URLSearchParams();
  if (params.page !== undefined) search.set('page', String(params.page));
  if (params.limit !== undefined) search.set('limit', String(params.limit));
  if (params.q !== undefined && params.q !== '') search.set('q', params.q);
  const value = search.toString();
  return value ? `?${value}` : '';
}

export function getFoxSessions(params: PaginationParams = {}, token: string): Promise<FoxSessionsResponse> {
  return api.get<FoxSessionsResponse>(`/api/fox/sessions${sessionQuery(params)}`, token);
}

/** One question/answer pair in a chat session, oldest-first. */
export interface FoxSessionTurn {
  question: string;
  answer: string;
}

export interface FoxSessionDetail {
  id: number;
  turns: FoxSessionTurn[];
}

export function getFoxSession(id: number | string, token: string): Promise<FoxSessionDetail> {
  return api.get<FoxSessionDetail>(`/api/fox/sessions/${id}`, token);
}
