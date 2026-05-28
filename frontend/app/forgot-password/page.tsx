'use client';

import { Suspense } from 'react';
import Link from 'next/link';
import { Logo } from '@/components/ui/Logo';
import { useSearchParams } from 'next/navigation';

import { AuthSplitLayout } from '@/components/auth/AuthSplitLayout';
import { ForgotPasswordFlow } from '@/components/auth/ForgotPasswordFlow';

function ForgotPasswordPageContent() {
  const searchParams = useSearchParams();
  const initialEmail = searchParams.get('email') ?? '';

  return (
    <AuthSplitLayout>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
          <Logo size={48} style={{ color: 'var(--text)' }} />
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
    </AuthSplitLayout>
  );
}

export default function ForgotPasswordPage() {
  return (
    <Suspense
      fallback={
        <AuthSplitLayout>
            <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
              <Logo size={48} style={{ color: 'var(--text)' }} />
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
        </AuthSplitLayout>
      }
    >
      <ForgotPasswordPageContent />
    </Suspense>
  );
}
