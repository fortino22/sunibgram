import type { Metadata } from 'next';
import Link from 'next/link';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutUseCases } from '@/components/about/AboutUseCases';
import { AboutStatMoment } from '@/components/about/AboutStatMoment';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';
import {
  GraduationCap,
  Users,
  CalendarDays,
  MessageCircle,
  Trophy,
  Library,
} from 'lucide-react';

export const metadata: Metadata = {
  title: 'Use Cases — Sunibgram',
  description:
    'See how Binus students use Sunibgram every day — from finding study spots to running clubs and organizing events.',
};

type Scenario = {
  Icon: typeof GraduationCap;
  audience: string;
  title: string;
  body: string;
  steps: string[];
  cta: { label: string; href: string };
};

const SCENARIOS: Scenario[] = [
  {
    Icon: GraduationCap,
    audience: 'First-year student',
    title: "Find your faculty before week 1.",
    body:
      'Verify with your @binus.ac.id email, follow your faculty feed, and the first names you see are people who share your hallways — not strangers on the internet.',
    steps: [
      'Sign up with your Binus email and pick your faculty.',
      'Open the Faculty tab to see what classmates are talking about.',
      'Follow a few people and join the welcome thread.',
    ],
    cta: { label: 'Sign up', href: '/signup' },
  },
  {
    Icon: Users,
    audience: 'Club president',
    title: 'Run your community without the chat chaos.',
    body:
      'Create a community, set rules, promote moderators, and move the noisy WhatsApp groups into one place with proper threads and pinned announcements.',
    steps: [
      'Create a public or private community.',
      'Promote trusted members to ADMIN so they can manage rules and members.',
      'Pin the latest announcement so newcomers see it first.',
    ],
    cta: { label: 'How communities work', href: '/about/communities' },
  },
  {
    Icon: Library,
    audience: 'Study group',
    title: 'Coordinate around an actual schedule.',
    body:
      'Ask one question in your faculty feed, get answers from the people in your batch. Turn it into a private community when you find your crew.',
    steps: [
      'Post the question in your faculty feed.',
      'DM the people who replied to lock in a time and place.',
      'Spin up a private community for ongoing study chats and shared files.',
    ],
    cta: { label: 'How DMs work', href: '/about/messages' },
  },
  {
    Icon: CalendarDays,
    audience: 'Event organizer',
    title: 'Publish once, reach the whole campus.',
    body:
      'Submit your event with the right details, get verified for official ones, and let the RSVP counter handle the rest. Event threads keep questions and updates in one place.',
    steps: [
      'Submit the event with date, location, and audience.',
      'Wait for faculty approval if you want the Official badge.',
      'Use the event thread for live updates during the day.',
    ],
    cta: { label: 'How events work', href: '/about/events' },
  },
  {
    Icon: MessageCircle,
    audience: 'Project team',
    title: 'Move project chat off your personal phone.',
    body:
      'Create a group room with just your teammates, share files inline, and stop digging through screenshots in WhatsApp at 11 PM.',
    steps: [
      'Create a group chat and add your teammates.',
      'Share notes, slides, and deliverables — they stay in the room.',
      'Edit a typo, unsend a wrong file — nothing leaves the room.',
    ],
    cta: { label: 'How messages work', href: '/about/messages' },
  },
  {
    Icon: Trophy,
    audience: 'Rising voice',
    title: 'Get recognized for actually showing up.',
    body:
      'The leaderboard surfaces students whose threads spark replies and reposts from verified peers — not the loudest accounts, the most useful ones.',
    steps: [
      'Post things that classmates would want to talk about.',
      'Reply substantively in faculty threads and communities.',
      'Check your weekly score and the faculty board.',
    ],
    cta: { label: 'How the leaderboard works', href: '/about/leaderboard' },
  },
];

export default function UseCasesPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="Use cases"
        title="However you show up on campus."
        description="Sunibgram is one app for first-years finding their faculty, clubs running their rooms, study groups coordinating, and organizers running real events."
        primaryCta={{ label: 'Join with your Binus email', href: '/signup' }}
        secondaryCta={{ label: 'See the features', href: '/about/threads' }}
      />

      <AboutUseCases />

      {/* Scenarios with steps */}
      <section className="about-content-section">
        <div className="about-container">
          <div style={{ textAlign: 'center', maxWidth: '40rem', margin: '0 auto 36px' }}>
            <span className="about-content-eyebrow" style={{ justifyContent: 'center' }}>
              Real flows
            </span>
            <h2 className="about-content-h2" style={{ marginInline: 'auto' }}>
              Six scenarios, end-to-end.
            </h2>
            <p className="about-content-lede" style={{ marginInline: 'auto' }}>
              Here&apos;s exactly what each kind of user does on Sunibgram — three steps,
              no fluff.
            </p>
          </div>

          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
              gap: 18,
              maxWidth: '64rem',
              margin: '0 auto',
            }}
          >
            {SCENARIOS.map(({ Icon, audience, title, body, steps, cta }) => (
              <article
                key={title}
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  border: '1px solid var(--border)',
                  borderRadius: 16,
                  background: 'var(--bg-card)',
                  padding: '24px 26px',
                  gap: 16,
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <div
                    style={{
                      width: 40,
                      height: 40,
                      borderRadius: 10,
                      background: 'var(--accent-soft)',
                      color: 'var(--accent)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <Icon size={20} strokeWidth={2} />
                  </div>
                  <div style={{ color: 'var(--text-tertiary)', fontSize: 12, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase' }}>
                    {audience}
                  </div>
                </div>
                <h3
                  style={{
                    color: 'var(--text)',
                    fontSize: '1.25rem',
                    fontWeight: 600,
                    letterSpacing: '-0.015em',
                    lineHeight: 1.25,
                  }}
                >
                  {title}
                </h3>
                <p style={{ color: 'var(--text-muted)', fontSize: 15, lineHeight: 1.6 }}>{body}</p>
                <ol
                  style={{
                    listStyle: 'none',
                    padding: 0,
                    margin: 0,
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 10,
                  }}
                >
                  {steps.map((step, i) => (
                    <li
                      key={i}
                      style={{
                        display: 'flex',
                        gap: 12,
                        alignItems: 'flex-start',
                        color: 'var(--text)',
                        fontSize: 14,
                        lineHeight: 1.5,
                      }}
                    >
                      <span
                        style={{
                          flexShrink: 0,
                          width: 22,
                          height: 22,
                          borderRadius: 999,
                          background: 'var(--bg-subtle)',
                          color: 'var(--text-muted)',
                          fontSize: 11,
                          fontWeight: 700,
                          display: 'inline-flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontVariantNumeric: 'tabular-nums',
                        }}
                      >
                        {i + 1}
                      </span>
                      <span>{step}</span>
                    </li>
                  ))}
                </ol>
                <div style={{ marginTop: 'auto', paddingTop: 4 }}>
                  <Link
                    href={cta.href}
                    style={{
                      color: 'var(--accent)',
                      fontSize: 14,
                      fontWeight: 700,
                      textDecoration: 'none',
                    }}
                  >
                    {cta.label} →
                  </Link>
                </div>
              </article>
            ))}
          </div>
        </div>
      </section>

      <AboutStatMoment />
      <AboutFinalCTA />
    </>
  );
}
