import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';
import {
  LeaderboardDemoWeekly,
  LeaderboardDemoScores,
  LeaderboardDemoFaculty,
  LeaderboardDemoYourStats,
} from '@/components/about/demos/LeaderboardDemos';

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
      sections={[
        {
          id: 'weekly-rankings',
          title: 'Weekly rankings',
          description:
            'A fresh leaderboard every week. The conversation never stays the same for long, so the leaderboard does not either.',
          bullets: [
            'Resets Sunday at 00:00 local time',
            'Live re-sort as new engagement comes in',
            'Move-up and move-down arrows track shifts',
          ],
          demo: <LeaderboardDemoWeekly />,
        },
        {
          id: 'engagement-scores',
          title: 'Engagement scores',
          description:
            'Ranks are based on real interaction from verified peers — replies, reposts, and likes — not just raw post volume.',
          bullets: [
            'Replies weighted heavier than passive likes',
            'Bot accounts cannot influence scores',
            'Score breakdown is visible from your profile',
          ],
          demo: <LeaderboardDemoScores />,
        },
        {
          id: 'faculty-boards',
          title: 'Faculty leaderboards',
          description:
            'See who is leading the conversation inside your school, not just across the platform. Easier to spot your own classmates rising.',
          bullets: [
            'One leaderboard per faculty',
            'Switch faculties without leaving the page',
            'Highlights friends you already follow',
          ],
          demo: <LeaderboardDemoFaculty />,
        },
        {
          id: 'your-stats',
          title: 'Your stats',
          description:
            'Track your contribution journey week over week. Personal trend metrics make it clear when you are climbing — or when it is time to post again.',
          bullets: [
            'Week-over-week score trend',
            'Drill into replies, reposts, and likes',
            'Streaks and milestones unlock as you grow',
          ],
          demo: <LeaderboardDemoYourStats />,
        },
      ]}
    />
  );
}
