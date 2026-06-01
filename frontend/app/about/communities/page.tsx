import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';
import {
  CommunitiesDemoVisibility,
  CommunitiesDemoModeration,
  CommunitiesDemoDiscovery,
  CommunitiesDemoMemberThreads,
} from '@/components/about/demos/CommunitiesDemos';

export const metadata: Metadata = {
  title: 'Communities — Sunibgram',
  description:
    'Join groups around the courses, clubs, and interests that define your time at Binus.',
};

export default function CommunitiesAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="Groups · Interests"
      title="Find your people on campus."
      description="Join groups around the courses, clubs, and interests that define your time at Binus. Public for the open conversations, private for the close-knit ones."
      sections={[
        {
          id: 'public-private',
          title: 'Public & private rooms',
          description:
            'Decide who joins. Open communities let anyone in with one tap, private ones gate every join behind owner approval.',
          bullets: [
            'Public: anyone with a Binus account can join instantly',
            'Private: owners review every request before letting members in',
            'Switch visibility from settings at any time',
          ],
          demo: <CommunitiesDemoVisibility />,
        },
        {
          id: 'moderation',
          title: 'Moderation tools',
          description:
            'Owners and moderators keep groups on-topic. Promote trusted members, manage rules, and kick anyone who breaks the room.',
          bullets: [
            'Promote members to MODERATOR with a single tap',
            'Pin rules to the top so newcomers see them first',
            'Remove members or revoke roles instantly',
          ],
          demo: <CommunitiesDemoModeration />,
        },
        {
          id: 'category-discovery',
          title: 'Category discovery',
          description:
            'Browse communities by faculty, hobby, or location — and search across every Binus campus at once.',
          bullets: [
            'Curated categories cover academics, sports, tech, and arts',
            'Search matches names, descriptions, and rules',
            'See member counts so you know how active a room is',
          ],
          demo: <CommunitiesDemoDiscovery />,
        },
        {
          id: 'member-threads',
          title: 'Member-only threads',
          description:
            'Group threads stay inside the room. Members see the whole conversation; non-members see only that something is happening behind the door.',
          bullets: [
            'Posts visible only to verified members',
            'Replies, polls, and media all stay private',
            'Join to read — never to peek',
          ],
          demo: <CommunitiesDemoMemberThreads />,
        },
      ]}
    />
  );
}
