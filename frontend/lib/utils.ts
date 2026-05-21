import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
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
