import Link from 'next/link';
import { AboutHeroBackground } from '@/components/about/AboutHeroBackground';

export function AboutHero() {
  return (
    <section className="relative overflow-hidden" id="what">
      <AboutHeroBackground />
      <div className="about-hero-grain" aria-hidden />
      <div
        className="about-container-narrow relative z-10 flex flex-col items-center text-center"
        style={{ paddingTop: '40px', paddingBottom: '40px' }}
      >
        <h1
          style={{
            color: 'var(--text)',
            fontWeight: 500,
            lineHeight: 1.08,
            fontSize: 'clamp(2.5rem, 6vw, 5rem)',
            letterSpacing: '-0.025em',
            maxWidth: '52rem',
          }}
        >
          The{' '}
          <span className="about-inline-chip">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src="/assets/binus-logo.svg"
              alt="Binus University"
              style={{
                height: '1.12em',
                width: 'auto',
                objectFit: 'contain',
                display: 'block',
              }}
            />
          </span>{' '}
          social network.
        </h1>

        <p
          style={{
            color: 'var(--text)',
            fontSize: '1.0625rem',
            lineHeight: 1.6,
            maxWidth: '38rem',
            marginTop: '28px',
          }}
        >
          The campus app for Binus students. Threads, communities, and events —
          one verified identity, no anonymous noise.
        </p>

        <div className="flex items-center justify-center" style={{ marginTop: '36px' }}>
          <Link
            href="/signup"
            className="about-btn-primary about-btn-primary-inverted"
            style={{ color: 'var(--text)', borderColor: 'rgba(0, 0, 0, 0.18)' }}
          >
            Join with your Binus email
          </Link>
        </div>
      </div>
    </section>
  );
}
