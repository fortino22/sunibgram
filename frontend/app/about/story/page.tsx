import type { Metadata } from 'next';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';

export const metadata: Metadata = {
  title: 'Our story — Sunibgram',
  description:
    'Why Sunibgram exists — a campus-only social network built by Binus students, for Binus students.',
};

export default function StoryPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="Our story"
        title="Built by students, for the campus they actually live on."
        description="Sunibgram started as a simple idea: a place where every voice on your feed is a real person from your campus. No bots, no strangers, no anonymous noise — just Binus."
        primaryCta={{ label: 'Join with your Binus email', href: '/signup' }}
        secondaryCta={{ label: 'See what we are', href: '/about/company' }}
        secondaryCtaDark
      />

      <section className="about-content-section">
        <div className="about-container">
          <div className="about-content-split" style={{ maxWidth: 'none' }}>
            <div>
              <span className="about-content-eyebrow">Why we built it</span>
              <h2 className="about-content-h2">
                The campus deserved its own space.
              </h2>
              <p className="about-content-lede">
                Mainstream social apps treat every user the same, whether they
                share a lecture hall or live on opposite sides of the planet.
                That works for the open internet — it doesn&apos;t work for a
                campus. Sunibgram is a smaller, sharper feed where the people
                you see are people you can actually meet.
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
                Everyone here signs in with a verified{' '}
                <span style={{ color: 'var(--text)', fontWeight: 600 }}>
                  @binus.ac.id
                </span>{' '}
                email. One identity, no second accounts, no impersonation.
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
                  WHO WE ARE
                </div>
                <div
                  style={{
                    marginTop: 8,
                    color: 'var(--text)',
                    fontSize: 15,
                    lineHeight: 1.55,
                  }}
                >
                  A team of Binus students building the social app we wished we
                  had on day one.
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
                  WHO IT&apos;S FOR
                </div>
                <div
                  style={{
                    marginTop: 8,
                    color: 'var(--text)',
                    fontSize: 15,
                    lineHeight: 1.55,
                  }}
                >
                  Anyone with a verified Binus email — across faculties,
                  campuses, and years.
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
                  WHERE WE BUILD
                </div>
                <div
                  style={{
                    marginTop: 8,
                    color: 'var(--text)',
                    fontSize: 15,
                    lineHeight: 1.55,
                  }}
                >
                  Made in Jakarta.
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="about-content-section">
        <div className="about-container">
          <span className="about-content-eyebrow">What we believe</span>
          <h2 className="about-content-h2" style={{ maxWidth: '36rem' }}>
            A few things we keep coming back to.
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
                title: 'Real people, real campus',
                body: 'Verification first. We’d rather ship slower than open the door to anonymous accounts.',
              },
              {
                title: 'Small feed, big signal',
                body: 'A campus-sized feed beats an infinite one. You see fewer posts, but they actually matter to you.',
              },
              {
                title: 'Independent on purpose',
                body: 'We’re not an official Binus product. That keeps decisions about the app close to the students using it.',
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
