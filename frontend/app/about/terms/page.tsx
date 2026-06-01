import type { Metadata } from 'next';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutLegalDoc, type LegalSection } from '@/components/about/AboutLegalDoc';

export const metadata: Metadata = {
  title: 'Terms & Conditions — Sunibgram',
  description:
    'The rules for using Sunibgram — the campus social network for verified Binus students.',
};

const SECTIONS: LegalSection[] = [
  {
    id: 'acceptance',
    title: 'Acceptance of these terms',
    body: (
      <>
        <p>
          By creating an account or using Sunibgram, you agree to be bound by these
          terms. If you do not agree, please do not create an account or use the
          service.
        </p>
        <p>
          Sunibgram is operated by a team of students, independently of Binus
          University. Throughout this document, &ldquo;we&rdquo;, &ldquo;us&rdquo;,
          and &ldquo;Sunibgram&rdquo; refer to that team.
        </p>
      </>
    ),
  },
  {
    id: 'eligibility',
    title: 'Who can use Sunibgram',
    body: (
      <>
        <p>
          Sunibgram is open only to current Binus students. To create an account you
          must:
        </p>
        <ul>
          <li>
            Have a valid <code>@binus.ac.id</code> email address and complete
            verification by email.
          </li>
          <li>Provide accurate information for your profile, including your faculty.</li>
          <li>Hold no more than one Sunibgram account at a time.</li>
        </ul>
        <p>
          We may suspend or terminate accounts that we have a good-faith reason to
          believe are not eligible.
        </p>
      </>
    ),
  },
  {
    id: 'your-account',
    title: 'Your account',
    body: (
      <>
        <p>
          You are responsible for what happens under your account. Keep your password
          private, and don&apos;t share login details. If you suspect someone else has
          accessed your account, change your password and let us know.
        </p>
        <p>
          You can edit your profile, update your password, or delete your account at
          any time from your settings.
        </p>
      </>
    ),
  },
  {
    id: 'your-content',
    title: 'Your content',
    body: (
      <>
        <p>
          You keep ownership of the threads, replies, messages, files, images, polls,
          and other content you post on Sunibgram (your <em>Content</em>).
        </p>
        <p>
          By posting Content, you grant Sunibgram a non-exclusive, royalty-free licence
          to host, display, and distribute it on the service, only as necessary to
          show it to the people you intended (your followers, the community you
          posted in, the people you sent it to, etc.).
        </p>
        <p>
          You confirm that you have the right to post your Content — for example, that
          you own the media you upload or have permission to share it.
        </p>
      </>
    ),
  },
  {
    id: 'acceptable-use',
    title: 'Acceptable use',
    body: (
      <>
        <p>To keep Sunibgram useful and safe for campus, you agree not to:</p>
        <ul>
          <li>Harass, threaten, dox, or impersonate other students.</li>
          <li>Post content that is illegal, sexually explicit, violent, or designed to incite hate.</li>
          <li>Share academic dishonesty material (exam leaks, paid assignment services, etc.).</li>
          <li>Spam, scrape, or otherwise automate the service without our written permission.</li>
          <li>Circumvent moderation actions, blocks, bans, or rate limits.</li>
          <li>Attempt to break into, probe, or stress-test the service outside our security-report process.</li>
          <li>Use Sunibgram for commercial promotion that isn&apos;t clearly a campus event or community.</li>
        </ul>
        <p>
          We may remove Content, restrict features, or suspend accounts that break
          these rules.
        </p>
      </>
    ),
  },
  {
    id: 'communities',
    title: 'Communities and moderation',
    body: (
      <>
        <p>
          Communities are run by their owners and the admins they appoint. Owners and
          admins are responsible for enforcing community rules within the rooms they
          manage.
        </p>
        <p>
          We may step in if a community is being used to break these terms, but
          day-to-day moderation is a community-level responsibility.
        </p>
      </>
    ),
  },
  {
    id: 'reports',
    title: 'Reports and enforcement',
    body: (
      <>
        <p>
          Any verified user can report a thread, message, or profile they believe
          breaks the rules. Reports go through our moderation queue and are usually
          reviewed within a couple of working days.
        </p>
        <p>
          Actions we may take include warnings, content removal, feature
          restrictions, temporary suspensions, and permanent account termination —
          depending on the severity and history.
        </p>
      </>
    ),
  },
  {
    id: 'intellectual-property',
    title: 'Sunibgram intellectual property',
    body: (
      <>
        <p>
          The Sunibgram name, logo, and the software that powers the service belong to
          the Sunibgram team. You may not copy the source code, scrape branding, or
          present yourself as Sunibgram without our written permission.
        </p>
      </>
    ),
  },
  {
    id: 'termination',
    title: 'Termination',
    body: (
      <>
        <p>
          You can delete your account at any time from settings. We can suspend or
          terminate your access if you break these terms, if your eligibility changes
          (for example, you are no longer a Binus student), or if we&apos;re required
          to do so by law.
        </p>
      </>
    ),
  },
  {
    id: 'disclaimer',
    title: 'Disclaimer of warranties',
    body: (
      <>
        <p>
          Sunibgram is provided <em>as is</em> and <em>as available</em>. We do our
          best to keep it running, but we don&apos;t promise it will be free of bugs,
          outages, or data loss. Always keep your own copies of anything important.
        </p>
        <p>
          Sunibgram is a student project. While we take security and reliability
          seriously, we are not an enterprise vendor and don&apos;t offer service-level
          guarantees.
        </p>
      </>
    ),
  },
  {
    id: 'liability',
    title: 'Limitation of liability',
    body: (
      <>
        <p>
          To the extent permitted by applicable law, Sunibgram and its team are not
          liable for indirect, incidental, or consequential damages arising from your
          use of the service. Our total liability for any direct claim relating to the
          service is limited to the amount you have paid us in the previous twelve
          months (which, for free accounts, is zero).
        </p>
      </>
    ),
  },
  {
    id: 'changes',
    title: 'Changes to these terms',
    body: (
      <>
        <p>
          We may update these terms when the service changes. The <em>Last updated</em>{' '}
          date above always reflects the most recent revision. We&apos;ll surface
          substantial changes inside the app the next time you sign in. Continuing to
          use Sunibgram after a change means you accept the updated terms.
        </p>
      </>
    ),
  },
  {
    id: 'governing-law',
    title: 'Governing law',
    body: (
      <>
        <p>
          These terms are governed by the laws of the Republic of Indonesia. Any
          dispute that we can&apos;t resolve informally will be handled in the courts
          of Jakarta.
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
          Questions about these terms? Write to{' '}
          <strong>support@sunibgram.com</strong>.
        </p>
      </>
    ),
  },
];

export default function TermsAboutPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="Terms & conditions"
        title="The rules for using Sunibgram."
        description="Plain-language terms for everyone who uses the Sunibgram service. Please read carefully before signing up."
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
