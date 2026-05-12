import { LoaderCircle } from 'lucide-react';
import type { FormEvent } from 'react';
import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Avatar } from '@/components/shared/Avatar';
import { useMyProfile } from '@/hooks/useMyProfile';
import { createThread, type ThreadMutationParams, type ThreadResponse, type ThreadItem } from '@/lib/services';
import { useAuthStore } from '@/stores/authStore';
import type { ThreadDetailData } from '@/hooks/useThread';

interface ThreadComposerProps {
  placeholder: string;
  parentId: number;
}

interface MutateContext {
  previousChain: [readonly unknown[], ThreadDetailData | undefined][];
  savedContent: string;
}

export function ThreadComposer({ placeholder, parentId }: ThreadComposerProps) {
  const queryClient = useQueryClient();
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  const { data: profile } = useMyProfile();
  const [content, setContent] = useState('');
  const normalizedContent = content.trim();

  const replyMutation = useMutation<ThreadResponse, Error, ThreadMutationParams, MutateContext>({
    mutationFn: async (params) => {
      if (token === null) {
        throw new Error('Authentication token is required to reply to this thread.');
      }

      return createThread(params, token);
    },
    onMutate: async (params) => {
      await queryClient.cancelQueries({ queryKey: ['thread-chain'] });

      const previousChain = queryClient.getQueriesData<ThreadDetailData>({ queryKey: ['thread-chain'] });

      const now = new Date().toISOString();
      const optimisticReply: ThreadItem = {
        type: 'THREAD',
        id: -Date.now(),
        parent_id: parentId,
        root_id: null,
        community: undefined,
        content: params.content,
        created_at: now,
        updated_at: now,
        time_ago: 'now',
        author: {
          id: user?.id ?? 0,
          username: user?.username ?? '',
          is_premium: false,
          avatar_url: profile?.avatar ?? null,
          faculty: undefined,
        },
        stats: { likes: 0, reposts: 0, bookmarks: 0, comments: 0 },
        media: [],
        is_liked: false,
        is_bookmarked: false,
        is_reposted: false,
        is_pin: false,
      };

      queryClient.setQueriesData<ThreadDetailData>(
        { queryKey: ['thread-chain'] },
        (old) => {
          if (!old || old.root.id !== parentId) return old;
          return {
            ...old,
            root: {
              ...old.root,
              stats: { ...old.root.stats, comments: old.root.stats.comments + 1 },
            },
            replies: [...old.replies, optimisticReply],
          };
        }
      );

      const savedContent = content;
      setContent('');

      return { previousChain, savedContent };
    },
    onError: (_err, _params, context) => {
      if (context) {
        for (const [key, data] of context.previousChain) {
          queryClient.setQueryData(key, data);
        }
        setContent(context.savedContent);
      }
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['thread-chain'] }),
        queryClient.invalidateQueries({ queryKey: ['thread', parentId] }),
        queryClient.invalidateQueries({ queryKey: ['threads', 'feed'] }),
        queryClient.invalidateQueries({ queryKey: ['threads'] }),
      ]);
    },
  });

  const canSubmit = token !== null && normalizedContent.length > 0 && !replyMutation.isPending;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canSubmit) return;

    try {
      await replyMutation.mutateAsync({
        content: normalizedContent,
        parentId,
      });
    } catch {
      // Mutation state already exposes the error.
    }
  }

  return (
    <form
      onSubmit={(event) => void handleSubmit(event)}
      className="border-b"
      style={{
        borderColor: 'var(--border)',
        background: 'var(--bg-card)',
        padding: '14px 20px',
      }}
    >
      <div className="flex gap-3">
        <Avatar
          username={user?.username ?? 'Current user'}
          avatarUrl={profile?.avatar ?? null}
          size="md"
        />
        <div className="min-w-0 flex-1">
          <textarea
            value={content}
            onChange={(event) => setContent(event.target.value)}
            placeholder={placeholder}
            rows={3}
            className="min-h-[76px] w-full resize-none bg-transparent px-3 pt-3 pb-2 text-[15px] outline-none"
            style={{
              color: 'var(--text)',
              lineHeight: 1.6,
            }}
          />

          {replyMutation.isError && (
            <div
              className="mb-3 rounded-xl px-3 py-2 text-[13px]"
              style={{
                background: 'var(--coral-soft)',
                color: 'var(--coral)',
              }}
            >
              {replyMutation.error.message}
            </div>
          )}

          <div
            className="flex items-center justify-end border-t pt-3"
            style={{ borderColor: 'var(--border)' }}
          >
            <button
              type="submit"
              disabled={!canSubmit}
              className="inline-flex min-w-[92px] items-center justify-center gap-2 rounded-full px-4 py-2 text-sm font-semibold"
              style={{
                background: 'var(--accent)',
                border: 0,
                color: '#fff',
                cursor: canSubmit ? 'pointer' : 'not-allowed',
                opacity: canSubmit ? 1 : 0.5,
                fontFamily: 'inherit',
                boxShadow: canSubmit ? '0 8px 20px rgba(15, 20, 25, 0.14)' : 'none',
                transition: 'transform 160ms ease, box-shadow 160ms ease, opacity 160ms ease',
              }}
            >
              {replyMutation.isPending && <LoaderCircle size={16} className="animate-spin" />}
              <span>{replyMutation.isPending ? 'Replying' : 'Reply'}</span>
            </button>
          </div>
        </div>
      </div>
    </form>
  );
}
