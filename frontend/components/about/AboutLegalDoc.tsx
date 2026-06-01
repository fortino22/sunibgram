import type { ReactNode } from 'react';

export type LegalSection = {
  id: string;
  title: string;
  body: ReactNode;
};

export function AboutLegalDoc({ sections }: { sections: LegalSection[] }) {
  return (
    <div className="about-legal-layout">
      <nav className="about-legal-toc" aria-label="On this page">
        <span className="about-legal-toc-heading">On this page</span>
        {sections.map((s, i) => (
          <a key={s.id} href={`#${s.id}`}>
            {String(i + 1).padStart(2, '0')} · {s.title}
          </a>
        ))}
      </nav>
      <div className="about-legal-body">
        {sections.map((s, i) => (
          <section key={s.id} id={s.id} className="about-legal-section">
            <div className="about-legal-section-num">{String(i + 1).padStart(2, '0')}</div>
            <h2 className="about-legal-section-title">{s.title}</h2>
            <div className="about-legal-section-body">{s.body}</div>
          </section>
        ))}
      </div>
    </div>
  );
}
