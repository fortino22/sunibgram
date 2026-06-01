import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';
import {
  EventsDemoOfficial,
  EventsDemoStudentMeetups,
  EventsDemoEventThreads,
  EventsDemoRSVPs,
} from '@/components/about/demos/EventsDemos';

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
      sections={[
        {
          id: 'official-events',
          title: 'Official events',
          description:
            'Events submitted by faculty or student councils go through a quick approval flow. Once approved, they get a verified badge so students know what is real.',
          bullets: [
            'Submit with an attachment for verification',
            'Admin can approve, decline, or request revisions',
            'Approved events earn a verified badge in the feed',
          ],
          demo: <EventsDemoOfficial />,
        },
        {
          id: 'student-meetups',
          title: 'Student-run meetups',
          description:
            'Anyone can organize. Submit a meetup in a few fields and the whole campus can find it — futsal, study sessions, coffee hangs.',
          bullets: [
            'Quick form: title, time, location, audience',
            'Publish instantly — no admin gate for unofficial events',
            'Edit or cancel from your event dashboard',
          ],
          demo: <EventsDemoStudentMeetups />,
        },
        {
          id: 'event-threads',
          title: 'Event threads',
          description:
            'Every event has its own thread feed. Discuss before, during, and after — official announcements and student commentary side by side.',
          bullets: [
            'Filter to Official only or see everything',
            'Reply with photos and live updates',
            'Pin the latest announcement to the top',
          ],
          demo: <EventsDemoEventThreads />,
        },
        {
          id: 'rsvps',
          title: 'RSVPs',
          description:
            'Track who is going and reach out to attendees ahead of time. See the headcount grow in real time as classmates tap RSVP.',
          bullets: [
            'Live attendance counter',
            'See familiar faces in the going list',
            'DM any attendee straight from the event',
          ],
          demo: <EventsDemoRSVPs />,
        },
      ]}
    />
  );
}
