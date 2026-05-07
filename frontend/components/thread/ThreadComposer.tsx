import { Avatar } from '@/components/shared/Avatar';

interface ThreadComposerProps {
  placeholder: string;
  username?: string;
  avatarUrl?: string | null;
}

export function ThreadComposer({
  placeholder,
  username = 'Current user',
  avatarUrl = null,
}: ThreadComposerProps) {
  return (
    <section className="flex gap-3 border-b p-4" style={{ borderColor: 'var(--border)' }}>
      <Avatar username={username} avatarUrl={avatarUrl} size="md" />
      <textarea
        placeholder={placeholder}
        rows={3}
        className="min-h-[76px] flex-1 resize-y bg-transparent text-[15px] outline-none"
        style={{
          color: 'var(--text)',
          lineHeight: 1.5,
        }}
      />
    </section>
  );
}
