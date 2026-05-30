import Link from 'next/link';
import {
  MessageSquare,
  Repeat2,
  Bookmark,
  Heart,
  Users,
  UserCheck,
  BookOpen,
  Calendar,
  Paperclip,
  Clock,
  GraduationCap,
  UserPlus,
  Trophy,
  AtSign,
  ArrowUpRight,
  ArrowLeft,
  ArrowRight,
} from 'lucide-react';
import type { ComponentType, SVGProps } from 'react';

type Icon = ComponentType<SVGProps<SVGSVGElement> & { size?: number | string }>;

type UseCase = {
  title: string;
  description: string;
  apps: Icon[];
};

const USE_CASES: UseCase[] = [
  {
    title: 'Just hanging out',
    description:
      'Post threads, follow your friends, save the stuff you want to come back to. The everyday Binus feed.',
    apps: [MessageSquare, Repeat2, Bookmark, Heart],
  },
  {
    title: 'Running a club',
    description:
      'Spin up a community for your club or study group. Roles, join requests, and rules — no more chaotic group chats.',
    apps: [Users, UserCheck, BookOpen, MessageSquare],
  },
  {
    title: 'Hosting an event',
    description:
      'Publish a campus event, attach the rundown, get real RSVPs from people who actually go to Binus.',
    apps: [Calendar, Paperclip, MessageSquare, Clock],
  },
  {
    title: 'Finding your people',
    description:
      'See what your faculty is talking about. Find peers in your major before the third week of class.',
    apps: [GraduationCap, UserPlus, Trophy, AtSign],
  },
];

export function AboutUseCases() {
  return (
    <section
      id="use-cases"
      className="about-container"
      style={{
        paddingTop: '96px',
        paddingBottom: '96px',
        borderTop: '1px solid var(--border)',
      }}
    >
      <div
        className="flex items-end justify-between flex-wrap gap-6"
        style={{ marginBottom: '36px' }}
      >
        <div style={{ maxWidth: '40rem' }}>
          <span className="about-eyebrow">Made for the Binus crowd</span>
          <h2
            style={{
              color: 'var(--text)',
              fontWeight: 600,
              fontSize: 'clamp(1.875rem, 3.5vw, 2.625rem)',
              lineHeight: 1.1,
              letterSpacing: '-0.02em',
              marginTop: '16px',
            }}
          >
            However you&apos;re showing up on campus.
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
            Whether you&apos;re posting memes between classes, running a club,
            or just trying to find your faculty&apos;s group chat — there&apos;s
            a place for you on Sunibgram.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link href="/signup" className="about-link-mono">
            See it all <ArrowUpRight size={14} />
          </Link>
          <div className="flex items-center gap-2" aria-hidden>
            <span
              style={{
                width: 36,
                height: 36,
                border: '1px solid var(--border)',
                borderRadius: 6,
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--text-muted)',
              }}
            >
              <ArrowLeft size={16} />
            </span>
            <span
              style={{
                width: 36,
                height: 36,
                border: '1px solid var(--border)',
                borderRadius: 6,
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--text)',
              }}
            >
              <ArrowRight size={16} />
            </span>
          </div>
        </div>
      </div>

      <div
        className="grid gap-5"
        style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))' }}
      >
        {USE_CASES.map((uc) => (
          <div key={uc.title} className="about-usecase-card">
            <div>
              <h3
                style={{
                  color: 'var(--text)',
                  fontWeight: 600,
                  fontSize: '1.25rem',
                  letterSpacing: '-0.01em',
                  marginBottom: '8px',
                }}
              >
                {uc.title}
              </h3>
              <p
                style={{
                  color: 'var(--text-muted)',
                  fontSize: '0.9375rem',
                  lineHeight: 1.55,
                }}
              >
                {uc.description}
              </p>
            </div>

            <div style={{ marginTop: 'auto' }}>
              <div className="flex items-center gap-2" style={{ marginBottom: '16px' }}>
                {uc.apps.map((Icon, i) => (
                  <span
                    key={i}
                    style={{
                      width: 32,
                      height: 32,
                      borderRadius: 8,
                      background: 'var(--bg)',
                      border: '1px solid var(--border)',
                      display: 'inline-flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: 'var(--text)',
                    }}
                  >
                    <Icon size={16} />
                  </span>
                ))}
              </div>
              <Link
                href="/signup"
                className="mono"
                style={{
                  fontSize: '0.75rem',
                  textTransform: 'uppercase',
                  letterSpacing: '0.08em',
                  color: 'var(--text)',
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 6,
                  fontWeight: 600,
                }}
              >
                Try it →
              </Link>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
