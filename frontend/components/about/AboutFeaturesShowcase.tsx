'use client';

import { useRef } from 'react';
import Link from 'next/link';
import {
  MessageSquare,
  Repeat2,
  Heart,
  Bookmark,
  Users,
  Shield,
  UserCheck,
  BookOpen,
  Send,
  AtSign,
  Bell,
  Lock,
  Calendar,
  MapPin,
  Clock,
  Ticket,
  Trophy,
  TrendingUp,
  Star,
  Award,
  Sparkles,
  BarChart3,
  FileText,
  Brain,
  ArrowUpRight,
  ArrowLeft,
  ArrowRight,
} from 'lucide-react';
import type { ComponentType, SVGProps } from 'react';

type Icon = ComponentType<SVGProps<SVGSVGElement> & { size?: number | string }>;

type Feature = {
  title: string;
  description: string;
  href: string;
  apps: Icon[];
  badge?: string;
};

const FEATURES: Feature[] = [
  {
    title: 'Threads',
    description:
      'Post what you think, reply to your faculty, like and bookmark — the verified Binus feed, no anonymous noise.',
    href: '/about/threads',
    apps: [MessageSquare, Repeat2, Heart, Bookmark],
  },
  {
    title: 'Communities',
    description:
      'Run your club or study group with real roles, join requests, and rules. Group chats grown up.',
    href: '/about/communities',
    apps: [Users, Shield, UserCheck, BookOpen],
  },
  {
    title: 'Messages',
    description:
      'Real-time DMs and group chats with the people you already share campus with — no usernames to guess.',
    href: '/about/messages',
    apps: [Send, AtSign, Bell, Lock],
  },
  {
    title: 'Events',
    description:
      'Publish a campus event, attach the rundown, and get real RSVPs from students who actually go to Binus.',
    href: '/about/events',
    apps: [Calendar, MapPin, Clock, Ticket],
  },
  {
    title: 'Leaderboard',
    description:
      'See top contributors and rising voices on campus — recognition for the people carrying the conversation.',
    href: '/about/leaderboard',
    apps: [Trophy, TrendingUp, Star, Award],
  },
  {
    title: 'Fox AI',
    description:
      'The on-platform analyst that reviews events and writes daily and weekly reports for the team.',
    href: '/about/fox',
    apps: [Sparkles, BarChart3, FileText, Brain],
    badge: 'Developer',
  },
];

export function AboutFeaturesShowcase() {
  const trackRef = useRef<HTMLDivElement>(null);

  const scrollBy = (dir: 1 | -1) => {
    const el = trackRef.current;
    if (!el) return;
    const card = el.querySelector<HTMLElement>('.about-feature-card');
    const step = card ? card.offsetWidth + 20 : el.clientWidth * 0.8;
    el.scrollBy({ left: dir * step, behavior: 'smooth' });
  };

  return (
    <section
      id="features"
      className="about-container"
      style={{ paddingTop: '96px', paddingBottom: '96px' }}
    >
      <div
        className="flex items-end justify-between flex-wrap gap-6"
        style={{ marginBottom: '36px' }}
      >
        <div style={{ maxWidth: '40rem' }}>
          <span className="about-eyebrow">Features</span>
          <h2
            style={{
              color: 'var(--text)',
              fontWeight: 500,
              fontSize: 'clamp(1.75rem, 3.5vw, 2.75rem)',
              lineHeight: 1.12,
              letterSpacing: '-0.02em',
              marginTop: '16px',
            }}
          >
            Built for how you Binus.
          </h2>
          <p
            style={{
              color: 'var(--text-muted)',
              fontSize: '1rem',
              lineHeight: 1.6,
              marginTop: '16px',
              maxWidth: '34rem',
            }}
          >
            Sunibgram has the features your campus life already needs —
            threads, communities, events, messages — wired to one verified
            Binus identity.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link href="/about/threads" className="about-link-mono">
            All features <ArrowUpRight size={14} />
          </Link>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => scrollBy(-1)}
              aria-label="Scroll features left"
              className="about-feature-nav-btn"
            >
              <ArrowLeft size={16} />
            </button>
            <button
              type="button"
              onClick={() => scrollBy(1)}
              aria-label="Scroll features right"
              className="about-feature-nav-btn"
            >
              <ArrowRight size={16} />
            </button>
          </div>
        </div>
      </div>

      <div
        ref={trackRef}
        className="about-feature-track"
      >
        {FEATURES.map((f) => (
          <Link
            key={f.title}
            href={f.href}
            className="about-feature-card"
          >
            <div className="about-feature-card-bg" aria-hidden />
            <div className="about-feature-card-grid" aria-hidden />
            <div className="about-feature-card-grain" aria-hidden />

            <div className="about-feature-card-body">
              <h3 className="about-feature-card-title">
                {f.title}
                {f.badge && (
                  <span className="about-nav-dev-badge">{f.badge}</span>
                )}
              </h3>
              <p className="about-feature-card-desc">{f.description}</p>
            </div>

            <div className="about-feature-card-foot">
              <div className="flex items-center gap-2" style={{ marginBottom: '18px' }}>
                {f.apps.map((Icon, i) => (
                  <span key={i} className="about-feature-card-app">
                    <Icon size={16} />
                  </span>
                ))}
              </div>
              <span className="about-feature-card-explore">
                Explore <ArrowRight size={14} />
              </span>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
}
