'use client';

import { ImagePlus, LoaderCircle, X, Loader2, Globe, ChevronDown, Smile, Film, Users, Check } from 'lucide-react';
import {
  type ChangeEvent,
  type FormEvent,
  useEffect,
  useRef,
  useState,
} from 'react';
import dynamic from 'next/dynamic';
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
import {
  createCommunityThread,
  type CommunityData,
  type CommunityMessageResponse,
} from '@/lib/services/communities';
import { useJoinedCommunities } from '@/hooks/useCommunities';
import type { ThreadDetailData } from '@/hooks/useThread';
import { compressImages, IMAGE_MAX_BYTES } from '@/lib/compressImage';
import { toast } from 'sonner';
import { GifPicker } from '@/components/chat/GifPicker';
import type { EmojiClickData } from 'emoji-picker-react';

const EmojiPicker = dynamic(() => import('emoji-picker-react'), { ssr: false });

type PreviewKind = 'image' | 'video' | 'file';

interface FilePreview {
  file: File;
  kind: PreviewKind;
  url: string;
}

export interface PostComposerProps {
  placeholder?: string;
  parentId?: number;
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
  const { data: joinedCommData } = useJoinedCommunities();
  const communities = joinedCommData?.data ?? [];

  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);
  const emojiPickerRef = useRef<HTMLDivElement | null>(null);
  const audienceMenuRef = useRef<HTMLDivElement | null>(null);
  const previewsRef = useRef<FilePreview[]>([]);

  const [content, setContent] = useState('');
  const [previews, setPreviews] = useState<FilePreview[]>([]);
  const [compressing, setCompressing] = useState(false);

  // Audience
  const [selectedCommunity, setSelectedCommunity] = useState<CommunityData | null>(null);
  const [showAudienceMenu, setShowAudienceMenu] = useState(false);

  // Emoji / GIF
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [showGifPicker, setShowGifPicker] = useState(false);
  const [gifFetching, setGifFetching] = useState(false);

  useEffect(() => {
    previewsRef.current = previews;
  }, [previews]);

  useEffect(() => {
    return () => {
      previewsRef.current.forEach((p) => URL.revokeObjectURL(p.url));
    };
  }, []);

  // Close dropdowns on outside click
  useEffect(() => {
    if (!showEmojiPicker) return;
    const handler = (e: MouseEvent) => {
      if (emojiPickerRef.current && !emojiPickerRef.current.contains(e.target as Node)) {
        setShowEmojiPicker(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [showEmojiPicker]);

  useEffect(() => {
    if (!showAudienceMenu) return;
    const handler = (e: MouseEvent) => {
      if (audienceMenuRef.current && !audienceMenuRef.current.contains(e.target as Node)) {
        setShowAudienceMenu(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [showAudienceMenu]);

  const avatarSrc = profile?.avatar ?? null;
  const selectedFiles = previews.map((p) => p.file);
  const normalizedContent = content.trim();

  // Regular thread mutation
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

  // Community thread mutation
  const communityMutation = useMutation<
    CommunityMessageResponse,
    Error,
    { communityId: number; content: string; files: File[] }
  >({
    mutationFn: async ({ communityId, content: c, files }) => {
      if (!token) throw new Error('Authentication token is required.');
      return createCommunityThread(communityId, { content: c, files }, token);
    },
    onSuccess: (_, { communityId }) => {
      void queryClient.invalidateQueries({ queryKey: ['community', communityId, 'threads'] });
      void queryClient.invalidateQueries({ queryKey: ['communities', 'threads', 'feed'] });
      void queryClient.invalidateQueries({ queryKey: ['threads', 'feed'] });
      void queryClient.invalidateQueries({ queryKey: ['threads'] });
    },
  });

  const isPending = mutation.isPending || communityMutation.isPending;
  const isError = mutation.isError || communityMutation.isError;
  const errorMessage = mutation.error?.message ?? communityMutation.error?.message ?? '';

  const canSubmit =
    token !== null &&
    !isPending &&
    !compressing &&
    !gifFetching &&
    (normalizedContent.length > 0 || selectedFiles.length > 0);

  async function handleFilesChange(event: ChangeEvent<HTMLInputElement>) {
    const incoming = Array.from(event.target.files ?? []);
    event.target.value = '';
    if (incoming.length === 0) return;
    setCompressing(true);
    try {
      const ready = await compressImages(incoming);
      const oversized = ready.filter(
        (f) => f.type.startsWith('image/') && f.size > IMAGE_MAX_BYTES.media,
      );
      if (oversized.length > 0) {
        toast.error(
          `${oversized.length} image${oversized.length > 1 ? 's' : ''} could not be compressed to the size limit and ${oversized.length > 1 ? 'were' : 'was'} removed.`,
        );
      }
      const accepted = ready.filter((f) => !oversized.includes(f));
      setPreviews((cur) => [
        ...cur,
        ...accepted.map((file) => ({
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

  function handleEmojiClick(emojiData: EmojiClickData) {
    const el = textareaRef.current;
    if (!el) {
      setContent((prev) => prev + emojiData.emoji);
    } else {
      const start = el.selectionStart ?? content.length;
      const end = el.selectionEnd ?? content.length;
      const next = content.slice(0, start) + emojiData.emoji + content.slice(end);
      setContent(next);
      requestAnimationFrame(() => {
        el.focus();
        const pos = start + emojiData.emoji.length;
        el.setSelectionRange(pos, pos);
      });
    }
    setShowEmojiPicker(false);
  }

  async function handleGifSelect(gifUrl: string) {
    setShowGifPicker(false);
    setGifFetching(true);
    try {
      const res = await fetch(gifUrl);
      const blob = await res.blob();
      const file = new File([blob], `gif-${Date.now()}.gif`, { type: blob.type || 'image/gif' });
      const url = URL.createObjectURL(file);
      setPreviews((cur) => [...cur, { file, kind: 'image', url }]);
    } catch {
      // silently ignore if CORS blocks
    } finally {
      setGifFetching(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canSubmit) return;
    try {
      if (!isReply && selectedCommunity) {
        await communityMutation.mutateAsync({
          communityId: selectedCommunity.community_id,
          content: normalizedContent,
          files: selectedFiles,
        });
      } else {
        await mutation.mutateAsync({
          content: normalizedContent,
          files: selectedFiles,
          isPrivate: false,
          ...(isReply && { parentId }),
        });
      }
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
        position: 'relative',
      }}
    >
      <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
        <Avatar username={user?.username ?? ''} avatarUrl={avatarSrc} size="md" />

        <div style={{ flex: 1, minWidth: 0 }}>

          {/* Audience selector — new posts only */}
          {!isReply && (
            <div style={{ position: 'relative', display: 'inline-block', marginBottom: 8 }} ref={audienceMenuRef}>
              <button
                type="button"
                onClick={() => setShowAudienceMenu((v) => !v)}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 6,
                  padding: '4px 10px',
                  borderRadius: 20,
                  border: '1px solid var(--accent)',
                  background: 'transparent',
                  color: 'var(--accent)',
                  fontSize: 13,
                  fontWeight: 600,
                  cursor: 'pointer',
                  fontFamily: 'inherit',
                }}
              >
                {selectedCommunity ? (
                  <>
                    <Users size={13} />
                    {selectedCommunity.name}
                  </>
                ) : (
                  <>
                    <Globe size={13} />
                    Everyone
                  </>
                )}
                <ChevronDown size={13} />
              </button>

              {showAudienceMenu && (
                <div
                  style={{
                    position: 'absolute',
                    top: 'calc(100% + 6px)',
                    left: 0,
                    minWidth: 240,
                    background: 'var(--bg-card)',
                    border: '1px solid var(--border)',
                    borderRadius: 16,
                    boxShadow: '0 4px 24px rgba(0,0,0,0.22)',
                    zIndex: 60,
                    overflow: 'hidden',
                  }}
                >
                  <div style={{ padding: '14px 16px 8px', fontSize: 16, fontWeight: 700, color: 'var(--text)' }}>
                    Choose audience
                  </div>

                  {/* Everyone option */}
                  <button
                    type="button"
                    onClick={() => { setSelectedCommunity(null); setShowAudienceMenu(false); }}
                    style={{
                      width: '100%',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 12,
                      padding: '10px 16px',
                      background: selectedCommunity === null ? 'var(--bg-subtle)' : 'transparent',
                      border: 'none',
                      cursor: 'pointer',
                      textAlign: 'left',
                      fontFamily: 'inherit',
                    }}
                  >
                    <div style={{
                      width: 40, height: 40, borderRadius: '50%',
                      background: 'var(--accent)',
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      flexShrink: 0,
                    }}>
                      <Globe size={20} color="#fff" />
                    </div>
                    <span style={{ flex: 1, fontSize: 15, fontWeight: 600, color: 'var(--text)' }}>Everyone</span>
                    {selectedCommunity === null && <Check size={16} style={{ color: 'var(--accent)', flexShrink: 0 }} />}
                  </button>

                  {communities.length > 0 && (
                    <>
                      <div style={{ padding: '8px 16px 4px', fontSize: 12, fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                        My Communities
                      </div>
                      {communities.map((c) => (
                        <button
                          key={c.community_id}
                          type="button"
                          onClick={() => { setSelectedCommunity(c); setShowAudienceMenu(false); }}
                          style={{
                            width: '100%',
                            display: 'flex',
                            alignItems: 'center',
                            gap: 12,
                            padding: '10px 16px',
                            background: selectedCommunity?.community_id === c.community_id ? 'var(--bg-subtle)' : 'transparent',
                            border: 'none',
                            cursor: 'pointer',
                            textAlign: 'left',
                            fontFamily: 'inherit',
                          }}
                        >
                          <div style={{ position: 'relative', width: 56, height: 40, borderRadius: 8, overflow: 'hidden', flexShrink: 0, background: 'var(--bg-subtle)' }}>
                            {c.banner_url ? (
                              // eslint-disable-next-line @next/next/no-img-element
                              <img src={c.banner_url} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
                            ) : (
                              <div style={{ width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                                <Users size={18} style={{ color: 'var(--text-muted)' }} />
                              </div>
                            )}
                            {c.avatar_url && (
                              // eslint-disable-next-line @next/next/no-img-element
                              <img src={c.avatar_url} alt={c.name} style={{ position: 'absolute', bottom: 2, right: 2, width: 18, height: 18, borderRadius: '50%', objectFit: 'cover', border: '1px solid var(--bg-card)' }} />
                            )}
                          </div>
                          <div style={{ flex: 1, minWidth: 0 }}>
                            <div style={{ fontSize: 15, fontWeight: 600, color: 'var(--text)', overflow: 'hidden', whiteSpace: 'nowrap', textOverflow: 'ellipsis' }}>{c.name}</div>
                            <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
                              {c.member_count >= 1000
                                ? `${(c.member_count / 1000).toFixed(1)}K`
                                : c.member_count} Members
                            </div>
                          </div>
                          {selectedCommunity?.community_id === c.community_id && (
                            <Check size={16} style={{ color: 'var(--accent)', flexShrink: 0 }} />
                          )}
                        </button>
                      ))}
                    </>
                  )}
                  <div style={{ height: 8 }} />
                </div>
              )}
            </div>
          )}

          <textarea
            ref={textareaRef}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder={placeholder}
            rows={isReply ? 2 : 2}
            style={{
              width: '100%',
              resize: 'vertical',
              minHeight: isReply ? 48 : 48,
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
                    maxHeight: 220,
                  }}
                >
                  {preview.kind === 'video' ? (
                    <video
                      src={preview.url}
                      muted
                      playsInline
                      controls
                      style={{ width: '100%', height: '100%', display: 'block', objectFit: 'cover' }}
                    />
                  ) : preview.kind === 'image' ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={preview.url}
                      alt={preview.file.name}
                      style={{ width: '100%', height: '100%', display: 'block', objectFit: 'cover' }}
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

          {isError && (
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
              {errorMessage}
            </div>
          )}

          {/* Emoji picker popover */}
          {showEmojiPicker && (
            <div
              ref={emojiPickerRef}
              style={{ position: 'relative', zIndex: 50, marginBottom: 8 }}
            >
              <EmojiPicker onEmojiClick={handleEmojiClick} lazyLoadEmojis height={360} width="100%" />
            </div>
          )}

          {/* GIF picker popover */}
          {showGifPicker && (
            <div style={{ position: 'relative', marginBottom: 8 }}>
              <GifPicker
                onSelect={(url) => void handleGifSelect(url)}
                onClose={() => setShowGifPicker(false)}
              />
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
            <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
              {/* Image upload */}
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
                title="Add image or video"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 36,
                  height: 36,
                  borderRadius: 8,
                  color: compressing ? 'var(--accent)' : 'var(--text-secondary)',
                  cursor: isPending || compressing ? 'not-allowed' : 'pointer',
                  opacity: isPending || compressing ? 0.6 : 1,
                  pointerEvents: compressing ? 'none' : undefined,
                }}
              >
                {compressing ? <Loader2 size={18} style={{ animation: 'spin 1s linear infinite' }} /> : <ImagePlus size={18} />}
              </label>

              {/* GIF button */}
              <button
                type="button"
                title="Add GIF"
                onClick={() => { setShowGifPicker((v) => !v); setShowEmojiPicker(false); }}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 36,
                  height: 36,
                  borderRadius: 8,
                  background: 'none',
                  border: 'none',
                  color: showGifPicker ? 'var(--accent)' : 'var(--text-secondary)',
                  cursor: gifFetching ? 'not-allowed' : 'pointer',
                  opacity: gifFetching ? 0.6 : 1,
                  transition: 'color 0.15s',
                }}
              >
                {gifFetching ? <Loader2 size={18} style={{ animation: 'spin 1s linear infinite' }} /> : <Film size={18} />}
              </button>

              {/* Emoji button */}
              <button
                type="button"
                title="Add emoji"
                onClick={() => { setShowEmojiPicker((v) => !v); setShowGifPicker(false); }}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 36,
                  height: 36,
                  borderRadius: 8,
                  background: 'none',
                  border: 'none',
                  color: showEmojiPicker ? 'var(--accent)' : 'var(--text-secondary)',
                  cursor: 'pointer',
                  transition: 'color 0.15s',
                }}
              >
                <Smile size={18} />
              </button>

              {compressing && (
                <span style={{ color: 'var(--accent)', fontSize: 13, marginLeft: 4 }}>Compressing…</span>
              )}
              {!compressing && !gifFetching && selectedFiles.length > 0 && (
                <span style={{ color: 'var(--text-tertiary)', fontSize: 13, marginLeft: 4 }}>
                  {selectedFiles.length} file{selectedFiles.length > 1 ? 's' : ''}
                </span>
              )}
              {gifFetching && (
                <span style={{ color: 'var(--accent)', fontSize: 13, marginLeft: 4 }}>Adding GIF…</span>
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
