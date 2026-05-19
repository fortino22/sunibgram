'use client';

import { useState, useRef } from 'react';
import { toast } from 'sonner';
import { useCategories } from '@/hooks/useCommunities';
import { useCreateCommunity } from '@/hooks/useCommunityMutations';
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
  const [rules, setRules] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [visibility, setVisibility] = useState<CommunityVisibility>('PUBLIC');
  const [bannerFile, setBannerFile] = useState<File | null>(null);
  const [bannerPreview, setBannerPreview] = useState<string | null>(null);
  const bannerRef = useRef<HTMLInputElement>(null);

  function handleImageChange(
    file: File | null,
    setFile: (f: File | null) => void,
    setPreview: (s: string | null) => void
  ) {
    if (!file) return;
    setFile(file);
    const url = URL.createObjectURL(file);
    setPreview(url);
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
    try {
      const community = await mutateAsync({
        name: name.trim(),
        description: description.trim(),
        rules: rules.trim(),
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
            height: 100,
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
            onChange={(e) => handleImageChange(e.target.files?.[0] ?? null, setBannerFile, setBannerPreview)} />
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
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>
              Description
            </label>
            <textarea
              style={{ ...inputStyle, minHeight: 72, resize: 'vertical' }}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="What is this community about?"
            />
          </div>

          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>
              Rules
            </label>
            <textarea
              style={{ ...inputStyle, minHeight: 72, resize: 'vertical' }}
              value={rules}
              onChange={(e) => setRules(e.target.value)}
              placeholder="Community rules and guidelines"
            />
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
  );
}
