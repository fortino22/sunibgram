import type { Metadata } from 'next';
import { AboutSplitValue } from '@/components/about/AboutSplitValue';

export const metadata: Metadata = {
  title: 'What is Sunibgram — The social network for Binus',
  description:
    'Everything campus in one place. Threads, communities, events, and DMs — all from people you share a campus with.',
};

export default function WhatPage() {
  return <AboutSplitValue />;
}
