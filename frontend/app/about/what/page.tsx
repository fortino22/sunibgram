import type { Metadata } from 'next';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutSplitValue } from '@/components/about/AboutSplitValue';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';
import { AboutWhatDiagram } from '@/components/about/AboutWhatDiagram';
import { AboutPrincipleCards } from '@/components/about/AboutPrincipleCards';

export const metadata: Metadata = {
  title: 'What is Sunibgram — The social network for Binus',
  description:
    'Sunibgram is the campus social network for Binus students — verified, on-campus, built by students. No bots, no anonymous noise.',
};

export default function WhatPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="What is Sunibgram"
        title="A social network built around one campus."
        description="Sunibgram is a verified-only platform for Binus students. Threads, communities, events, and DMs — all from people who actually share a campus with you."
        primaryCta={{ label: 'Join with your Binus email', href: '/signup' }}
        secondaryCta={{ label: 'See how it works', href: '/about/threads' }}
      />

      <AboutSplitValue />

      {/* Why it exists */}
      <section className="about-content-section">
        <div className="about-container">
          <div className="about-content-split">
            <div>
              <span className="about-content-eyebrow">Why we built it</span>
              <h2 className="about-content-h2">
                Campus life deserved better than a group chat.
              </h2>
              <p className="about-content-lede">
                The questions students ask each other every day — where to study, who&apos;s
                joining the event, did the lab notes get posted — kept getting buried in
                group chats, public timelines, and platforms that don&apos;t know what a
                faculty is.
              </p>
              <p
                style={{
                  color: 'var(--text-muted)',
                  fontSize: '1rem',
                  lineHeight: 1.65,
                  marginTop: 18,
                  maxWidth: '36rem',
                }}
              >
                Sunibgram is an attempt to put all of that in one verified place. The
                people are real, the feed is on-topic, and the rules are written for
                campus — not for advertisers.
              </p>
            </div>
            <div>
              <AboutWhatDiagram />
            </div>
          </div>
        </div>
      </section>

      {/* Principles */}
      <section className="about-content-section">
        <div className="about-container">
          <div style={{ textAlign: 'center', maxWidth: '40rem', margin: '0 auto' }}>
            <span className="about-content-eyebrow" style={{ justifyContent: 'center' }}>
              What we stand for
            </span>
            <h2 className="about-content-h2" style={{ marginInline: 'auto' }}>
              Six things we won&apos;t compromise on.
            </h2>
            <p className="about-content-lede" style={{ marginInline: 'auto' }}>
              These are the trade-offs we make, even when other platforms would do the
              opposite.
            </p>
          </div>
          <AboutPrincipleCards />
        </div>
      </section>

      <AboutFinalCTA />
    </>
  );
}
