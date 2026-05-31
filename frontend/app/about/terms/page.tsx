import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';

export const metadata: Metadata = {
  title: 'Terms & Conditions — Sunibgram',
  description:
    'Rules and policies for using Sunibgram — keeping the platform safe for every Binus student.',
};

export default function TermsAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="Legal · Terms"
      title="Rules of the road."
      description="By using Sunibgram, you agree to these terms — they keep the platform safe and useful for every Binus student. The full policy is available in your account settings."
      ctaLabel="Join with your Binus email"
      ctaHref="/signup"
      highlights={[
        {
          title: 'Eligibility',
          description:
            'A verified @binus.ac.id email is required. One account per student, no exceptions.',
        },
        {
          title: 'Conduct',
          description:
            'Respect, honesty, and zero tolerance for harassment, hate speech, or impersonation.',
        },
        {
          title: 'Content ownership',
          description:
            'You own your posts. We host them and grant the platform a limited license to display them.',
        },
        {
          title: 'Termination',
          description:
            'Accounts can be suspended or removed for policy violations. You can also delete yours anytime.',
        },
      ]}
    />
  );
}
