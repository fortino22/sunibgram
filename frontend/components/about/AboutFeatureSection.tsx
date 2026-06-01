import type { FeatureSection } from '@/components/about/AboutFeaturePage';

export function AboutFeatureSection({
  index,
  section,
}: {
  index: number;
  section: FeatureSection;
}) {
  return (
    <section
      id={section.id}
      className="about-container about-feature-section"
    >
      <div className="about-feature-section-grid">
        <div className="about-feature-section-copy">
          <span className="about-feature-section-num">
            {String(index + 1).padStart(2, '0')}
          </span>
          <h2 className="about-feature-section-title">{section.title}</h2>
          <p className="about-feature-section-desc">{section.description}</p>
          {section.bullets && section.bullets.length > 0 && (
            <ul className="about-feature-section-bullets">
              {section.bullets.map((b) => (
                <li key={b}>{b}</li>
              ))}
            </ul>
          )}
        </div>
        <div className="about-feature-section-demo">{section.demo}</div>
      </div>
    </section>
  );
}
