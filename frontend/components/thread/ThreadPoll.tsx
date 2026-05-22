'use client';

import { Check, Loader2 } from 'lucide-react';
import { useEffect, useState, type KeyboardEvent, type MouseEvent } from 'react';
import { useVotePoll } from '@/hooks/useVotePoll';
import type { ThreadItem, ThreadMediaResponse } from '@/lib/services/threads';
import { useAuthStore } from '@/stores/authStore';
import { getPollMedia, getPollOptionMedia, isOwnThread } from './threadUtils';

interface ThreadPollProps {
  thread: ThreadItem;
  compact?: boolean;
  onImageClick?: (media: ThreadMediaResponse[], index: number) => void;
}

function formatPercent(value: number): string {
  return Number.isInteger(value) ? `${value.toFixed(0)}%` : `${value.toFixed(1)}%`;
}

function formatVoteCount(totalVotes: number): string {
  return `${totalVotes} vote${totalVotes === 1 ? '' : 's'}`;
}

function parsePollEndTime(endsAt: string): number | null {
  const end = new Date(endsAt);
  const endMs = end.getTime();

  if (Number.isNaN(endMs)) {
    return null;
  }

  return endMs;
}

function isPollExpired(endsAt: string, now: number, fallback = false): boolean {
  const endMs = parsePollEndTime(endsAt);

  if (endMs === null) {
    return fallback;
  }

  return now >= endMs;
}

