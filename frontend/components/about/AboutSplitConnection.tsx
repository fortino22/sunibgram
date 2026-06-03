import Link from 'next/link';
import { AboutGlobe } from './AboutGlobe';

export function AboutSplitConnection() {
  return (
    <section
      className="about-container"
      style={{
        paddingTop: '32px',
        paddingBottom: '32px',
        overflow: 'hidden',
      }}
    >
      <div
        className="grid gap-6 items-center"
        style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))' }}
      >
        {/* Left — copy */}
        <div style={{ maxWidth: '34rem', justifySelf: 'start' }}>
          <span className="about-eyebrow">Verified · one account</span>
          <h2
            style={{
              color: 'var(--text)',
              fontWeight: 500,
              fontSize: 'clamp(1.75rem, 3.5vw, 2.75rem)',
              lineHeight: 1.12,
              letterSpacing: '-0.02em',
              marginTop: '20px',
            }}
          >
            One sign-in.
            <br />
            Everything Binus.
          </h2>
          <p
            style={{
              color: 'var(--text-muted)',
              fontSize: '1rem',
              lineHeight: 1.65,
              marginTop: '20px',
              maxWidth: '32rem',
            }}
          >
            Sign in once with your <span style={{ color: 'var(--text)', fontWeight: 600 }}>@binus.ac.id</span>.
            Your faculty, your profile, your DMs, your communities — they all
            know it&apos;s really you. No second account, no second password.
          </p>
          <div style={{ marginTop: '28px' }}>
            <Link href="/signup" className="about-btn-dark">
              Sign up for free →
            </Link>
          </div>
        </div>

        {/* Right — Three.js 3D globe of Binus people */}
        <div className="about-orbit-stage" style={{ justifySelf: 'end' }}>
          <AboutGlobe />
        </div>
      </div>
    </section>
  );
}
