import {
  Home,
  Compass,
  Users,
  Bell,
  Bookmark,
  User,
  Plus,
  MoreHorizontal,
  TrendingUp,
  Search,
  ArrowRight,
  MessageCircle,
  MessageSquare,
  Image,
  Video,
  Hash,
  BarChart2,
  Heart,
  Repeat2,
  BookmarkCheck,
  Calendar,
  Trophy,
  Shield,
  Settings,
  Lock,
  KeyRound,
  Sparkles,
  UserCog,
  ChevronRight,
  Megaphone,
  Palette,
  Check,
  SquarePen,
  UserX,
  VolumeX,
  type LucideProps,
} from 'lucide-react';
import type { FC } from 'react';

const ICON_MAP: Record<string, FC<LucideProps>> = {
  home:         Home,
  explore:      Compass,
  users:        Users,
  bell:         Bell,
  bookmark:     Bookmark,
  bookmarkFill: BookmarkCheck,
  user:         User,
  plus:         Plus,
  more:         MoreHorizontal,
  trending:     TrendingUp,
  search:       Search,
  arrowRight:   ArrowRight,
  reply:        MessageCircle,
  messages:     MessageSquare,
  image:        Image,
  video:        Video,
  hash:         Hash,
  poll:         BarChart2,
  heart:        Heart,
  repost:       Repeat2,
  calendar:     Calendar,
  trophy:       Trophy,
  shield:       Shield,
  settings:     Settings,
  lock:         Lock,
  key:          KeyRound,
  sparkles:     Sparkles,
  userCog:      UserCog,
  chevronRight: ChevronRight,
  feedback:     Megaphone,
  palette:      Palette,
  check:        Check,
  compose:      SquarePen,
  userX:        UserX,
  volumeX:      VolumeX,
};

interface IconProps {
  name: string;
  size?: number;
  color?: string;
  strokeWidth?: number;
  fill?: string;
}

export function Icon({ name, size = 18, color = 'currentColor', strokeWidth = 1.8, fill = 'none' }: IconProps) {
  const LucideIcon = ICON_MAP[name];
  if (!LucideIcon) return null;
  return <LucideIcon size={size} color={color} strokeWidth={strokeWidth} fill={fill} />;
}
