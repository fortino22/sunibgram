import Link from 'next/link';
import { Logo } from '@/components/ui/Logo';

export default function LinkExpiredPage() {
  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--bg)',
        padding: 24,
      }}
    >
      <div
        style={{
          width: '100%',
          maxWidth: 400,
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          padding: '40px 32px',
          textAlign: 'center',
          animation: 'modal-panel-in 0.25s ease',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 24 }}>
          <Logo size={44} bg="var(--bg-card)" style={{ color: 'var(--text)' }} />
        </div>

        <h1
          style={{
            fontSize: 24,
            fontWeight: 800,
            letterSpacing: '-0.03em',
            color: 'var(--text)',
            margin: '0 0 10px 0',
          }}
        >
          Link expired
        </h1>
        <p
          style={{
            fontSize: 15,
            color: 'var(--text-tertiary)',
            lineHeight: 1.5,
            margin: '0 0 28px 0',
          }}
        >
          This sign-in confirmation link is no longer valid. If you didn&apos;t recognize that
          sign-in, change your password to secure your account.
        </p>

        <Link
          href="/login"
          className="btn-solid"
          style={{
            display: 'block',
            width: '100%',
            padding: '12px 0',
            background: 'var(--text)',
            color: 'var(--bg)',
            textDecoration: 'none',
            borderRadius: 9999,
            fontSize: 15,
            fontWeight: 700,
          }}
        >
          Back to login
        </Link>

        <div style={{ marginTop: 18, fontSize: 14, color: 'var(--text-tertiary)' }}>
          Worried about your account?{' '}
          <Link href="/forgot-password" className="hover-dim" style={{ color: 'var(--accent)', fontWeight: 500, textDecoration: 'none' }}>
            Reset password
          </Link>
        </div>
      </div>
    </div>
  );
}
