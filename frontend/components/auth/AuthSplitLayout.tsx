import { AuthBrandPanel } from '@/components/auth/AuthBrandPanel';

/**
 * x.com-style split layout for the auth pages: the form column sits on the
 * left, the cursor-tracking brand panel on the right (right panel hidden
 * below the `lg` breakpoint so mobile keeps a clean centered form).
 */
export function AuthSplitLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen" style={{ background: 'var(--bg)' }}>
      <div className="flex flex-1 items-center justify-center lg:pl-[18%]" style={{ padding: '24px 16px' }}>
        <div style={{ width: '100%', maxWidth: 400 }}>{children}</div>
      </div>
      <AuthBrandPanel />
    </div>
  );
}
