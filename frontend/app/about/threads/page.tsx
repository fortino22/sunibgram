import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';
import {
  ThreadsDemoReplies,
  ThreadsDemoFacultyFeed,
  ThreadsDemoReposts,
  ThreadsDemoPolls,
} from '@/components/about/demos/ThreadsDemos';

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
      sections={[
        {
          id: 'threads-replies',
          title: 'Threads & replies',
          description:
            'Continue conversations with classmates beyond the lecture hall. Reply chains keep every discussion threaded so nothing gets lost.',
          bullets: [
            'Threaded replies with the full chain in one view',
            'Mention classmates with @username autocomplete',
            'Edit or pin your own threads at any time',
          ],
          demo: <ThreadsDemoReplies />,
        },
        {
          id: 'faculty-feed',
          title: 'Faculty feed',
          description:
            'See what is happening across your school in real time. Filter the feed to the people who share your faculty so the most relevant voices rise.',
          bullets: [
            'Switch between For you, Following, and Faculty in one tap',
            'Powered by your verified faculty from Binus',
            'No bots, no off-campus noise',
          ],
          demo: <ThreadsDemoFacultyFeed />,
        },
        {
          id: 'reposts-quotes',
          title: 'Reposts & quotes',
          description:
            'Amplify the posts that matter — add your take with a quote, or pass it along with a clean repost. Either way, the original credit stays put.',
          bullets: [
            'One-tap repost, or quote with your own comment',
            'Quoted threads embed inline for context',
            'Undo with a single tap if you change your mind',
          ],
          demo: <ThreadsDemoReposts />,
        },
        {
          id: 'polls',
          title: 'Polls',
          description:
            'Ask quick questions and get answers from people you actually know. Text or image polls, with live results as students vote.',
          bullets: [
            'Up to 4 options, with optional images',
            'Set a duration from a few hours to a week',
            'Results animate in live — no refresh needed',
          ],
          demo: <ThreadsDemoPolls />,
        },
      ]}
    />
  );
}
