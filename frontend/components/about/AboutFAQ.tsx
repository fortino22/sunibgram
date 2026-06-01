'use client';

import { useState, type ReactNode } from 'react';
import { Plus } from 'lucide-react';

export type FAQItem = {
  q: string;
  a: ReactNode;
};

export function AboutFAQ({ items }: { items: FAQItem[] }) {
  const [openIdx, setOpenIdx] = useState<number | null>(0);

  return (
    <div className="about-faq-list">
      {items.map((item, i) => {
        const isOpen = openIdx === i;
        return (
          <div key={i} className="about-faq-item">
            <button
              className="about-faq-question"
              aria-expanded={isOpen}
              onClick={() => setOpenIdx(isOpen ? null : i)}
            >
              <span>{item.q}</span>
              <Plus size={18} className="about-faq-icon" data-open={isOpen} />
            </button>
            {isOpen && <div className="about-faq-answer">{item.a}</div>}
          </div>
        );
      })}
    </div>
  );
}
