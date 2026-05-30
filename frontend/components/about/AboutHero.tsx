import Link from 'next/link';
import { AboutHeroBackground } from '@/components/about/AboutHeroBackground';

export function AboutHero() {
  return (
    <section className="relative overflow-hidden" id="what">
      <AboutHeroBackground />
      <div
        className="about-container-narrow relative z-10 flex flex-col items-center text-center"
        style={{ paddingTop: '120px', paddingBottom: '120px' }}
      >
        <h1
          style={{
            color: 'var(--text)',
            fontWeight: 600,
            lineHeight: 1.08,
            fontSize: 'clamp(2.5rem, 6vw, 5rem)',
            letterSpacing: '-0.025em',
            maxWidth: '52rem',
          }}
        >
          The social network just for{' '}
          <span className="about-inline-chip">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src="/assets/binus-logo.svg"
              alt="Binus University"
              style={{
                height: '1.4em',
                width: 'auto',
                objectFit: 'contain',
                display: 'block',
              }}
            />
          </span>
        </h1>

        <p
          style={{
            color: 'var(--text-muted)',
            fontSize: '1.0625rem',
            lineHeight: 1.6,
            maxWidth: '38rem',
            marginTop: '28px',
          }}
        >
          Sunibgram is the campus app for Binus University students.
          Threads, communities, events, and DMs — one place, one verified
          identity, zero anonymous noise.
        </p>

        <div className="flex flex-wrap items-center justify-center gap-3" style={{ marginTop: '36px' }}>
          <Link href="/signup" className="about-btn-primary">
            Sign up with your Binus email
          </Link>
          <Link href="/login" className="about-btn-secondary">
            I already have an account
          </Link>
        </div>
      </div>
    </section>
  );
}
