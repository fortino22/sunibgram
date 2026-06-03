import Link from 'next/link';
import { AboutHeroBackground } from '@/components/about/AboutHeroBackground';

export function AboutFinalCTA() {
  return (
    <section id="team" className="relative overflow-hidden">
      <AboutHeroBackground />
      <div className="about-hero-grain" aria-hidden />
      <div
        className="about-container relative z-10 flex flex-wrap items-center justify-between gap-8"
        style={{ paddingTop: '72px', paddingBottom: '72px' }}
      >
        <div style={{ maxWidth: '40rem' }}>
          <h2
            style={{
              color: 'var(--text)',
              fontWeight: 500,
              fontSize: 'clamp(1.75rem, 3.5vw, 2.75rem)',
              lineHeight: 1.12,
              letterSpacing: '-0.02em',
            }}
          >
            Made for Binus. Built by Binusian.
          </h2>
          <p
            style={{
              color: 'var(--text-muted)',
              fontSize: '1rem',
              lineHeight: 1.6,
              marginTop: '14px',
              maxWidth: '34rem',
            }}
          >
            Sunibgram is built by students at Binus, for students at Binus.
            The only feed where everyone you see is actually on your campus.
          </p>
        </div>

        <div className="flex items-center gap-3 flex-wrap">
          <Link href="/feedback" className="about-btn-secondary">
            Talk to the team
          </Link>
        </div>
      </div>
    </section>
  );
}
