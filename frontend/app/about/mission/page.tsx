import type { Metadata } from 'next';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';

export const metadata: Metadata = {
  title: 'Mission — Sunibgram',
  description:
    'Our mission — give every campus a social space where every voice on the feed is a real, verified student.',
};

export default function MissionPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="Mission"
        title="A feed where everyone is real, and everyone is here."
        description="Our mission is simple: give the campus a social space built on verified identity, not anonymous noise. One Binus email, one real person, one feed that actually belongs to the people on it."
        primaryCta={{ label: 'Join with your Binus email', href: '/signup' }}
        secondaryCta={{ label: 'Read our story', href: '/about/story' }}
        secondaryCtaDark
      />

      <section className="about-content-section">
        <div className="about-container">
          <span className="about-content-eyebrow">What drives us</span>
          <h2 className="about-content-h2" style={{ maxWidth: '36rem' }}>
            Closer connections, less noise.
          </h2>
          <p className="about-content-lede" style={{ maxWidth: '40rem' }}>
            The open internet rewards reach. A campus rewards proximity — the
            people you can meet between lectures, in the same clubs, on the same
            grounds. Sunibgram is built for that smaller, sharper feed: fewer
            posts, but every one from someone you could actually run into.
          </p>
          <p
            style={{
              color: 'var(--text-muted)',
              fontSize: '1rem',
              lineHeight: 1.65,
              marginTop: 18,
              maxWidth: '40rem',
            }}
          >
            Verification comes first. Everyone signs in with a verified{' '}
            <span style={{ color: 'var(--text)', fontWeight: 600 }}>
              @binus.ac.id
            </span>{' '}
            email — no bots, no impersonation, no second accounts. That single
            constraint is what makes the rest of the platform worth trusting.
          </p>
        </div>
      </section>

      <AboutFinalCTA />
    </>
  );
}
