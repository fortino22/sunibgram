import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Compress a verbose relative time string ("5 days ago", "1 minute ago",
// "Just now") into an X-style short form ("5d", "1m", "now"). Leaves
// already-short values and absolute dates unchanged.
export function shortTimeAgo(value: string | null | undefined): string {
  if (!value) return '';
  const trimmed = value.trim();
  if (/^(just\s*now|now)$/i.test(trimmed)) return 'now';
  const match = trimmed.match(/^(\d+)\s*(second|minute|hour|day|week|month|year)s?(?:\s+ago)?$/i);
  if (!match) return trimmed;
  const n = match[1];
  const unit = match[2].toLowerCase();
  const map: Record<string, string> = {
    second: 's',
    minute: 'm',
    hour: 'h',
    day: 'd',
    week: 'w',
    month: 'mo',
    year: 'y',
  };
  return `${n}${map[unit] ?? ''}`;
}

// Parses "YYYY-MM-DD HH:mm" from backend and returns a relative time string.
export function timeAgo(dateStr: string): string {
  if (!dateStr) return '';
  const parsed = new Date(dateStr.replace(' ', 'T'));
  if (isNaN(parsed.getTime()) || parsed.getFullYear() < 2000) return '';
  const diff = Date.now() - parsed.getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  const weeks = Math.floor(days / 7);
  if (weeks < 5) return `${weeks}w ago`;
  return parsed.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}
