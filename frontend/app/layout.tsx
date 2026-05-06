import type { Metadata } from 'next';
import { DM_Sans } from 'next/font/google';
import './globals.css';
import { MainWrapper } from '@/components/layout/MainWrapper';
import { Providers } from '@/components/Providers';

const dmSans = DM_Sans({
  subsets: ['latin'],
  variable: '--font-dm-sans',
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'Binusgram — Campus Discourse',
  description: 'University social media platform for Binus students',
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className={dmSans.variable}>
      <body>
        <Providers>
          <MainWrapper>{children}</MainWrapper>
        </Providers>
      </body>
    </html>
  );
}
