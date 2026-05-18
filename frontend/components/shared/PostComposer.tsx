'use client';

import { ImagePlus, LoaderCircle, X, Loader2 } from 'lucide-react';
import {
  type ChangeEvent,
  type FormEvent,
  useEffect,
  useRef,
  useState,
} from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Avatar } from './Avatar';
import { useMyProfile } from '@/hooks/useMyProfile';
import { useAuthStore } from '@/stores/authStore';
import {
  createThread,
  type ThreadMutationParams,
  type ThreadResponse,
  type ThreadItem,
} from '@/lib/services/threads';
import type { ThreadDetailData } from '@/hooks/useThread';
import { compressImages } from '@/lib/compressImage';

type PreviewKind = 'image' | 'video' | 'file';

interface FilePreview {
  file: File;
  kind: PreviewKind;
  url: string;
}

export interface PostComposerProps {
  placeholder?: string;
  /** When set, the composer posts a reply to this thread ID instead of a new root thread. */
  parentId?: number;
  /** HTML id attribute for the <form> element. */
  formId?: string;
}

interface MutateContext {
  previousChain: [readonly unknown[], ThreadDetailData | undefined][];
  savedContent: string;
}

function getPreviewKind(file: File): PreviewKind {
  if (file.type.startsWith('image/')) return 'image';
  if (file.type.startsWith('video/')) return 'video';
  return 'file';
}

function isSameFile(left: File, right: File): boolean {
  return (
    left.name === right.name &&
    left.lastModified === right.lastModified &&
    left.size === right.size
  );
}

