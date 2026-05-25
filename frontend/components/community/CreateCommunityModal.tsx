'use client';

import { useState, useRef } from 'react';
import { toast } from 'sonner';
import { useCategories } from '@/hooks/useCommunities';
import { useCreateCommunity } from '@/hooks/useCommunityMutations';
import { IMAGE_MAX_BYTES } from '@/lib/compressImage';

import { ImageCropModal } from '@/components/profile/ImageCropModal';
import type { CommunityVisibility } from '@/lib/services/communities';

interface CreateCommunityModalProps {
  onClose: () => void;
  onSuccess: (communityId: number) => void;
}

export function CreateCommunityModal({ onClose, onSuccess }: CreateCommunityModalProps) {
  const { data: categoriesData } = useCategories();
  const categories = categoriesData?.data ?? [];
  const { mutateAsync, isPending, error } = useCreateCommunity();

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [rules, setRules] = useState<string[]>([]);
  const [categoryId, setCategoryId] = useState('');
  const [visibility, setVisibility] = useState<CommunityVisibility>('PUBLIC');
  const [bannerFile, setBannerFile] = useState<File | null>(null);
  const [bannerPreview, setBannerPreview] = useState<string | null>(null);
  const [pendingCrop, setPendingCrop] = useState<{ src: string; filename: string } | null>(null);
  const bannerRef = useRef<HTMLInputElement>(null);

  function handleBannerPick(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setPendingCrop({ src: URL.createObjectURL(file), filename: file.name });
    e.target.value = '';
  }

  function onCropComplete(file: File, preview: string) {
    setBannerFile(file);
    setBannerPreview(preview);
    setPendingCrop(null);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Please enter a community name.');
      return;
    }
    if (!categoryId) {
      toast.error('Please select a category.');
      return;
    }
    if (!bannerFile) {
      toast.error('Please add a banner image.');
      return;
    }
    if (description.length > 2000) {
      toast.error('Description cannot exceed 2000 characters.');
      return;
    }
    const badRuleIdx = rules.findIndex((r) => r.trim().length > 500);
    if (badRuleIdx !== -1) {
      toast.error(`Rule ${badRuleIdx + 1} exceeds 500 characters.`);
      return;
    }
    try {
      const community = await mutateAsync({
        name: name.trim(),
        description: description.trim(),
        rules: rules.map((r) => r.trim()).filter(Boolean),
        categoryId,
        visibility,
        banner: bannerFile,
      });
      onSuccess(community.community_id);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Something went wrong.');
    }
  }

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '10px 12px',
    borderRadius: 8,
    border: '1px solid var(--border)',
    background: 'var(--bg-subtle)',
    color: 'var(--text)',
    fontSize: 14,
    boxSizing: 'border-box',
  };

  return (
    <>
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 100,
        padding: 16,
      }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          borderRadius: 16,
          width: '100%',
          maxWidth: 480,
          maxHeight: '90vh',
          overflowY: 'auto',
          border: '1px solid var(--border)',
        }}
      >
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '16px 20px',
            borderBottom: '1px solid var(--border)',
            position: 'sticky',
            top: 0,
            background: 'var(--bg-card)',
            zIndex: 1,
          }}
        >
          <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)' }}>
            Create Community
          </span>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--text-tertiary)',
              fontSize: 20,
              cursor: 'pointer',
              lineHeight: 1,
            }}
          >
            ✕
          </button>
        </div>

        {/* Banner preview */}
        <div
          onClick={() => bannerRef.current?.click()}
          style={{
            aspectRatio: '2.31',
            width: '100%',
            background: bannerPreview
              ? `url(${bannerPreview}) center/cover no-repeat`
              : 'var(--bg-subtle)',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {!bannerPreview && (
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Click to add banner</span>
          )}
          <input ref={bannerRef} type="file" accept="image/*" style={{ display: 'none' }}
            onChange={handleBannerPick} />
        </div>

        <form onSubmit={handleSubmit} style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>
              Name *
            </label>
            <input
              style={inputStyle}
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Community name"
              maxLength={100}
              required
            />
          </div>

          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>
              Category *
            </label>
            <select
              style={{ ...inputStyle, appearance: 'none' }}
              value={categoryId}
              onChange={(e) => setCategoryId(e.target.value)}
              required
            >
              <option value="">Select category</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </div>

          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>
              Visibility
            </label>
            <div style={{ display: 'flex', gap: 8 }}>
              {(['PUBLIC', 'PRIVATE'] as CommunityVisibility[]).map((v) => (
                <button
                  key={v}
                  type="button"
                  onClick={() => setVisibility(v)}
                  style={{
                    flex: 1,
                    padding: '8px',
                    borderRadius: 8,
                    border: `1px solid ${visibility === v ? 'var(--text)' : 'var(--border)'}`,
                    background: visibility === v ? 'var(--text)' : 'transparent',
                    color: visibility === v ? 'var(--bg)' : 'var(--text)',
                    fontSize: 13,
                    fontWeight: 600,
                    cursor: 'pointer',
                  }}
                >
                  {v === 'PUBLIC' ? 'Public' : 'Private'}
                </button>
              ))}
            </div>
          </div>

          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
              <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}>
                Description
              </label>
              <span style={{ fontSize: 11, color: description.length > 2000 ? 'var(--error, #e53e3e)' : 'var(--text-tertiary)' }}>
                {description.length}/2000
              </span>
            </div>
            <textarea
              style={{ ...inputStyle, minHeight: 72, resize: 'vertical' }}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="What is this community about?"
              maxLength={2000}
            />
          </div>

          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
              <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}>
                Rules
              </label>
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>{rules.length}/10</span>
            </div>
            {rules.map((rule, i) => (
              <div key={i} style={{ marginBottom: 6 }}>
                <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                  <input
                    style={{ ...inputStyle, flex: 1 }}
                    value={rule}
                    onChange={(e) => {
                      const next = [...rules];
                      next[i] = e.target.value;
                      setRules(next);
                    }}
                    placeholder={`Rule ${i + 1}`}
                    maxLength={500}
                  />
                  <button
                    type="button"
                    onClick={() => setRules(rules.filter((_, j) => j !== i))}
                    style={{
                      background: 'none', border: '1px solid var(--border)', borderRadius: 6,
                      color: 'var(--text-tertiary)', cursor: 'pointer', padding: '8px 10px', fontSize: 14,
                      flexShrink: 0,
                    }}
                  >
                    ✕
                  </button>
                </div>
                {rule.length > 480 && (
                  <div style={{ fontSize: 11, color: rule.length > 500 ? 'var(--error, #e53e3e)' : 'var(--text-tertiary)', marginTop: 2 }}>
                    {rule.length}/500
                  </div>
                )}
              </div>
            ))}
            {rules.length < 10 && (
              <button
                type="button"
                onClick={() => setRules([...rules, ''])}
                style={{
                  fontSize: 13, color: 'var(--text-secondary)', background: 'none',
                  border: '1px dashed var(--border)', borderRadius: 8, padding: '7px 14px',
                  cursor: 'pointer', width: '100%', marginTop: 2,
                }}
              >
                + Add rule
              </button>
            )}
          </div>

          {error && (
            <div style={{ fontSize: 13, color: 'var(--error, #e53e3e)' }}>
              {error instanceof Error ? error.message : 'Something went wrong'}
            </div>
          )}

          <button
            type="submit"
            disabled={isPending}
            style={{
              width: '100%',
              padding: '12px',
              borderRadius: 999,
              border: 'none',
              background: 'var(--text)',
              color: 'var(--bg)',
              fontSize: 15,
              fontWeight: 700,
              cursor: isPending ? 'default' : 'pointer',
              opacity: isPending ? 0.5 : 1,
            }}
          >
            {isPending ? 'Creating…' : 'Create Community'}
          </button>
        </form>
      </div>
    </div>

    {pendingCrop && (
      <ImageCropModal
        imageSrc={pendingCrop.src}
        filename={pendingCrop.filename}
        aspect={2.31}
        maxSizeBytes={IMAGE_MAX_BYTES.banner}
        onComplete={onCropComplete}
        onClose={() => setPendingCrop(null)}
      />
    )}
    </>
  );
}
