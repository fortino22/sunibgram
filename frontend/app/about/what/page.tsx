import type { Metadata } from 'next';
import { AboutSplitValue } from '@/components/about/AboutSplitValue';
import { AboutSplitConnection } from '@/components/about/AboutSplitConnection';

export const metadata: Metadata = {
  title: 'What is Sunibgram — The social network for Binus',
  description:
    'Everything campus in one place. Threads, communities, events, and DMs — all from people you share a campus with.',
};

export default function WhatPage() {
  return (
    <>
      <AboutSplitValue />
      <AboutSplitConnection />
    </>
  );
}