export function PostComposer({ placeholder = 'What is happening?!', parentId, formId }: PostComposerProps) {
  const isReply = parentId !== undefined;
  const queryClient = useQueryClient();
  const user = useAuthStore((state) => state.user);
  const token = useAuthStore((state) => state.token);
  const { data: profile } = useMyProfile();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const previewsRef = useRef<FilePreview[]>([]);
  const [content, setContent] = useState('');
  const [previews, setPreviews] = useState<FilePreview[]>([]);
  const [compressing, setCompressing] = useState(false);

  useEffect(() => {
    previewsRef.current = previews;
  }, [previews]);

  useEffect(() => {
    return () => {
      previewsRef.current.forEach((p) => URL.revokeObjectURL(p.url));
    };
  }, []);

  const avatarSrc = profile?.avatar ?? null;
  const selectedFiles = previews.map((p) => p.file);
  const normalizedContent = content.trim();

  const mutation = useMutation<ThreadResponse, Error, ThreadMutationParams, MutateContext | undefined>({
    mutationFn: async (params) => {
      if (token === null) throw new Error('Authentication token is required.');
      return createThread(params, token);
    },
    onMutate: isReply
      ? async (params) => {
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
        }
      : undefined,
    onError: (_err, _params, context) => {
      if (context) {
        for (const [key, data] of context.previousChain) {
          queryClient.setQueryData(key, data);
        }
        setContent(context.savedContent);
      }
    },
    onSuccess: async () => {
      const invalidations: Promise<void>[] = [
        queryClient.invalidateQueries({ queryKey: ['threads', 'feed'] }),
        queryClient.invalidateQueries({ queryKey: ['threads'] }),
      ];
      if (isReply) {
        invalidations.push(
          queryClient.invalidateQueries({ queryKey: ['thread-chain'] }),
          queryClient.invalidateQueries({ queryKey: ['thread', parentId] }),
        );
      }
      await Promise.all(invalidations);
    },
  });

  const isPending = mutation.isPending;
  const canSubmit =
    token !== null &&
    !isPending &&
    !compressing &&
    (normalizedContent.length > 0 || selectedFiles.length > 0);

  async function handleFilesChange(event: ChangeEvent<HTMLInputElement>) {
    const incoming = Array.from(event.target.files ?? []);
    event.target.value = '';
    if (incoming.length === 0) return;
    setCompressing(true);
    try {
      const ready = await compressImages(incoming);
      setPreviews((cur) => [
        ...cur,
        ...ready.map((file) => ({
          file,
          kind: getPreviewKind(file),
          url: URL.createObjectURL(file),
        })),
      ]);
    } finally {
      setCompressing(false);
    }
  }

  function removeSelectedFile(target: File) {
    setPreviews((cur) => {
      const toRemove = cur.find((p) => isSameFile(p.file, target));
      if (toRemove) URL.revokeObjectURL(toRemove.url);
      return cur.filter((p) => !isSameFile(p.file, target));
    });
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canSubmit) return;
    try {
      await mutation.mutateAsync({
        content: normalizedContent,
        files: selectedFiles,
        ...(isReply && { parentId }),
      });
      // For reply mode, content is already cleared optimistically in onMutate.
      if (!isReply) setContent('');
      previews.forEach((p) => URL.revokeObjectURL(p.url));
      setPreviews([]);
      if (fileInputRef.current) fileInputRef.current.value = '';
    } catch {
      // Mutation state already exposes the error.
    }
  }

  const inputId = `composer-files-${parentId ?? 'new'}`;

  return (
    <form
      id={formId}
      onSubmit={(e) => void handleSubmit(e)}
      style={{
        padding: '16px 20px',
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg-card)',
      }}
    >
      <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
        <Avatar username={user?.username ?? ''} avatarUrl={avatarSrc} size="md" />

        <div style={{ flex: 1, minWidth: 0 }}>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder={placeholder}
            rows={isReply ? 3 : 4}
            style={{
              width: '100%',
              resize: 'vertical',
              minHeight: isReply ? 76 : 108,
              background: 'transparent',
              border: 'none',
              outline: 'none',
              color: 'var(--text)',
              fontSize: 15,
              lineHeight: 1.6,
              fontFamily: 'inherit',
              padding: '12px 14px 10px',
            }}
          />

          {previews.length > 0 && (
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
                gap: 10,
                marginBottom: 12,
              }}
            >
              {previews.map((preview) => (
                <div
                  key={`${preview.file.name}-${preview.file.lastModified}-${preview.file.size}`}
                  style={{
                    position: 'relative',
                    borderRadius: 12,
                    overflow: 'hidden',
                    border: '1px solid var(--border)',
                    background: 'var(--bg-subtle)',
                    minHeight: 140,
                  }}
                >
                  {preview.kind === 'video' ? (
                    <video
                      src={preview.url}
                      muted
                      playsInline
                      controls
                      style={{ width: '100%', height: '100%', maxHeight: 220, display: 'block', objectFit: 'cover' }}
                    />
                  ) : preview.kind === 'image' ? (
                    <img
                      src={preview.url}
                      alt={preview.file.name}
                      style={{ width: '100%', height: '100%', maxHeight: 220, display: 'block', objectFit: 'cover' }}
                    />
                  ) : (
                    <div
                      style={{
                        minHeight: 140,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: 'var(--text-tertiary)',
                        fontSize: 13,
                        padding: 16,
                        textAlign: 'center',
                        lineHeight: 1.5,
                      }}
                    >
                      {preview.file.name}
                    </div>
                  )}
                  <button
                    type="button"
                    onClick={() => removeSelectedFile(preview.file)}
                    aria-label={`Remove ${preview.file.name}`}
                    style={{
                      position: 'absolute',
                      top: 8,
                      right: 8,
                      width: 28,
                      height: 28,
                      borderRadius: '50%',
                      border: '1px solid rgba(255,255,255,0.22)',
                      background: 'rgba(0,0,0,0.55)',
                      color: '#fff',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      cursor: 'pointer',
                    }}
                  >
                    <X size={14} />
                  </button>
                </div>
              ))}
            </div>
          )}

          <div style={{ height: 1, background: 'var(--border)', margin: '0 0 12px' }} />

          {mutation.isError && (
            <div
              style={{
                marginBottom: 12,
                padding: '10px 12px',
                borderRadius: 10,
                background: 'var(--coral-soft)',
                color: 'var(--coral)',
                fontSize: 13,
                lineHeight: 1.5,
              }}
            >
              {mutation.error.message}
            </div>
          )}

          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: 12,
              flexWrap: 'wrap',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*,video/*"
                multiple
                onChange={handleFilesChange}
                style={{ display: 'none' }}
                id={inputId}
              />
              <label
                htmlFor={inputId}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 8,
                  padding: '8px 12px',
                  borderRadius: 10,
                  border: '1px solid var(--border)',
                  color: compressing ? 'var(--accent)' : 'var(--text-secondary)',
                  fontSize: 13,
                  fontWeight: 500,
                  cursor: isPending || compressing ? 'not-allowed' : 'pointer',
                  opacity: isPending || compressing ? 0.6 : 1,
                  pointerEvents: compressing ? 'none' : undefined,
                }}
              >
                {compressing ? <Loader2 size={16} style={{ animation: 'spin 1s linear infinite' }} /> : <ImagePlus size={16} />}
              </label>

              {compressing && (
                <span style={{ color: 'var(--accent)', fontSize: 13 }}>Compressing…</span>
              )}
              {!compressing && selectedFiles.length > 0 && (
                <span style={{ color: 'var(--text-tertiary)', fontSize: 13 }}>
                  {selectedFiles.length} file{selectedFiles.length > 1 ? 's' : ''} selected
                </span>
              )}
            </div>

            <button
              type="submit"
              disabled={!canSubmit}
              style={{
                background: 'var(--accent)',
                border: 'none',
                borderRadius: 10,
                color: '#fff',
                fontWeight: 600,
                fontSize: 14,
                padding: '9px 20px',
                cursor: canSubmit ? 'pointer' : 'not-allowed',
                opacity: canSubmit ? 1 : 0.5,
                fontFamily: 'inherit',
                display: 'inline-flex',
                alignItems: 'center',
                gap: 8,
              }}
            >
              {isPending && <LoaderCircle size={16} />}
              <span>
                {isPending
                  ? isReply ? 'Replying...' : 'Posting...'
                  : isReply ? 'Reply' : 'Post'}
              </span>
            </button>
          </div>
        </div>
      </div>
    </form>
  );
}
