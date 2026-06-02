import Link from 'next/link';
import type { ReactNode } from 'react';
import { AboutHeroBackground } from '@/components/about/AboutHeroBackground';

type CTA = { label: string; href: string };

export function AboutPageHero({
  eyebrow,
  title,
  description,
  primaryCta,
  secondaryCta,
  secondaryCtaDark = false,
  align = 'center',
  meta,
}: {
  eyebrow: string;
  title: ReactNode;
  description?: ReactNode;
  primaryCta?: CTA;
  secondaryCta?: CTA;
  secondaryCtaDark?: boolean;
  align?: 'center' | 'left';
  meta?: ReactNode;
}) {
  const isLeft = align === 'left';
  return (
    <section className="relative overflow-hidden">
      <AboutHeroBackground />
      <div className="about-hero-grain" aria-hidden />
      <div
        className={`about-container-narrow relative z-10 flex flex-col${isLeft ? '' : ' items-center text-center'}`}
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
            marginTop: 20,
            maxWidth: '46rem',
          }}
        >
          {title}
        </h1>
        {description && (
          <p
            style={{
              color: 'var(--text)',
              fontSize: '1.0625rem',
              lineHeight: 1.65,
              marginTop: 24,
              maxWidth: '38rem',
            }}
          >
            {description}
          </p>
        )}
        {meta && (
          <div
            style={{
              marginTop: 24,
              fontSize: 13,
              color: 'var(--text-muted)',
              letterSpacing: '0.01em',
            }}
          >
            {meta}
          </div>
        )}
        {(primaryCta || secondaryCta) && (
          <div style={{ marginTop: 36, display: 'flex', gap: 12, flexWrap: 'wrap', justifyContent: isLeft ? 'flex-start' : 'center' }}>
            {primaryCta && (
              <Link
                href={primaryCta.href}
                className="about-btn-primary about-btn-primary-inverted"
                style={{ color: 'var(--text)', borderColor: 'rgba(0, 0, 0, 0.18)' }}
              >
                {primaryCta.label}
              </Link>
            )}
            {secondaryCta && (
              <Link
                href={secondaryCta.href}
                className={
                  secondaryCtaDark
                    ? 'about-btn-dark'
                    : 'about-btn-secondary about-btn-secondary-inverted'
                }
                style={secondaryCtaDark ? undefined : { color: 'var(--text)' }}
              >
                {secondaryCta.label}
              </Link>
            )}
          </div>
        )}
      </div>
    </section>
  );
}
