import type { Metadata } from 'next';
import Link from 'next/link';
import {
  Rocket,
  ShieldCheck,
  MessageSquare,
  Users,
  CalendarDays,
  Trophy,
  Mail,
} from 'lucide-react';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutFAQ, type FAQItem } from '@/components/about/AboutFAQ';

export const metadata: Metadata = {
  title: 'Help Center — Sunibgram',
  description:
    'Guides, FAQs, and answers to the questions Binus students ask most often about Sunibgram.',
};

const CATEGORIES = [
  {
    Icon: Rocket,
    title: 'Getting started',
    desc: 'Sign up with your Binus email, finish onboarding, set up your profile.',
    href: '/signup',
  },
  {
    Icon: ShieldCheck,
    title: 'Account & privacy',
    desc: 'Verification, private accounts, blocked users, password changes.',
    href: '/about/privacy',
  },
  {
    Icon: MessageSquare,
    title: 'Threads & replies',
    desc: 'Posting, editing, deleting, replying, polls, and reposts.',
    href: '/about/threads',
  },
  {
    Icon: Users,
    title: 'Communities',
    desc: 'Joining, leaving, requesting, moderating, and managing members.',
    href: '/about/communities',
  },
  {
    Icon: CalendarDays,
    title: 'Events',
    desc: 'Submitting, getting Official approval, RSVPing, and event threads.',
    href: '/about/events',
  },
  {
    Icon: Trophy,
    title: 'Leaderboard',
    desc: 'Weekly scoring, faculty boards, and how your engagement is counted.',
    href: '/about/leaderboard',
  },
];

const FAQS: FAQItem[] = [
  {
    q: 'Who can sign up for Sunibgram?',
    a: (
      <>
        Only current Binus students with a verified <strong>@binus.ac.id</strong> email
        address. Accounts are locked to one per student, and verification is required
        before you can post or reply.
      </>
    ),
  },
  {
    q: 'I never received my OTP email. What now?',
    a: (
      <>
        Check your spam folder first — university filtering sometimes catches our mail.
        If it&apos;s still missing after a few minutes, request a new code from the
        verification screen. Repeated failures usually mean your @binus.ac.id mailbox is
        full or your faculty has temporarily blocked external sender domains.
      </>
    ),
  },
  {
    q: 'Can I change my username after sign up?',
    a: (
      <>
        Your display name and profile details can be edited at any time from your
        profile settings. Usernames (the <code>@handle</code>) are stable to prevent
        identity confusion across mentions — reach out to support if you have a strong
        reason to change yours.
      </>
    ),
  },
  {
    q: 'How do public and private communities differ?',
    a: (
      <>
        Public communities let any verified Binus student join with one tap and see all
        threads. Private communities require the owner or an admin to approve each join
        request, and only members can read threads inside them.
      </>
    ),
  },
  {
    q: 'Why are some events marked “Official” and others not?',
    a: (
      <>
        Official events have been reviewed and approved by an admin from the faculty or
        student council that submitted them. Unofficial events are open to anyone to
        publish — they show up in the feed, just without the verified badge.
      </>
    ),
  },
  {
    q: 'What counts toward my leaderboard score?',
    a: (
      <>
        Replies you receive count the most, followed by reposts and likes from other
        verified students. Posting frequently helps, but quality interactions weigh
        heavier than raw post volume. Bots and self-engagement are filtered out.
      </>
    ),
  },
  {
    q: 'How do I report a thread, message, or user?',
    a: (
      <>
        Use the three-dot menu on any thread or profile and pick <strong>Report</strong>.
        You&apos;ll select a violation category and add a short explanation. Reports go
        to our moderation queue and are usually reviewed within a couple of days.
      </>
    ),
  },
  {
    q: 'Can I delete my account?',
    a: (
      <>
        Yes — open your account settings and choose <strong>Delete account</strong>.
        Your threads, replies, communities, and messages are removed permanently from
        the active app. Some logs may be retained for a short period for safety and
        legal reasons, as described in the Privacy Center.
      </>
    ),
  },
];

export default function HelpAboutPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="Help center"
        title="How can we help?"
        description="Guides, FAQs, and direct answers — written for the questions Binus students actually ask. If you can&apos;t find what you need, write to us."
        primaryCta={{ label: 'Email support', href: 'mailto:support@sunibgram.com' }}
        secondaryCta={{ label: 'Privacy center', href: '/about/privacy' }}
      />

      {/* Categories */}
      <section className="about-content-section">
        <div className="about-container">
          <div style={{ textAlign: 'center', maxWidth: '40rem', margin: '0 auto 36px' }}>
            <span className="about-content-eyebrow" style={{ justifyContent: 'center' }}>
              Browse by topic
            </span>
            <h2 className="about-content-h2" style={{ marginInline: 'auto' }}>
              Pick the area you need help with.
            </h2>
          </div>
          <div className="about-help-cats">
            {CATEGORIES.map(({ Icon, title, desc, href }) => (
              <Link key={title} href={href} className="about-help-cat">
                <div className="about-help-cat-icon">
                  <Icon size={16} strokeWidth={2} />
                </div>
                <div className="about-help-cat-title">{title}</div>
                <div className="about-help-cat-desc">{desc}</div>
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* FAQ */}
      <section className="about-content-section">
        <div className="about-container">
          <div style={{ textAlign: 'center', maxWidth: '40rem', margin: '0 auto 24px' }}>
            <span className="about-content-eyebrow" style={{ justifyContent: 'center' }}>
              Frequently asked
            </span>
            <h2 className="about-content-h2" style={{ marginInline: 'auto' }}>
              The questions we get the most.
            </h2>
          </div>
          <AboutFAQ items={FAQS} />
        </div>
      </section>

      {/* Contact support */}
      <section className="about-content-section">
        <div className="about-container">
          <div
            style={{
              maxWidth: '56rem',
              margin: '0 auto',
              border: '1px solid var(--border)',
              borderRadius: 20,
              background: 'var(--bg-card)',
              padding: '32px 36px',
              display: 'flex',
              alignItems: 'center',
              gap: 24,
              flexWrap: 'wrap',
            }}
          >
            <div
              style={{
                width: 52,
                height: 52,
                borderRadius: 14,
                background: 'var(--accent-soft)',
                color: 'var(--accent)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Mail size={24} strokeWidth={2} />
            </div>
            <div style={{ flex: 1, minWidth: 220 }}>
              <div style={{ color: 'var(--text)', fontSize: '1.25rem', fontWeight: 600, letterSpacing: '-0.01em' }}>
                Still stuck?
              </div>
              <div style={{ color: 'var(--text-muted)', fontSize: 14, lineHeight: 1.5, marginTop: 4 }}>
                Email us at <span style={{ color: 'var(--text)', fontWeight: 600 }}>support@sunibgram.com</span>.
                We read everything — usually within a couple of working days.
              </div>
            </div>
            <Link
              href="mailto:support@sunibgram.com"
              className="about-btn-primary about-btn-primary-inverted"
              style={{ color: 'var(--text)', borderColor: 'rgba(0, 0, 0, 0.18)' }}
            >
              Contact support
            </Link>
          </div>
        </div>
      </section>
    </>
  );
}
