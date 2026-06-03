import type { Metadata } from 'next';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';
import { AboutHeroBackground } from '@/components/about/AboutHeroBackground';
import { getNewsBySlug, formatNewsDate } from '@/lib/news';

const SLUG = 'sunibgram-enters-beta';
const item = getNewsBySlug(SLUG)!;

export const metadata: Metadata = {
  title: `${item.title} — Sunibgram`,
  description: item.lede,
};

export default function BetaAnnouncementPage() {
  return (
    <>
      <section className="relative overflow-hidden">
        <AboutHeroBackground />
        <div className="about-hero-grain" aria-hidden />
        <div
          className="about-container-narrow relative z-10"
          style={{ paddingTop: 56, paddingBottom: 56 }}
        >
          <Link
            href="/about/news"
            className="about-news-back"
          >
            <ArrowLeft size={14} /> All news
          </Link>
          <div style={{ marginTop: 28 }}>
            <span className="about-eyebrow">{item.category}</span>
            <h1
              style={{
                color: 'var(--text)',
                fontWeight: 500,
                fontSize: 'clamp(2.25rem, 5vw, 3.5rem)',
                lineHeight: 1.1,
                letterSpacing: '-0.025em',
                marginTop: 20,
                maxWidth: '46rem',
              }}
            >
              {item.title}
            </h1>
            <p
              style={{
                color: 'var(--text)',
                fontSize: '1.0625rem',
                lineHeight: 1.65,
                marginTop: 22,
                maxWidth: '38rem',
              }}
            >
              {item.lede}
            </p>
            <div
              style={{
                marginTop: 28,
                fontSize: 13,
                color: 'var(--text-muted)',
                letterSpacing: '0.02em',
                textTransform: 'uppercase',
                fontWeight: 600,
              }}
            >
              {formatNewsDate(item.date)} · Sunibgram team
            </div>
          </div>
        </div>
      </section>

      <article className="about-content-section">
        <div className="about-container">
          <div className="about-news-article">
            <p className="about-news-p">
              Today we&apos;re opening Sunibgram to every Binusian with a
              verified <strong>@binus.ac.id</strong> email. The campus-only
              social network has been in private testing — now anyone on
              campus can sign in, set up a profile, and start posting.
            </p>
            <p className="about-news-p">
              One identity, one feed, no anonymous accounts. The people you
              follow are people you can actually run into between classes.
            </p>

            <h2 className="about-news-h2">What you can do today</h2>
            <p className="about-news-p">
              The beta ships with the core of the app — enough to make it a
              place you can actually live on between lectures:
            </p>
            <ul className="about-news-ul">
              <li>
                <strong>Verified sign-in</strong> with your Binus email, a
                one-time code, and a single account that follows you across
                features.
              </li>
              <li>
                <strong>Threads</strong> — post, reply, like, bookmark, and
                repost. Pin the ones you want at the top of your profile.
              </li>
              <li>
                <strong>Profiles &amp; follows</strong> with editable avatar,
                banner, and bio, plus follower / following lists.
              </li>
              <li>
                <strong>Communities</strong> — start a group for your club,
                study cohort, or interest. Set rules, manage roles, and
                handle join requests for private rooms.
              </li>
              <li>
                <strong>Faculty feed</strong> so you can see what people in
                your own faculty are actually talking about.
              </li>
            </ul>

            <h2 className="about-news-h2">What&apos;s still cooking</h2>
            <p className="about-news-p">
              A few things are landing through the rest of the beta and
              we&apos;d rather flag them than pretend they&apos;re done:
            </p>
            <ul className="about-news-ul">
              <li>
                Real-time <strong>notifications</strong> and in-app sockets.
              </li>
              <li>
                <strong>Direct messages</strong> — private and group rooms,
                read receipts, and attachments.
              </li>
              <li>
                <strong>Events</strong> with official approvals, attachments,
                and event-linked threads.
              </li>
              <li>
                <strong>Explore &amp; search</strong> across top threads,
                latest threads, people, media, and communities.
              </li>
              <li>
                Trending <strong>hashtags</strong>, @mention autocomplete,
                and a private-account toggle.
              </li>
            </ul>

            <h2 className="about-news-h2">How to join</h2>
            <p className="about-news-p">
              Head to the sign-up page, drop in your{' '}
              <strong>@binus.ac.id</strong> email, and verify with the code
              we send. That&apos;s it — no waitlist, no invite codes.
            </p>
            <p className="about-news-p">
              Found something broken or have an idea? Use the in-app
              feedback link. The team reads everything, and beta is when
              feedback actually shapes the roadmap.
            </p>

            <div className="about-news-cta-row">
              <Link
                href="/signup"
                className="about-btn-primary about-btn-primary-inverted"
                style={{
                  color: 'var(--text)',
                  borderColor: 'rgba(0, 0, 0, 0.18)',
                }}
              >
                Join with your Binus email
              </Link>
              <Link href="/about/threads" className="about-btn-dark">
                See the features
              </Link>
            </div>
          </div>
        </div>
      </article>

      <AboutFinalCTA />
    </>
  );
}
