import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';
import {
  FoxDemoEventReview,
  FoxDemoDailyReport,
  FoxDemoWeeklyReport,
  FoxDemoHumanInTheLoop,
} from '@/components/about/demos/FoxDemos';

export const metadata: Metadata = {
  title: 'Fox AI — Sunibgram',
  description:
    'Fox is Sunibgram’s on-platform AI analyst — it reviews official events and turns live activity into a daily brief and a weekly story.',
};

export default function FoxAboutPage() {
  return (
    <AboutFeaturePage
      badge="Developer"
      eyebrow="Intelligence · Always on"
      title="Meet Fox — the AI that helps run Sunibgram."
      description="Fox reads what’s happening on the platform and gives leadership a clear answer instead of a dashboard. It reviews every official event, writes a morning brief, and tells the story of the week — every number tied back to live data."
      sections={[
        {
          id: 'fox-event-review',
          title: 'Reviews every official event',
          description:
            'When an organizer submits an event, Fox reads the proposal, checks it against Sunibgram’s event policy, and returns a clear verdict with a confidence score and a written reason.',
          bullets: [
            'Three calls — approve, decline, or ask for changes',
            'Low-risk, clear-cut events can be auto-approved',
            'Anything risky is flagged for a human to confirm',
          ],
          demo: <FoxDemoEventReview />,
        },
        {
          id: 'fox-daily-report',
          title: 'Writes the morning brief',
          description:
            'Every morning, Fox sends a two-minute briefing on the last 24 hours: which events were decided, which bugs need attention, and what users are asking for most.',
          bullets: [
            'At-a-glance summary with the day’s headlines',
            'Stat cards for events decided, auto-approvals, bugs, and feature requests',
            'Recommended actions — a short to-do list of the best moves',
          ],
          demo: <FoxDemoDailyReport />,
        },
        {
          id: 'fox-weekly-report',
          title: 'Tells the story of the week',
          description:
            'On Saturday morning, Fox delivers the big picture: trending hashtags, safety patterns, the community leaderboard, and an honest note on anything the data can’t fully cover.',
          bullets: [
            'Trending hashtags ranked with an insight on why they’re rising',
            'Top 10 most-engaging users — likes, comments, and reposts they earned',
            'Honest notes on data the report can’t fully cover that week',
          ],
          demo: <FoxDemoWeeklyReport />,
        },
        {
          id: 'fox-human-in-the-loop',
          title: 'Asks for a human when it should',
          description:
            'Fox never quietly auto-approves a risky event. When a proposal is borderline or low-confidence, it stops and routes the decision to an admin with its notes attached.',
          bullets: [
            'Low-confidence verdicts never auto-approve',
            'Borderline events come with Fox’s reasoning so the admin starts informed',
            'Every admin override loops back into how Fox judges the next one',
          ],
          demo: <FoxDemoHumanInTheLoop />,
        },
      ]}
    />
  );
}
