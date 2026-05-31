import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';

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
      highlights={[
        {
          title: 'Public & private rooms',
          description:
            'Decide who joins the conversation. Open communities for everyone, request-based for closed circles.',
        },
        {
          title: 'Moderation tools',
          description:
            'Owners and moderators can manage rules, members, and roles to keep groups on-topic.',
        },
        {
          title: 'Category discovery',
          description:
            'Browse communities by faculty, hobby, or location. Search across every Binus campus at once.',
        },
        {
          title: 'Member-only threads',
          description:
            'Group-specific posts and replies, visible only to members of the community.',
        },
      ]}
    />
  );
}
