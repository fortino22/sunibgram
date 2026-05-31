import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';

export const metadata: Metadata = {
  title: 'Events — Sunibgram',
  description:
    'Every campus event in one feed — official Binus events and student-run meetups.',
};

export default function EventsAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="Discover · Plan"
      title="Every campus event in one feed."
      description="From official Binus events to student-run meetups — see what is happening, RSVP, and never miss a thing on campus."
      highlights={[
        {
          title: 'Official events',
          description:
            'Verified by your faculty or student council, with attachments and program details.',
        },
        {
          title: 'Student-run meetups',
          description:
            'Anyone can organize, anyone can join. Submit your event and reach the whole campus.',
        },
        {
          title: 'Event threads',
          description:
            'Discuss the event before, during, and after — official and unofficial commentary in one place.',
        },
        {
          title: 'RSVPs',
          description:
            'Track who is going and reach out to attendees ahead of the event.',
        },
      ]}
    />
  );
}
