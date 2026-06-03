import Link from 'next/link';
import type { ReactNode } from 'react';
import { AboutHeroBackground } from '@/components/about/AboutHeroBackground';
import { AboutFeatureQuickNav } from '@/components/about/AboutFeatureQuickNav';
import { AboutFeatureSection } from '@/components/about/AboutFeatureSection';

export type FeatureSection = {
  id: string;
  title: string;
  description: string;
  bullets?: string[];
  demo: ReactNode;
};

export function AboutFeaturePage({
  eyebrow,
  title,
  description,
  sections,
  ctaLabel = 'Join with your Binus email',
  ctaHref = '/signup',
  badge,
}: {
  eyebrow: string;
  title: ReactNode;
  description: string;
  sections: FeatureSection[];
  ctaLabel?: string;
  ctaHref?: string;
  badge?: string;
}) {
  return (
    <>
      {/* Section 1 — Hero */}
      <section className="relative overflow-hidden">
        <AboutHeroBackground />
        <div className="about-hero-grain" aria-hidden />
        <div
          className="about-container-narrow relative z-10 flex flex-col items-center text-center"
          style={{ paddingTop: '72px', paddingBottom: '72px' }}
        >
          {badge && <span className="about-hero-dev-badge">{badge}</span>}
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

      {/* Section 2 — Quick navigate */}
      <AboutFeatureQuickNav sections={sections} />

      {/* Section 3+ — Explanation sections (left text · right demo) */}
      {sections.map((section, i) => (
        <AboutFeatureSection
          key={section.id}
          index={i}
          section={section}
        />
      ))}
    </>
  );
}
