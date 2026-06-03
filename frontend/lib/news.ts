export type NewsItem = {
  slug: string;
  category: string;
  title: string;
  lede: string;
  /** ISO date — used for sorting and rendering. */
  date: string;
};

export const NEWS_ITEMS: NewsItem[] = [
  {
    slug: 'sunibgram-enters-beta',
    category: 'Announcement',
    title: 'Sunibgram enters open beta for Binusian.',
    lede:
      'The verified campus social network is now open to anyone with a @binus.ac.id email. Here is what is live, what is still cooking, and how to join.',
    date: '2026-06-03',
  },
];

export function getNewsBySlug(slug: string): NewsItem | undefined {
  return NEWS_ITEMS.find((item) => item.slug === slug);
}

export function formatNewsDate(iso: string): string {
  const d = new Date(`${iso}T00:00:00Z`);
  return d.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    timeZone: 'UTC',
  });
}
