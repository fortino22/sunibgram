import type { Metadata } from 'next';
import { AboutFeaturePage } from '@/components/about/AboutFeaturePage';
import {
  AlgoDemoInterestScore,
  AlgoDemoReactionWeights,
  AlgoDemoTimeDecay,
  AlgoDemoFreshSlots,
  AlgoDemoReshuffle,
  AlgoDemoAuthorCap,
  AlgoDemoNewPostsBanner,
} from '@/components/about/demos/AlgorithmDemos';

export const metadata: Metadata = {
  title: 'How your feed works — Sunibgram',
  description:
    'A plain-language look at how the Sunibgram For You feed decides what to show you and in what order — interest scores, freshness, reaction weights, and the fairness limits that keep it varied.',
};

export default function AlgorithmAboutPage() {
  return (
    <AboutFeaturePage
      eyebrow="Transparency · For You"
      title="How your feed decides what to show."
      description="Your home feed isn’t random. For You blends posts from people you follow, the communities you’ve joined, and what’s buzzing across campus — then orders them by a quiet interest score. Here’s exactly how that works, in plain terms."
      sections={[
        {
          id: 'interest-score',
          title: 'What goes into For You',
          description:
            'Every post is given a quiet interest score. Three things shape it: how people are reacting to it, how fresh it is, and who posted it — your own posts, people you follow, your communities, and campus-wide trending posts, all blended together.',
          bullets: [
            'Reactions — likes, comments, reposts, bookmarks, and shares',
            'Freshness — newer posts get a boost',
            'Source — followed people, your communities, and trending posts',
          ],
          demo: <AlgoDemoInterestScore />,
        },
        {
          id: 'reaction-weights',
          title: 'Not all reactions count the same',
          description:
            'A like is great, but some reactions say more about how much a post resonated. Reposting, bookmarking, or commenting takes more effort — so those count for more than a plain like. Replies show up too, but are nudged down a little against original posts.',
          bullets: [
            'Reposts count the most',
            'Bookmarks and comments count a lot',
            'Shares count a bit more than a like; a like is the baseline',
          ],
          demo: <AlgoDemoReactionWeights />,
        },
        {
          id: 'time-decay',
          title: 'Newer posts rise, older ones fade',
          description:
            'Sunibgram is about what’s happening now. A post is strongest right after it’s posted, then gently fades over the next several hours. A buzzing post can hold near the top for a while, but a quiet older post slips down to make room for fresher ones.',
          bullets: [
            'Strongest the moment it’s posted, fading over the hours that follow',
            'Lots of reactions can keep a post up longer',
            'Posts older than about a week leave For You entirely',
          ],
          demo: <AlgoDemoTimeDecay />,
        },
        {
          id: 'fresh-slots',
          title: 'Fresh from the last few minutes',
          description:
            'Brand-new posts — from roughly the last 15 minutes — get a reserved spot near the top, newest first, even before they’ve collected reactions. That way you don’t miss something that just dropped.',
          bullets: [
            'A reserved spot for just-posted threads, newest first',
            'Shown before they’ve had time to gather reactions',
            'Only a couple appear at once so they don’t crowd everything out',
          ],
          demo: <AlgoDemoFreshSlots />,
        },
        {
          id: 'reshuffle',
          title: 'You won’t see the exact same feed twice',
          description:
            'When several posts are about equally interesting, Sunibgram gently shuffles them so the feed feels fresh each time you open it — instead of the identical order over and over. Posts stay in roughly the right place by interest; the shuffle just adds variety among similar ones.',
          bullets: [
            'Similar-interest posts are gently shuffled, not pinned',
            'Order varies between visits, ranking stays intact',
            'Refresh the feed to reshuffle and pull in the latest reactions',
          ],
          demo: <AlgoDemoReshuffle />,
        },
        {
          id: 'fairness',
          title: 'Fair to everyone',
          description:
            'No single account can take over your feed. Even if someone posts a lot, only a handful of their posts can appear near the top, with a limit on how many show up overall. This keeps your feed varied and gives more voices a chance to be seen.',
          bullets: [
            'A cap on how many posts one account places near the top',
            'A limit on how many of their posts show up overall',
            'More voices get a chance to surface',
          ],
          demo: <AlgoDemoAuthorCap />,
        },
        {
          id: 'new-posts-banner',
          title: 'The “New posts” banner',
          description:
            'While you’re scrolling, Sunibgram quietly checks for new posts every few minutes. When some arrive, a small “N new posts” banner appears at the top. Tap it and the new posts slide in on top — your current spot stays put and nothing reshuffles underneath you.',
          bullets: [
            'A quiet check for new posts every few minutes',
            'A small banner appears when fresh posts are waiting',
            'Tap to slide them in on top — your place is never lost',
          ],
          demo: <AlgoDemoNewPostsBanner />,
        },
      ]}
    />
  );
}
