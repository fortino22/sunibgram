import Link from 'next/link';

export function AboutFinalCTA() {
  return (
    <section
      id="team"
      style={{
        background:
          'linear-gradient(135deg, var(--accent) 0%, oklch(56% 0.18 240) 100%)',
      }}
    >
      <div
        className="about-container flex flex-wrap items-center justify-between gap-8"
        style={{ paddingTop: '72px', paddingBottom: '72px' }}
      >
        <div style={{ maxWidth: '40rem' }}>
          <h2
            style={{
              color: '#ffffff',
              fontSize: 'clamp(1.875rem, 3.5vw, 2.5rem)',
              fontWeight: 600,
              letterSpacing: '-0.02em',
              lineHeight: 1.15,
            }}
          >
            Made for Binus. Built by Binus.
          </h2>
          <p
            style={{
              color: 'rgba(255,255,255,0.85)',
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
          <Link href="/feedback" className="about-btn-secondary about-btn-secondary-inverted">
            Talk to the team
          </Link>
          <Link href="/signup" className="about-btn-primary about-btn-primary-inverted">
            Join with your Binus email
          </Link>
        </div>
      </div>
    </section>
  );
}
