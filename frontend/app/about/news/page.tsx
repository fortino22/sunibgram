import type { Metadata } from 'next';
import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { NEWS_ITEMS, formatNewsDate } from '@/lib/news';

export const metadata: Metadata = {
  title: 'News — Sunibgram',
  description:
    'Announcements, release notes, and updates from the Sunibgram team.',
};

export default function NewsPage() {
  const items = [...NEWS_ITEMS].sort((a, b) => b.date.localeCompare(a.date));

  return (
    <>
      <AboutPageHero
        eyebrow="News"
        title="What's new at Sunibgram."
        description="Announcements, release notes, and the occasional behind-the-scenes from the team."
      />

      <section className="about-content-section">
        <div className="about-container">
          <ul
            style={{
              listStyle: 'none',
              padding: 0,
              margin: '0 auto',
              maxWidth: '52rem',
              display: 'flex',
              flexDirection: 'column',
              gap: 0,
            }}
          >
            {items.map((item, i) => (
              <li
                key={item.slug}
                style={{
                  borderTop:
                    i === 0 ? '1px solid var(--border)' : undefined,
                  borderBottom: '1px solid var(--border)',
                }}
              >
                <Link
                  href={`/about/news/${item.slug}`}
                  className="about-news-row"
                >
                  <div className="about-news-meta">
                    <span className="about-news-date">
                      {formatNewsDate(item.date)}
                    </span>
                    <span className="about-news-cat">{item.category}</span>
                  </div>
                  <div className="about-news-body">
                    <h2 className="about-news-title">{item.title}</h2>
                    <p className="about-news-lede">{item.lede}</p>
                  </div>
                  <span className="about-news-arrow" aria-hidden>
                    <ArrowRight size={18} />
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </section>
    </>
  );
}
