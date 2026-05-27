import { Heart, Repeat2, UserPlus, MessageCircle, AtSign, Bookmark, MessageSquare, Bell } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

interface NotificationTypeConfig {
  Icon: LucideIcon;
  color: string;
}

const TYPE_MAP: Record<string, NotificationTypeConfig> = {
  like:         { Icon: Heart,          color: '#e11d48' },
  repost:       { Icon: Repeat2,        color: '#22c55e' },
  follow:       { Icon: UserPlus,       color: '#3b82f6' },
  comment:      { Icon: MessageCircle,  color: '#6366f1' },
  mention:      { Icon: AtSign,         color: '#a855f7' },
  bookmark:     { Icon: Bookmark,       color: '#f59e0b' },
  chat_message: { Icon: MessageSquare,  color: '#0ea5e9' },
};

const FALLBACK: NotificationTypeConfig = { Icon: Bell, color: '#6b7280' };

export function getNotificationTypeConfig(type: string): NotificationTypeConfig {
  return TYPE_MAP[type] ?? FALLBACK;
}
