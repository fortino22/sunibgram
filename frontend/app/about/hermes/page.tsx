import type { Metadata } from 'next';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';

export const metadata: Metadata = {
  title: 'How we operate — Sunibgram',
  description:
    'Hermes is the operational toolset that keeps Sunibgram running — automating reviews, surfacing intelligence, and powering on-platform AI like Fox.',
};

export default function HermesPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="Hermes"
        title="How we operate, behind the scenes."
        description="Hermes is the operational toolset that keeps Sunibgram running day to day. It’s the layer that automates the routine work, surfaces what needs a human, and powers on-platform AI like Fox."
        primaryCta={{ label: 'Meet Fox', href: '/about/fox' }}
        secondaryCta={{ label: 'See our story', href: '/about/story' }}
        secondaryCtaDark
      />

      {/* What Hermes is */}
      <section className="about-content-section">
        <div className="about-container">
          <div className="about-content-split" style={{ maxWidth: 'none' }}>
            <div>
              <span className="about-content-eyebrow">What it is</span>
              <h2 className="about-content-h2">
                The toolset, not the team.
              </h2>
              <p className="about-content-lede">
                Hermes is the operational layer of Sunibgram — the set of
                tools we use to review events, monitor activity, and turn
                live data into something a human can act on. It runs quietly
                in the background so leadership doesn’t have to live in
                dashboards.
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
                Fox — our on-platform AI analyst — is one of the agents that
                runs on top of Hermes. When you see Fox approve an event or
                deliver a morning brief, Hermes is the toolchain that made
                it possible.
              </p>
            </div>

            <div
              style={{
                border: '1px solid var(--border)',
                borderRadius: 16,
                background: 'var(--bg-card)',
                padding: '24px 28px',
                display: 'flex',
                flexDirection: 'column',
                gap: 18,
              }}
            >
              <div>
                <div
                  style={{
                    fontSize: 12,
                    fontWeight: 700,
                    color: 'var(--text-tertiary)',
                    letterSpacing: '0.08em',
                  }}
                >
                  WHAT HERMES IS
                </div>
                <div
                  style={{
                    marginTop: 8,
                    color: 'var(--text)',
                    fontSize: 15,
                    lineHeight: 1.55,
                  }}
                >
                  The operational toolset that powers Sunibgram’s automation
                  and intelligence layer.
                </div>
              </div>
              <div style={{ height: 1, background: 'var(--border)' }} />
              <div>
                <div
                  style={{
                    fontSize: 12,
                    fontWeight: 700,
                    color: 'var(--text-tertiary)',
                    letterSpacing: '0.08em',
                  }}
                >
                  WHAT HERMES IS NOT
                </div>
                <div
                  style={{
                    marginTop: 8,
                    color: 'var(--text)',
                    fontSize: 15,
                    lineHeight: 1.55,
                  }}
                >
                  An AI of its own, a moderator, or a replacement for human
                  judgment.
                </div>
              </div>
              <div style={{ height: 1, background: 'var(--border)' }} />
              <div>
                <div
                  style={{
                    fontSize: 12,
                    fontWeight: 700,
                    color: 'var(--accent)',
                    letterSpacing: '0.08em',
                  }}
                >
                  WHO RUNS ON IT
                </div>
                <div
                  style={{
                    marginTop: 8,
                    color: 'var(--text)',
                    fontSize: 15,
                    lineHeight: 1.55,
                  }}
                >
                  Fox — the platform’s AI analyst. See what Fox does on{' '}
                  <a
                    href="/about/fox"
                    style={{ color: 'var(--text)', fontWeight: 600 }}
                  >
                    the Fox page
                  </a>
                  .
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* How it shows up */}
      <section className="about-content-section">
        <div className="about-container">
          <span className="about-content-eyebrow">How it shows up</span>
          <h2 className="about-content-h2" style={{ maxWidth: '40rem' }}>
            Where you’ll actually feel Hermes at work.
          </h2>

          <div
            className="grid"
            style={{
              gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))',
              gap: 20,
              marginTop: 32,
            }}
          >
            {[
              {
                title: 'Event review',
                body: 'Submit an official event and Hermes routes it to Fox for a policy check. Clear-cut events move quickly; risky ones are flagged for a human to confirm.',
              },
              {
                title: 'Live reporting',
                body: 'Raw platform activity is pulled and shaped into the daily brief and weekly story Fox delivers — every number tied back to live data.',
              },
              {
                title: 'Human-in-the-loop',
                body: 'Hermes never quietly takes a destructive action. Anything sensitive — a decline, a removal, a policy edge case — surfaces to a person.',
              },
            ].map((card) => (
              <div
                key={card.title}
                style={{
                  border: '1px solid var(--border)',
                  borderRadius: 16,
                  background: 'var(--bg-card)',
                  padding: '22px 24px',
                }}
              >
                <div
                  style={{
                    color: 'var(--text)',
                    fontWeight: 700,
                    fontSize: 16,
                    letterSpacing: '-0.01em',
                  }}
                >
                  {card.title}
                </div>
                <p
                  style={{
                    color: 'var(--text-muted)',
                    fontSize: 14.5,
                    lineHeight: 1.6,
                    marginTop: 10,
                  }}
                >
                  {card.body}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <AboutFinalCTA />
    </>
  );
}