function formatTimeRemaining(endsAt: string, now: number): string {
  const endMs = parsePollEndTime(endsAt);

  if (endMs === null) {
    return 'Final results';
  }

  const diffMs = endMs - now;
  if (diffMs <= 0) {
    return 'Final results';
  }

  const totalSeconds = Math.ceil(diffMs / 1000);
  const days = Math.floor(totalSeconds / 86_400);
  const hours = Math.floor((totalSeconds % 86_400) / 3_600);
  const minutes = Math.floor((totalSeconds % 3_600) / 60);
  const seconds = totalSeconds % 60;

  if (days > 0) {
    return `${days}d ${hours}h ${minutes}m ${seconds}s left`;
  }

  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s left`;
  }

  if (minutes > 0) {
    return `${minutes}m ${seconds}s left`;
  }

  return `${seconds}s left`;
}

function optionFillColor(isChosen: boolean, canVote: boolean): string {
  if (isChosen) {
    return 'color-mix(in srgb, var(--accent) 22%, transparent)';
  }

  return canVote
    ? 'color-mix(in srgb, var(--text-tertiary) 10%, transparent)'
    : 'color-mix(in srgb, var(--text-tertiary) 14%, transparent)';
}

function handleOptionVote(
  event: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>,
  canVote: boolean,
  optionId: number,
  mutate: (optionId: number) => void
) {
  if (!canVote) {
    return;
  }

  event.stopPropagation();
  mutate(optionId);
}

export function ThreadPoll({ thread, compact = false, onImageClick }: ThreadPollProps) {
  const poll = thread.poll;
  const voteMutation = useVotePoll(thread);
  const currentUserId = useAuthStore((state) => state.user?.id);
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    if (!poll) {
      return undefined;
    }

    const endMs = parsePollEndTime(poll.ends_at);
    const currentTime = Date.now();

    if (endMs === null || currentTime >= endMs) {
      return undefined;
    }

    const timer = window.setInterval(() => {
      const nextNow = Date.now();
      setNow(nextNow);

      if (nextNow >= endMs) {
        window.clearInterval(timer);
      }
    }, 1_000);

    return () => window.clearInterval(timer);
  }, [poll]);

  if (!poll) {
    return null;
  }

  const isPollAuthor = isOwnThread(thread, currentUserId);
  const hasVoted = poll.user_voted_option_id !== undefined && poll.user_voted_option_id !== null;
  const isExpired = isPollExpired(poll.ends_at, now, poll.expired);
  const canVote = !isPollAuthor && !isExpired && !hasVoted && !voteMutation.isPending;
  const metadata = `${formatVoteCount(poll.total_votes)} · ${isExpired ? 'Final results' : formatTimeRemaining(poll.ends_at, now)}`;
  const pollOptionMedia = getPollOptionMedia(thread);
  const pollMedia = getPollMedia(thread);

  return (
    <div
      onClickCapture={(event) => event.stopPropagation()}
      style={{ marginTop: compact ? 12 : 16 }}
    >
      <div className="flex flex-col" style={{ gap: 8 }}>
        {poll.options.map((option, optionIndex) => {
          const isChosen = option.option_id === poll.user_voted_option_id;
          const isDisabled = !canVote;
          const optionMedia = pollOptionMedia[optionIndex] ?? null;
          const optionMediaIndex =
            optionMedia === null
              ? -1
              : pollMedia.findIndex((item) => item.media_id === optionMedia.media_id && item.url === optionMedia.url);

          return (
            <div
              key={option.option_id}
              role={canVote ? 'button' : undefined}
              aria-disabled={isDisabled}
              tabIndex={isDisabled ? -1 : 0}
              onClick={(event) => handleOptionVote(event, canVote, option.option_id, voteMutation.mutate)}
              onKeyDown={canVote ? (event) => {
                if (event.key === 'Enter' || event.key === ' ') {
                  event.preventDefault();
                  handleOptionVote(event, canVote, option.option_id, voteMutation.mutate);
                }
              } : undefined}
              className="relative w-full overflow-hidden rounded-2xl border text-left transition-colors"
              style={{
                borderColor: isChosen ? 'var(--accent)' : 'var(--border)',
                background: 'var(--bg-card)',
                cursor: canVote ? 'pointer' : 'default',
                opacity: voteMutation.isPending && !isChosen ? 0.92 : 1,
              }}
            >
              <div
                className="absolute inset-y-0 left-0 rounded-2xl"
                style={{
                  width: `${Math.max(0, Math.min(100, option.percent))}%`,
                  background: optionFillColor(isChosen, canVote),
                }}
              />

              <div
                className="relative flex items-center justify-between"
                style={{
                  gap: compact ? 10 : 12,
                  padding: compact ? '10px 12px' : '12px 14px',
                }}
              >
                <div className="min-w-0 flex-1">
                  <div className="flex min-w-0 items-center" style={{ gap: compact ? 10 : 12 }}>
                    {optionMedia && (
                      <button
                        type="button"
                        aria-label={`Open image for ${option.label}`}
                        onClick={onImageClick && optionMediaIndex >= 0 ? (event) => {
                          event.stopPropagation();
                          onImageClick(pollMedia, optionMediaIndex);
                        } : (event) => {
                          event.stopPropagation();
                        }}
                        className="shrink-0 overflow-hidden rounded-xl border"
                        style={{
                          width: compact ? 40 : 48,
                          height: compact ? 40 : 48,
                          padding: 0,
                          borderColor: 'var(--border)',
                          background: 'var(--bg-subtle)',
                          cursor: onImageClick && optionMediaIndex >= 0 ? 'zoom-in' : 'default',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                        }}
                      >
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img
                          src={optionMedia.url}
                          alt=""
                          width={compact ? 40 : 48}
                          height={compact ? 40 : 48}
                          style={{
                            width: '100%',
                            height: '100%',
                            objectFit: 'cover',
                            display: 'block',
                            flexShrink: 0,
                          }}
                        />
                      </button>
                    )}

                    <span
                      className="block truncate"
                      style={{
                        color: 'var(--text)',
                        fontSize: compact ? 13 : 14,
                        fontWeight: isChosen ? 700 : 600,
                      }}
                    >
                      {option.label}
                    </span>
                  </div>
                </div>

                <div className="flex shrink-0 items-center" style={{ gap: 8 }}>
                  {isChosen && (
                    <span
                      className="inline-flex items-center justify-center rounded-full"
                      style={{
                        width: compact ? 18 : 20,
                        height: compact ? 18 : 20,
                        background: 'var(--accent)',
                        color: '#fff',
                      }}
                    >
                      <Check size={compact ? 11 : 12} />
                    </span>
                  )}
                  <span
                    style={{
                      color: 'var(--text)',
                      fontSize: compact ? 13 : 14,
                      fontWeight: 700,
                      minWidth: compact ? 36 : 40,
                      textAlign: 'right',
                    }}
                  >
                    {formatPercent(option.percent)}
                  </span>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      <div
        className="flex items-center"
        style={{
          gap: 8,
          marginTop: 10,
          color: 'var(--text-tertiary)',
          fontSize: compact ? 12 : 13,
        }}
      >
        {voteMutation.isPending && <Loader2 size={compact ? 13 : 14} className="animate-spin" />}
        <span>{metadata}</span>
      </div>

      {voteMutation.isError && (
        <div
          className="rounded-xl"
          style={{
            marginTop: 10,
            padding: '8px 10px',
            background: 'var(--coral-soft)',
            color: 'var(--coral)',
            fontSize: compact ? 12 : 13,
            lineHeight: 1.5,
          }}
        >
          {voteMutation.error.message}
        </div>
      )}
    </div>
  );
}
