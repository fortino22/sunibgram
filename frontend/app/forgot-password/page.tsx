'use client';

import { Suspense } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { useSearchParams } from 'next/navigation';

import { ForgotPasswordFlow } from '@/components/auth/ForgotPasswordFlow';

function ForgotPasswordPageContent() {
  const searchParams = useSearchParams();
  const initialEmail = searchParams.get('email') ?? '';

  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'var(--bg)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px 16px',
      }}
    >
      <div style={{ width: '100%', maxWidth: 400 }}>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
          <Image src="/assets/sunibgram/logo.png" alt="Sunibgram" width={48} height={48} />
        </div>

        <div
          style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 16,
            padding: '28px 24px',
          }}
        >
          <ForgotPasswordFlow initialEmail={initialEmail} />
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 12, margin: '20px 0' }}>
          <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
          <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>or</span>
          <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
        </div>
        <div style={{ textAlign: 'center', fontSize: 14, color: 'var(--text-tertiary)' }}>
          Remember your password?{' '}
          <Link
            href="/login"
            style={{ color: 'var(--accent)', fontWeight: 500, textDecoration: 'none' }}
          >
            Sign in
          </Link>
        </div>
      </div>
    </div>
  );
}

export default function ForgotPasswordPage() {
  return (
    <Suspense
      fallback={
        <div
          style={{
            minHeight: '100vh',
            background: 'var(--bg)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '24px 16px',
          }}
        >
          <div style={{ width: '100%', maxWidth: 400 }}>
            <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
              <Image src="/assets/sunibgram/logo.png" alt="Sunibgram" width={48} height={48} />
            </div>

            <div
              style={{
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 16,
                padding: '28px 24px',
              }}
            >
              <div style={{ textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>
                Loading…
              </div>
            </div>
          </div>
        </div>
      }
    >
      <ForgotPasswordPageContent />
    </Suspense>
  );
}
