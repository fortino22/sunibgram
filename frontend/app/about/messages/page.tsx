import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';

export const metadata: Metadata = {
  title: 'Messages — Sunibgram',
  description:
    'Real-time private messaging with verified Binus students — DMs and group rooms.',
};

export default function MessagesAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="DMs · Real-time"
      title="Private chats with classmates."
      description="Real-time DMs and group rooms — verified, private, and built for campus collaboration. No strangers, no impostors."
      highlights={[
        {
          title: 'One-on-one DMs',
          description:
            'Talk to anyone with a verified Binus identity — no friend requests, just one click.',
        },
        {
          title: 'Group rooms',
          description:
            'Spin up project chats with classmates. Add members, rename rooms, share descriptions.',
        },
        {
          title: 'Media & files',
          description:
            'Share notes, slides, and images instantly — everything stays inside the room.',
        },
        {
          title: 'Read receipts',
          description:
            'Know when your message has been seen. Edit or unsend if you change your mind.',
        },
      ]}
    />
  );
}
