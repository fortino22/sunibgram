import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';

export const metadata: Metadata = {
  title: 'Privacy Center — Sunibgram',
  description:
    'How Sunibgram collects, stores, and protects the information you share with us.',
};

export default function PrivacyAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="Privacy · Data"
      title="Your data, your control."
      description="How Sunibgram collects, stores, and protects the information you share with us. Plain language, no fine print buried at the bottom."
      ctaLabel="Manage my data"
      ctaHref="/settings"
      highlights={[
        {
          title: 'What we collect',
          description:
            'Your profile, posts, and platform usage — all tied to your verified Binus identity.',
        },
        {
          title: 'How we use it',
          description:
            'Only to provide and improve Sunibgram. No advertising profiles, no data brokers.',
        },
        {
          title: 'Your controls',
          description:
            'Download, edit, or delete your data anytime from your account settings.',
        },
        {
          title: 'Sharing',
          description:
            'We never sell or trade your data to third parties. Internal access is logged and limited.',
        },
      ]}
    />
  );
}
