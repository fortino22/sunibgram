export type QuickNavItem = {
  id: string;
  title: string;
  description: string;
};

export function AboutFeatureQuickNav({ sections }: { sections: QuickNavItem[] }) {
  return (
    <section
      className="about-feature-quicknav-section"
      style={{ paddingTop: '56px', paddingBottom: '56px' }}
    >
      <div className="about-container">
        <div className="about-feature-quicknav">
          <div className="about-feature-quicknav-head">
            <span className="about-eyebrow">Quick navigate</span>
            <h2 className="about-feature-quicknav-title">
              What&apos;s in this page.
            </h2>
          </div>
          <ol className="about-feature-quicknav-grid">
            {sections.map((s, i) => (
              <li key={s.id} className="about-feature-quicknav-item">
                <a
                  href={`#${s.id}`}
                  className="about-feature-quicknav-link"
                >
                  <span className="about-feature-quicknav-num">
                    {String(i + 1).padStart(2, '0')}
                  </span>
                  <span className="about-feature-quicknav-text">
                    <span className="about-feature-quicknav-name">{s.title}</span>
                    <span className="about-feature-quicknav-desc">{s.description}</span>
                  </span>
                  <span className="about-feature-quicknav-arrow" aria-hidden>↓</span>
                </a>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </section>
  );
}
