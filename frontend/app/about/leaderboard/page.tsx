import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';

export const metadata: Metadata = {
  title: 'Leaderboard — Sunibgram',
  description:
    'Top contributors and rising voices on campus, ranked weekly by genuine engagement.',
};

export default function LeaderboardAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="Ranks · Contributors"
      title="Who is shaping the conversation."
      description="Discover the top contributors and rising voices on campus — ranked weekly by replies, reposts, and the engagement they spark."
      highlights={[
        {
          title: 'Weekly rankings',
          description:
            'A fresh leaderboard every week — the conversation never stays the same for long.',
        },
        {
          title: 'Engagement scores',
          description:
            'Ranks are based on real interaction — replies, likes, and reposts from verified peers.',
        },
        {
          title: 'Faculty leaderboards',
          description:
            'See who is leading the conversation inside your school, not just across the platform.',
        },
        {
          title: 'Your stats',
          description:
            'Track your own contribution journey week over week, with personal trend metrics.',
        },
      ]}
    />
  );
}
