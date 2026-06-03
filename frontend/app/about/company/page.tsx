import type { Metadata } from 'next';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutSplitValue } from '@/components/about/AboutSplitValue';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';

export const metadata: Metadata = {
  title: 'Company — Sunibgram',
  description:
    'Sunibgram is the verified campus social network for Binus students — built independently by Binusian in Jakarta. Learn what it is and who builds it.',
};

export default function CompanyPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="Company"
        title="A social network built around one campus."
        description="Sunibgram is a verified-only platform for Binus students. Threads, communities, events, and DMs — all from people who actually share a campus with you."
        primaryCta={{ label: 'Join with your Binus email', href: '/signup' }}
        secondaryCta={{ label: 'See how it works', href: '/about/threads' }}
        secondaryCtaDark
      />

      <AboutSplitValue />

      <AboutFinalCTA />

      {/* Independent, by design */}
      <section className="about-content-section">
        <div className="about-container">
          <div className="about-content-split" style={{ maxWidth: 'none' }}>
            <div>
              <span className="about-content-eyebrow">Independent, by design</span>
              <h2 className="about-content-h2">
                A student project, not an official Binus product.
              </h2>
              <p className="about-content-lede">
                Sunibgram was started by Binus students and is built outside the
                university&apos;s official channels. That independence matters: we make
                product decisions based on what students actually want, not what looks
                good in a corporate roadmap.
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
                Anyone with a verified <span style={{ color: 'var(--text)', fontWeight: 600 }}>@binus.ac.id</span> email
                is welcome — and the rest of the internet, politely, is not.
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
                <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--text-tertiary)', letterSpacing: '0.08em' }}>
                  WHAT WE ARE
                </div>
                <div style={{ marginTop: 8, color: 'var(--text)', fontSize: 15, lineHeight: 1.55 }}>
                  A campus-only social network, designed and built by Binus students.
                </div>
              </div>
              <div style={{ height: 1, background: 'var(--border)' }} />
              <div>
                <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--text-tertiary)', letterSpacing: '0.08em' }}>
                  WHAT WE ARE NOT
                </div>
                <div style={{ marginTop: 8, color: 'var(--text)', fontSize: 15, lineHeight: 1.55 }}>
                  An official Binus University product, sponsor, or partner. We don&apos;t speak for the university.
                </div>
              </div>
              <div style={{ height: 1, background: 'var(--border)' }} />
              <div>
                <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--accent)', letterSpacing: '0.08em' }}>
                  WHERE WE BUILD
                </div>
                <div style={{ marginTop: 8, color: 'var(--text)', fontSize: 15, lineHeight: 1.55 }}>
                  Made in Jakarta.
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
