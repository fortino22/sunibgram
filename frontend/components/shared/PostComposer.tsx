'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { ImagePlus, LoaderCircle, X, Loader2, Globe, ChevronDown, Smile, Film, Users, Check, BarChart3, Plus, UserCheck } from 'lucide-react';
import {
  type ChangeEvent,
  type FormEvent,
  useEffect,
  useRef,
  useState,
} from 'react';
import dynamic from 'next/dynamic';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useFieldArray, useForm, useWatch } from 'react-hook-form';
import { z } from 'zod';
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
import { Input } from '@/components/ui/input';
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
  onSuccess?: () => void;
}

interface MutateContext {
  previousChain: [readonly unknown[], ThreadDetailData | undefined][];
  savedContent: string;
}

const MAX_POLL_OPTIONS = 4;
const MAX_POLL_LABEL_LENGTH = 25;
const MIN_POLL_DURATION_SECONDS = 5 * 60;
const MAX_POLL_DURATION_SECONDS = 7 * 24 * 60 * 60;

const pollSchema = z.object({
  enabled: z.boolean(),
  options: z.array(
    z.object({
      label: z.string().max(MAX_POLL_LABEL_LENGTH, `Choice must be ${MAX_POLL_LABEL_LENGTH} characters or less.`),
      image: z.custom<File | undefined>((value) => value === undefined || value instanceof File),
    })
  ).min(2).max(MAX_POLL_OPTIONS),
  days: z.number().int().min(0).max(7),
  hours: z.number().int().min(0).max(23),
  minutes: z.number().int().min(0).max(59),
}).superRefine((value, ctx) => {
  if (!value.enabled) {
    return;
  }

  value.options.forEach((option, index) => {
    const label = option.label.trim();
    if (label.length === 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Choice is required.',
        path: ['options', index, 'label'],
      });
    }
  });

  const durationSeconds =
    value.days * 24 * 60 * 60 +
    value.hours * 60 * 60 +
    value.minutes * 60;

  if (durationSeconds < MIN_POLL_DURATION_SECONDS || durationSeconds > MAX_POLL_DURATION_SECONDS) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'Poll length must be between 5 minutes and 7 days.',
      path: ['minutes'],
    });
  }

  const imageCount = value.options.filter((option) => option.image instanceof File).length;
  if (imageCount > 0 && imageCount !== value.options.length) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'Add images to every choice or remove them all.',
      path: ['options'],
    });
  }
});

type PollFormValues = z.infer<typeof pollSchema>;

const defaultPollValues: PollFormValues = {
  enabled: false,
  options: [
    { label: '', image: undefined },
    { label: '', image: undefined },
  ],
  days: 1,
  hours: 0,
  minutes: 0,
};

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

function countRunes(value: string): number {
  return Array.from(value).length;
}

function buildPollDuration(values: Pick<PollFormValues, 'days' | 'hours' | 'minutes'>): number {
  return values.days * 24 * 60 * 60 + values.hours * 60 * 60 + values.minutes * 60;
}

