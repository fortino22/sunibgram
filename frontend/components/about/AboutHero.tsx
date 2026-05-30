import Link from 'next/link';

export function AboutHero() {
  return (
    <section className="about-hero-mesh" id="what">
      <div aria-hidden className="about-hero-grid" />
      <div
        className="about-container-narrow relative flex flex-col items-center text-center"
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
            <span className="about-inline-chip-avatar" />
            <span
              style={{
                color: 'var(--text)',
                fontSize: '1.05rem',
                fontWeight: 600,
              }}
            >
              @binus.ac.id
            </span>
            <span className="about-inline-chip-badge">VERIFIED</span>
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
