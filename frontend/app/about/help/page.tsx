import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';

export const metadata: Metadata = {
  title: 'Help Center — Sunibgram',
  description:
    'Guides, FAQs, and answers to the questions Binus students ask most often.',
};

export default function HelpAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="Support · Guides"
      title="How can we help?"
      description="Guides, FAQs, and answers to the questions Binus students ask most often. Stuck on something? Start here."
      ctaLabel="Contact support"
      ctaHref="mailto:support@sunibgram.com"
      highlights={[
        {
          title: 'Getting started',
          description:
            'Set up your verified account with your @binus.ac.id email and finish onboarding.',
        },
        {
          title: 'Account & privacy',
          description:
            'Manage who sees what — private accounts, blocked users, and visibility controls.',
        },
        {
          title: 'Threads & replies',
          description:
            'Posting, editing, deleting, and what happens when a thread you replied to is removed.',
        },
        {
          title: 'Communities & DMs',
          description:
            'Joining, leaving, and moderating groups. Reporting abuse and unsending messages.',
        },
      ]}
    />
  );
}