export function PostComposer({ placeholder = 'What is happening?!', parentId, formId, onSuccess }: PostComposerProps) {
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
  const replyPermMenuRef = useRef<HTMLDivElement | null>(null);
  const previewsRef = useRef<FilePreview[]>([]);
  const pollOptionInputRefs = useRef<Record<string, HTMLInputElement | null>>({});
  const pollImageUrlsRef = useRef<Record<string, string>>({});

  const [content, setContent] = useState('');
  const [previews, setPreviews] = useState<FilePreview[]>([]);
  const [compressing, setCompressing] = useState(false);
  const [pollImageUrls, setPollImageUrls] = useState<Record<string, string>>({});

  // Audience
  const [selectedCommunity, setSelectedCommunity] = useState<CommunityData | null>(null);
  const [showAudienceMenu, setShowAudienceMenu] = useState(false);

  // Reply permission
  const [replyPermission, setReplyPermission] = useState<'EVERYONE' | 'FOLLOWING' | 'MUTUAL'>('EVERYONE');
  const [showReplyPermMenu, setShowReplyPermMenu] = useState(false);

  // Emoji / GIF
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [showGifPicker, setShowGifPicker] = useState(false);
  const [gifFetching, setGifFetching] = useState(false);

  const pollForm = useForm<PollFormValues>({
    resolver: zodResolver(pollSchema),
    defaultValues: defaultPollValues,
    mode: 'onChange',
  });
  const pollFields = useFieldArray({
    control: pollForm.control,
    name: 'options',
  });
  const pollEnabled = useWatch({
    control: pollForm.control,
    name: 'enabled',
    defaultValue: false,
  });
  const pollOptions = useWatch({
    control: pollForm.control,
    name: 'options',
    defaultValue: defaultPollValues.options,
  });
  const pollDays = useWatch({
    control: pollForm.control,
    name: 'days',
    defaultValue: defaultPollValues.days,
  });
  const pollHours = useWatch({
    control: pollForm.control,
    name: 'hours',
    defaultValue: defaultPollValues.hours,
  });
  const pollMinutes = useWatch({
    control: pollForm.control,
    name: 'minutes',
    defaultValue: defaultPollValues.minutes,
  });

  useEffect(() => {
    previewsRef.current = previews;
  }, [previews]);

  useEffect(() => {
    pollImageUrlsRef.current = pollImageUrls;
  }, [pollImageUrls]);

  useEffect(() => {
    return () => {
      previewsRef.current.forEach((p) => URL.revokeObjectURL(p.url));
      Object.values(pollImageUrlsRef.current).forEach((url) => URL.revokeObjectURL(url));
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

  useEffect(() => {
    if (!showReplyPermMenu) return;
    const handler = (e: MouseEvent) => {
      if (replyPermMenuRef.current && !replyPermMenuRef.current.contains(e.target as Node)) {
        setShowReplyPermMenu(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [showReplyPermMenu]);

  const avatarSrc = profile?.avatar ?? null;
  const selectedFiles = previews.map((p) => p.file);
  const normalizedContent = content.trim();
  const pollDurationSeconds = buildPollDuration({ days: pollDays, hours: pollHours, minutes: pollMinutes });
  const rawOptionsError = pollForm.formState.errors.options;
  const pollOptionsError =
    !Array.isArray(rawOptionsError) && typeof rawOptionsError?.message === 'string'
      ? rawOptionsError.message
      : undefined;

  function revokePollImage(key: string) {
    setPollImageUrls((current) => {
      const next = { ...current };
      const url = next[key];
      if (url) {
        URL.revokeObjectURL(url);
        delete next[key];
      }
      return next;
    });
  }

  function resetPollComposer() {
    Object.values(pollImageUrlsRef.current).forEach((url) => URL.revokeObjectURL(url));
    pollImageUrlsRef.current = {};
    setPollImageUrls({});
    Object.values(pollOptionInputRefs.current).forEach((input) => {
      if (input) {
        input.value = '';
      }
    });
    pollForm.reset(defaultPollValues);
  }

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
              name: profile?.name ?? '',
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
    (normalizedContent.length > 0 || selectedFiles.length > 0 || (pollEnabled && !isReply));

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

  function handleSelectAudience(community: CommunityData | null) {
    if (community && pollEnabled) {
      toast.error('Polls can only be posted to Everyone with the current backend.');
      return;
    }

    setSelectedCommunity(community);
    setShowAudienceMenu(false);
  }

  function handleAddPoll() {
    if (isReply) {
      return;
    }

    if (selectedCommunity) {
      toast.error('Polls can only be posted to Everyone with the current backend.');
      return;
    }

    pollForm.setValue('enabled', true, { shouldDirty: true, shouldValidate: true });
  }

  function handleRemovePoll() {
    resetPollComposer();
  }

  async function handlePollImageChange(index: number, fieldId: string, event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }

    if (file.size > IMAGE_MAX_BYTES.raw) {
      toast.error('Poll option images must be 5 MB or smaller.');
      event.target.value = '';
      return;
    }

    revokePollImage(fieldId);
    const url = URL.createObjectURL(file);
    setPollImageUrls((current) => ({ ...current, [fieldId]: url }));
    pollForm.setValue(`options.${index}.image`, file, {
      shouldDirty: true,
      shouldValidate: true,
    });
    event.target.value = '';
  }

  function handleRemovePollImage(index: number, fieldId: string) {
    revokePollImage(fieldId);
    pollForm.setValue(`options.${index}.image`, undefined, {
      shouldDirty: true,
      shouldValidate: true,
    });

    const input = pollOptionInputRefs.current[fieldId];
    if (input) {
      input.value = '';
    }
  }

  function handleRemovePollOption(index: number) {
    const field = pollFields.fields[index];
    if (!field) {
      return;
    }

    revokePollImage(field.id);
    pollFields.remove(index);
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

    if (pollEnabled) {
      const valid = await pollForm.trigger();
      if (!valid) {
        return;
      }
    }

    try {
      if (!isReply && selectedCommunity) {
        await communityMutation.mutateAsync({
          communityId: selectedCommunity.community_id,
          content: normalizedContent,
          files: selectedFiles,
        });
      } else {
        const pollValues = pollForm.getValues();
        await mutation.mutateAsync({
          content: normalizedContent,
          files: selectedFiles,
          isPrivate: false,
          ...(pollEnabled && {
            poll: {
              question: '',
              options: pollValues.options.map((option) => ({
                label: option.label.trim(),
                image: option.image,
              })),
              durationSeconds: buildPollDuration(pollValues),
            },
          }),
          ...(isReply && { parentId }),
          ...(!isReply && { replyPermission }),
        });
      }
      if (!isReply) setContent('');
      previews.forEach((p) => URL.revokeObjectURL(p.url));
      setPreviews([]);
      if (!isReply) {
        resetPollComposer();
        setReplyPermission('EVERYONE');
      }
      if (fileInputRef.current) fileInputRef.current.value = '';
      onSuccess?.();
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
                    onClick={() => handleSelectAudience(null)}
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
                          onClick={() => handleSelectAudience(c)}
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

          {pollEnabled && !isReply && (
            <div
              style={{
                marginBottom: 12,
                border: '1px solid var(--border)',
                borderRadius: 16,
                overflow: 'hidden',
                background: 'var(--bg-card)',
              }}
            >
              <div style={{ padding: '14px 14px 0', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
                <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--text)' }}>Poll</div>
                <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>Up to {MAX_POLL_OPTIONS} choices</div>
              </div>

              <div style={{ padding: '10px 14px 14px', display: 'flex', flexDirection: 'column', gap: 10 }}>
                {pollFields.fields.map((field, index) => {
                  const optionValue = pollOptions?.[index];
                  const labelValue = optionValue?.label ?? '';
                  const imageUrl = pollImageUrls[field.id];
                  const optionError = pollForm.formState.errors.options?.[index];

                  return (
                    <div key={field.id}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <input
                          ref={(node) => {
                            pollOptionInputRefs.current[field.id] = node;
                          }}
                          type="file"
                          accept="image/*"
                          onChange={(event) => void handlePollImageChange(index, field.id, event)}
                          style={{ display: 'none' }}
                        />
                        <button
                          type="button"
                          onClick={() => pollOptionInputRefs.current[field.id]?.click()}
                          aria-label={`Add image to choice ${index + 1}`}
                          style={{
                            width: 42,
                            height: 42,
                            borderRadius: 12,
                            border: '1px solid var(--border)',
                            background: imageUrl ? 'transparent' : 'var(--bg-subtle)',
                            color: imageUrl ? 'transparent' : 'var(--text-secondary)',
                            display: 'inline-flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            overflow: 'hidden',
                            cursor: 'pointer',
                            flexShrink: 0,
                            position: 'relative',
                          }}
                        >
                          {imageUrl ? (
                            <>
                              {/* eslint-disable-next-line @next/next/no-img-element */}
                              <img
                                src={imageUrl}
                                alt=""
                                style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
                              />
                              <span
                                style={{
                                  position: 'absolute',
                                  right: 4,
                                  bottom: 4,
                                  width: 18,
                                  height: 18,
                                  borderRadius: '50%',
                                  background: 'rgba(15, 23, 42, 0.8)',
                                  color: '#fff',
                                  display: 'inline-flex',
                                  alignItems: 'center',
                                  justifyContent: 'center',
                                }}
                              >
                                <ImagePlus size={11} />
                              </span>
                            </>
                          ) : (
                            <ImagePlus size={16} />
                          )}
                        </button>

                        <div style={{ flex: 1, minWidth: 0 }}>
                          <Input
                            placeholder={`Choice ${index + 1}`}
                            maxLength={MAX_POLL_LABEL_LENGTH}
                            {...pollForm.register(`options.${index}.label`)}
                            style={{
                              height: 42,
                              borderRadius: 12,
                              background: 'var(--bg-card)',
                            }}
                          />
                        </div>

                        <div style={{ minWidth: 54, textAlign: 'right', fontSize: 12, color: 'var(--text-tertiary)' }}>
                          {countRunes(labelValue)} / {MAX_POLL_LABEL_LENGTH}
                        </div>

                        {pollFields.fields.length > 2 && (
                          <button
                            type="button"
                            onClick={() => handleRemovePollOption(index)}
                            aria-label={`Remove choice ${index + 1}`}
                            style={{
                              width: 30,
                              height: 30,
                              borderRadius: '50%',
                              border: 'none',
                              background: 'transparent',
                              color: 'var(--text-tertiary)',
                              display: 'inline-flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              cursor: 'pointer',
                              flexShrink: 0,
                            }}
                          >
                            <X size={14} />
                          </button>
                        )}
                      </div>

                      {(optionError?.label?.message || imageUrl) && (
                        <div style={{ marginTop: 6, display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
                          <div style={{ fontSize: 12, color: 'var(--coral)' }}>
                            {optionError?.label?.message}
                          </div>
                          {imageUrl && (
                            <button
                              type="button"
                              onClick={() => handleRemovePollImage(index, field.id)}
                              style={{
                                border: 'none',
                                background: 'transparent',
                                color: 'var(--text-tertiary)',
                                fontSize: 12,
                                cursor: 'pointer',
                                padding: 0,
                              }}
                            >
                              Remove image
                            </button>
                          )}
                        </div>
                      )}
                    </div>
                  );
                })}

                {pollFields.fields.length < MAX_POLL_OPTIONS && (
                  <button
                    type="button"
                    onClick={() => pollFields.append({ label: '', image: undefined })}
                    style={{
                      marginTop: 2,
                      width: 34,
                      height: 34,
                      borderRadius: '50%',
                      border: '1px solid var(--border)',
                      background: 'var(--bg-subtle)',
                      color: 'var(--text)',
                      display: 'inline-flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      cursor: 'pointer',
                    }}
                  >
                    <Plus size={16} />
                  </button>
                )}

                {pollOptionsError && (
                  <div style={{ fontSize: 12, color: 'var(--coral)' }}>
                    {pollOptionsError}
                  </div>
                )}

                <div
                  style={{
                    marginTop: 4,
                    paddingTop: 14,
                    borderTop: '1px solid var(--border)',
                  }}
                >
                  <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--text)', marginBottom: 10 }}>
                    Poll length
                  </div>
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: 10 }}>
                    <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                      <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>Days</span>
                      <select
                        value={pollDays}
                        onChange={(e) => {
                          const val = Number(e.target.value);
                          pollForm.setValue('days', val, { shouldDirty: true, shouldValidate: true });
                          // If now only minutes selectable and current value is below minimum, clamp to 5
                          if (val === 0 && pollHours === 0 && pollMinutes < 5) {
                            pollForm.setValue('minutes', 5, { shouldDirty: true, shouldValidate: true });
                          }
                        }}
                        style={{
                          height: 42,
                          borderRadius: 12,
                          border: '1px solid var(--border)',
                          background: 'var(--bg-card)',
                          color: 'var(--text)',
                          padding: '0 12px',
                          outline: 'none',
                          fontSize: 14,
                        }}
                      >
                        {Array.from({ length: 8 }, (_, v) => (
                          <option key={v} value={v}>{v}</option>
                        ))}
                      </select>
                    </label>

                    <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                      <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>Hours</span>
                      <select
                        value={pollHours}
                        onChange={(e) => {
                          const val = Number(e.target.value);
                          pollForm.setValue('hours', val, { shouldDirty: true, shouldValidate: true });
                          // If now only minutes selectable and current value is below minimum, clamp to 5
                          if (val === 0 && pollDays === 0 && pollMinutes < 5) {
                            pollForm.setValue('minutes', 5, { shouldDirty: true, shouldValidate: true });
                          }
                        }}
                        style={{
                          height: 42,
                          borderRadius: 12,
                          border: '1px solid var(--border)',
                          background: 'var(--bg-card)',
                          color: 'var(--text)',
                          padding: '0 12px',
                          outline: 'none',
                          fontSize: 14,
                        }}
                      >
                        {Array.from({ length: 24 }, (_, v) => (
                          <option key={v} value={v}>{v}</option>
                        ))}
                      </select>
                    </label>

                    <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                      <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>Minutes</span>
                      <select
                        value={pollMinutes < ((pollDays > 0 || pollHours > 0) ? 0 : 5) ? ((pollDays > 0 || pollHours > 0) ? 0 : 5) : pollMinutes}
                        onChange={(e) => {
                          const val = Number(e.target.value);
                          pollForm.setValue('minutes', val, { shouldDirty: true, shouldValidate: true });
                        }}
                        style={{
                          height: 42,
                          borderRadius: 12,
                          border: '1px solid var(--border)',
                          background: 'var(--bg-card)',
                          color: 'var(--text)',
                          padding: '0 12px',
                          outline: 'none',
                          fontSize: 14,
                        }}
                      >
                        {Array.from(
                          { length: (pollDays > 0 || pollHours > 0) ? 60 : 55 },
                          (_, i) => i + ((pollDays > 0 || pollHours > 0) ? 0 : 5)
                        ).map((v) => (
                          <option key={v} value={v}>{v}</option>
                        ))}
                      </select>
                    </label>
                  </div>
                  {pollForm.formState.errors.minutes?.message && (
                    <div style={{ marginTop: 8, fontSize: 12, color: 'var(--coral)' }}>
                      {pollForm.formState.errors.minutes.message}
                    </div>
                  )}
                  <div style={{ marginTop: 8, fontSize: 12, color: 'var(--text-tertiary)' }}>
                    {`${Math.floor(pollDurationSeconds / 86400)}d ${Math.floor((pollDurationSeconds % 86400) / 3600)}h ${Math.floor((pollDurationSeconds % 3600) / 60)}m`}
                  </div>
                </div>

                <button
                  type="button"
                  onClick={handleRemovePoll}
                  style={{
                    marginTop: 4,
                    border: 'none',
                    background: 'transparent',
                    color: 'var(--coral)',
                    fontSize: 14,
                    fontWeight: 700,
                    padding: 0,
                    textAlign: 'left',
                    cursor: 'pointer',
                  }}
                >
                  Remove poll
                </button>
              </div>
            </div>
          )}

          {/* Reply permission indicator — new posts only */}
          {!isReply && (
            <div style={{ position: 'relative', marginBottom: 10 }} ref={replyPermMenuRef}>
              <button
                type="button"
                onClick={() => setShowReplyPermMenu((v) => !v)}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 5,
                  padding: '2px 6px 2px 2px',
                  borderRadius: 20,
                  border: 'none',
                  background: 'transparent',
                  color: 'var(--accent)',
                  fontSize: 13,
                  fontWeight: 600,
                  cursor: 'pointer',
                  fontFamily: 'inherit',
                }}
              >
                {replyPermission === 'EVERYONE' && <Globe size={14} />}
                {replyPermission === 'FOLLOWING' && <UserCheck size={14} />}
                {replyPermission === 'MUTUAL' && <Users size={14} />}
                {replyPermission === 'EVERYONE' ? 'Everyone' : replyPermission === 'FOLLOWING' ? 'Followers' : 'Mutual followers'} can reply
              </button>

              {showReplyPermMenu && (
                <div
                  style={{
                    position: 'absolute',
                    bottom: 'calc(100% + 6px)',
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
                    Who can reply?
                  </div>
                  {(
                    [
                      { value: 'EVERYONE', label: 'Everyone', Icon: Globe },
                      { value: 'FOLLOWING', label: 'Followers', Icon: UserCheck },
                      { value: 'MUTUAL', label: 'Mutual followers', Icon: Users },
                    ] as const
                  ).map(({ value, label, Icon }) => (
                    <button
                      key={value}
                      type="button"
                      onClick={() => { setReplyPermission(value); setShowReplyPermMenu(false); }}
                      style={{
                        width: '100%',
                        display: 'flex',
                        alignItems: 'center',
                        gap: 12,
                        padding: '10px 16px',
                        background: replyPermission === value ? 'var(--bg-subtle)' : 'transparent',
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
                        <Icon size={20} color="#fff" />
                      </div>
                      <span style={{ flex: 1, fontSize: 15, fontWeight: 600, color: 'var(--text)' }}>{label}</span>
                      {replyPermission === value && <Check size={16} style={{ color: 'var(--accent)', flexShrink: 0 }} />}
                    </button>
                  ))}
                  <div style={{ height: 8 }} />
                </div>
              )}
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

              {!isReply && (
                <button
                  type="button"
                  title={selectedCommunity ? 'Polls are only available for posts to Everyone' : pollEnabled ? 'Poll added' : 'Add poll'}
                  onClick={handleAddPoll}
                  disabled={pollEnabled || selectedCommunity !== null}
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    width: 36,
                    height: 36,
                    borderRadius: 8,
                    background: 'none',
                    border: 'none',
                    color: pollEnabled ? 'var(--accent)' : 'var(--text-secondary)',
                    cursor: pollEnabled || selectedCommunity !== null ? 'not-allowed' : 'pointer',
                    opacity: pollEnabled || selectedCommunity !== null ? 0.55 : 1,
                    transition: 'color 0.15s',
                  }}
                >
                  <BarChart3 size={18} />
                </button>
              )}

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
