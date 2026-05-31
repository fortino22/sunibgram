import Link from 'next/link';
import type { ReactNode } from 'react';
import { AboutHeroBackground } from '@/components/about/AboutHeroBackground';

export type FeatureHighlight = {
  title: string;
  description: string;
};

export function AboutFeaturePage({
  eyebrow,
  title,
  description,
  highlights,
  ctaLabel = 'Join with your Binus email',
  ctaHref = '/signup',
}: {
  eyebrow: string;
  title: ReactNode;
  description: string;
  highlights: FeatureHighlight[];
  ctaLabel?: string;
  ctaHref?: string;
}) {
  return (
    <>
      <section className="relative overflow-hidden">
        <AboutHeroBackground />
        <div className="about-hero-grain" aria-hidden />
        <div
          className="about-container-narrow relative z-10 flex flex-col items-center text-center"
          style={{ paddingTop: '72px', paddingBottom: '72px' }}
        >
          <span className="about-eyebrow">{eyebrow}</span>
          <h1
            style={{
              color: 'var(--text)',
              fontWeight: 500,
              fontSize: 'clamp(2.25rem, 5vw, 3.75rem)',
              lineHeight: 1.1,
              letterSpacing: '-0.025em',
              marginTop: '20px',
              maxWidth: '46rem',
            }}
          >
            {title}
          </h1>
          <p
            style={{
              color: 'var(--text)',
              fontSize: '1.0625rem',
              lineHeight: 1.65,
              marginTop: '24px',
              maxWidth: '38rem',
            }}
          >
            {description}
          </p>
          <div style={{ marginTop: '36px' }}>
            <Link
              href={ctaHref}
              className="about-btn-primary about-btn-primary-inverted"
              style={{ color: 'var(--text)', borderColor: 'rgba(0, 0, 0, 0.18)' }}
            >
              {ctaLabel}
            </Link>
          </div>
        </div>
      </section>

      <section
        className="about-container"
        style={{ paddingTop: '64px', paddingBottom: '96px' }}
      >
        <div
          className="grid gap-8"
          style={{
            gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
            maxWidth: '64rem',
            marginLeft: 'auto',
            marginRight: 'auto',
          }}
        >
          {highlights.map((h) => (
            <div
              key={h.title}
              style={{
                borderTop: '1px solid var(--border)',
                paddingTop: '20px',
              }}
            >
              <h3
                style={{
                  color: 'var(--text)',
                  fontSize: '1.0625rem',
                  fontWeight: 700,
                  letterSpacing: '-0.01em',
                }}
              >
                {h.title}
              </h3>
              <p
                style={{
                  color: 'var(--text-muted)',
                  fontSize: '0.9375rem',
                  lineHeight: 1.6,
                  marginTop: '8px',
                }}
              >
                {h.description}
              </p>
            </div>
          ))}
        </div>
      </section>
    </>
  );
}
