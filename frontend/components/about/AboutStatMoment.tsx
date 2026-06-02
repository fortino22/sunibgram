'use client';

import Link from 'next/link';
import { useUserCount } from '@/hooks/useUserCount';

export function AboutStatMoment() {
  const { data } = useUserCount();
  const display =
    typeof data === 'number' ? new Intl.NumberFormat('id-ID').format(data) : null;
  return (
    <section
      className="about-container"
      style={{
        paddingTop: '96px',
        paddingBottom: '96px',
      }}
    >
      <div
        className="grid gap-16 items-center"
        style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))' }}
      >
        {/* Left — copy */}
        <div style={{ maxWidth: '34rem' }}>
          <span className="about-eyebrow">A feed that knows you</span>
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
            A feed that knows your campus.
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
            Sunibgram blends fresh posts with what&apos;s actually trending around
            you — your faculty, your communities, your batch. The Leaderboard
            rotates who&apos;s shaping the conversation right now. Show up, get seen.
          </p>
          <div style={{ marginTop: '28px' }}>
            <Link href="/signup" className="about-link-mono">
              Get on the feed →
            </Link>
          </div>
        </div>

        {/* Right — big stat */}
        <div
          style={{
            justifySelf: 'end',
            textAlign: 'right',
            maxWidth: '36rem',
            width: '100%',
          }}
        >
          <div
            style={{
              fontSize: '0.75rem',
              textTransform: 'uppercase',
              letterSpacing: '0.1em',
              color: 'var(--accent)',
              marginBottom: '20px',
              fontWeight: 600,
            }}
          >
            Students already on Sunibgram
          </div>
          <div
            style={{
              fontSize: 'clamp(3.5rem, 8vw, 6rem)',
              fontWeight: 500,
              color: 'var(--text)',
              letterSpacing: '-0.03em',
              lineHeight: 1,
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            {display === null ? (
              <span
                aria-label="Loading data"
                style={{
                  display: 'inline-block',
                  fontSize: 'clamp(1.5rem, 3vw, 2.25rem)',
                  fontWeight: 600,
                  letterSpacing: '0.01em',
                  background:
                    'linear-gradient(90deg, var(--text-muted) 25%, var(--text) 50%, var(--text-muted) 75%)',
                  backgroundSize: '200% 100%',
                  WebkitBackgroundClip: 'text',
                  backgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  animation:
                    'statShimmer 1.6s linear infinite, statWobble 2.2s ease-in-out infinite',
                }}
              >
                Loading data…
              </span>
            ) : (
              <>
                {display}
                <span style={{ color: 'var(--text-muted)' }}>+</span>
              </>
            )}
          </div>
          <style>{`
            @keyframes statShimmer { 0% { background-position: 200% 0 } 100% { background-position: -200% 0 } }
            @keyframes statWobble { 0%,100% { transform: rotate(-1.5deg) translateY(0) } 50% { transform: rotate(1.5deg) translateY(-2px) } }
          `}</style>
          <p
            style={{
              marginTop: '20px',
              color: 'var(--text-muted)',
              fontSize: '0.9375rem',
            }}
          >
            Building Sunibgram with us — post by post.
          </p>
        </div>
      </div>
    </section>
  );
}
