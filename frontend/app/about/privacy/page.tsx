import type { Metadata } from 'next';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutLegalDoc, type LegalSection } from '@/components/about/AboutLegalDoc';

export const metadata: Metadata = {
  title: 'Privacy Center — Sunibgram',
  description:
    'How Sunibgram collects, uses, and protects the information of verified Binus students.',
};

const SECTIONS: LegalSection[] = [
  {
    id: 'who-we-are',
    title: 'Who this notice applies to',
    body: (
      <>
        <p>
          This notice describes how the Sunibgram service handles personal information.
          Sunibgram is a student-built campus social network operated independently and
          is not officially affiliated with Binus University.
        </p>
        <p>
          It applies to everyone who creates an account, browses content, or
          communicates with us through Sunibgram.
        </p>
      </>
    ),
  },
  {
    id: 'what-we-collect',
    title: 'Information we collect',
    body: (
      <>
        <p>We collect only what we need to run the service responsibly.</p>
        <ul>
          <li>
            <strong>Account information</strong> — your verified <code>@binus.ac.id</code>{' '}
            email address, chosen username, display name, faculty, and password (stored
            as a one-way hash).
          </li>
          <li>
            <strong>Content you post</strong> — threads, replies, polls, community
            posts, direct messages, files and media you upload, and reactions you give.
          </li>
          <li>
            <strong>Usage signals</strong> — basic logs of which features you use, when,
            and from which device, used to keep the service stable and to investigate
            abuse.
          </li>
          <li>
            <strong>Technical metadata</strong> — IP address, browser, and device type,
            kept only as long as needed for security and rate-limiting.
          </li>
        </ul>
        <p>
          We do not collect contact lists, browsing history outside Sunibgram, or
          location beyond what is implied by your network address.
        </p>
      </>
    ),
  },
  {
    id: 'how-we-use-it',
    title: 'How we use your information',
    body: (
      <>
        <p>We use the information above only to:</p>
        <ul>
          <li>Provide and improve the Sunibgram features you sign up for.</li>
          <li>Verify that each account belongs to a current Binus student.</li>
          <li>Show the right content to the right people (faculty feeds, communities you joined, conversations you&apos;re part of).</li>
          <li>Protect the platform from spam, abuse, and unauthorized access.</li>
          <li>Respond to your support requests and reports.</li>
        </ul>
        <p>
          We do not sell your data, rent it out to advertisers, or build advertising
          profiles. There are no third-party ad trackers on Sunibgram.
        </p>
      </>
    ),
  },
  {
    id: 'who-sees-what',
    title: 'Who sees your content',
    body: (
      <>
        <p>By default, content visibility follows where you posted it:</p>
        <ul>
          <li>
            <strong>Public threads</strong> are visible to any verified Sunibgram user.
          </li>
          <li>
            <strong>Community threads</strong> are visible only to verified members of
            that community.
          </li>
          <li>
            <strong>Direct messages and group rooms</strong> are visible only to the
            people you sent them to.
          </li>
          <li>
            <strong>Faculty feeds</strong> show posts authored by people in the same
            faculty; only verified students see them.
          </li>
        </ul>
        <p>
          If you switch your account to private from your settings, only people you
          approve can see your full profile and posts.
        </p>
      </>
    ),
  },
  {
    id: 'your-controls',
    title: 'Your controls',
    body: (
      <>
        <p>You stay in charge of your information.</p>
        <ul>
          <li>
            <strong>Edit your profile, avatar, and banner</strong> at any time from
            account settings.
          </li>
          <li>
            <strong>Block or unfollow</strong> any user from their profile.
          </li>
          <li>
            <strong>Unsend messages</strong> to remove them from a conversation for
            everyone.
          </li>
          <li>
            <strong>Switch to a private account</strong> so new followers need your
            approval.
          </li>
          <li>
            <strong>Delete your account</strong> from settings — your threads, replies,
            and messages are removed from the active service.
          </li>
        </ul>
        <p>
          For anything that account settings don&apos;t cover — including data export
          requests — write to <strong>support@sunibgram.com</strong>.
        </p>
      </>
    ),
  },
  {
    id: 'security',
    title: 'How we protect your information',
    body: (
      <>
        <p>
          Passwords are hashed before storage — never stored or transmitted in plain
          text. Sessions use short-lived tokens refreshed through a secure, httpOnly
          cookie.
        </p>
        <p>
          Uploaded files are kept on managed object storage with restricted access. We
          patch the operating system and dependencies on a regular cadence and treat
          security reports with priority.
        </p>
        <p>
          If you believe you&apos;ve found a security issue, please write directly to{' '}
          <strong>support@sunibgram.com</strong> with the subject{' '}
          <em>Security report</em> and give us a reasonable window to respond before
          public disclosure.
        </p>
      </>
    ),
  },
  {
    id: 'retention',
    title: 'How long we keep data',
    body: (
      <>
        <p>
          Content you posted stays on Sunibgram until you delete it or your account.
          When you delete your account, your active content is removed from the
          service. Some operational logs may be retained for a short period for
          safety, fraud prevention, and to satisfy legal obligations.
        </p>
      </>
    ),
  },
  {
    id: 'changes',
    title: 'Changes to this notice',
    body: (
      <>
        <p>
          We update this notice when our practices change. The <em>Last updated</em>{' '}
          date above always reflects the most recent revision. For substantial changes
          we&apos;ll surface a notice inside the app the next time you sign in.
        </p>
      </>
    ),
  },
  {
    id: 'contact',
    title: 'Contact',
    body: (
      <>
        <p>
          Questions about how Sunibgram handles your information? Write to{' '}
          <strong>support@sunibgram.com</strong>. We read every message — usually
          within a couple of working days.
        </p>
      </>
    ),
  },
];

export default function PrivacyAboutPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="Privacy center"
        title="Your data, plainly explained."
        description="How Sunibgram collects, uses, and protects the information of verified Binus students. Written in plain language, no buried fine print."
        meta="Last updated · January 2026"
      />
      <section className="about-content-section">
        <div className="about-container">
          <AboutLegalDoc sections={SECTIONS} />
        </div>
      </section>
    </>
  );
}
