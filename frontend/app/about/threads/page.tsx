import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';

export const metadata: Metadata = {
  title: 'Threads — Sunibgram',
  description:
    'Post, reply, and follow conversations from people who actually share a campus with you.',
};

export default function ThreadsAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="Share · Verified"
      title="Threads built for campus life."
      description="Post, reply, and follow conversations from people who actually share a campus with you. No bots, no anonymous noise — every thread is tied to a real Binus identity."
      highlights={[
        {
          title: 'Threads & replies',
          description:
            'Continue conversations with classmates beyond the lecture hall, with full reply chains and quote reposts.',
        },
        {
          title: 'Faculty feed',
          description:
            'See what is happening across your school in real time, filtered to people who share your faculty.',
        },
        {
          title: 'Reposts & quotes',
          description:
            'Amplify the posts that matter. Add your take, or just pass it along.',
        },
        {
          title: 'Polls',
          description:
            'Ask quick questions and get answers from people you actually know — image polls included.',
        },
      ]}
    />
  );
}
