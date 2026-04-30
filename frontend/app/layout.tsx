import type { Metadata } from 'next';
import { DM_Sans } from 'next/font/google';
import './globals.css';
import { Sidebar } from '@/components/layout/Sidebar';
import { RightPanel } from '@/components/layout/RightPanel';

const dmSans = DM_Sans({
  subsets: ['latin'],
  variable: '--font-dm-sans',
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'Binusgram — Campus Discourse',
  description: 'University social media platform',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={dmSans.variable}>
      <body>
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            minHeight: '100vh',
            background: 'var(--bg)',
          }}
        >
          <div
            style={{
              display: 'flex',
              width: '100%',
              maxWidth: 1200,
              minHeight: '100vh',
            }}
          >
            <Sidebar />
            <main
              style={{
                flex: 1,
                minWidth: 0,
                minHeight: '100vh',
                background: 'var(--bg-card)',
                borderLeft: '1px solid var(--border)',
                borderRight: '1px solid var(--border)',
              }}
            >
              {children}
            </main>
            <RightPanel />
          </div>
        </div>
      </body>
    </html>
  );
}
