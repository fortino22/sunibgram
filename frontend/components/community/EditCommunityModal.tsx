'use client';

import { useState, useRef, useMemo } from 'react';
import { useCategories } from '@/hooks/useCommunities';
import { useUpdateCommunity } from '@/hooks/useCommunityMutations';
import { ImageCropModal } from '@/components/profile/ImageCropModal';
import type { CommunityData, CommunityVisibility } from '@/lib/services/communities';

type PendingCrop = { src: string; filename: string };

interface EditCommunityModalProps {
  community: CommunityData;
  onClose: () => void;
  onSuccess: () => void;
}

export function EditCommunityModal({ community, onClose, onSuccess }: EditCommunityModalProps) {
  const { data: categoriesData } = useCategories();
  const categories = categoriesData?.data ?? [];
  const categoryId = useMemo(
    () => categories.find((c) => c.name === community.category_name)?.id,
    [categories, community.category_name]
  );
  const { mutateAsync, isPending, error } = useUpdateCommunity(community.community_id);

  const [name, setName] = useState(community.name);
  const [description, setDescription] = useState(community.description ?? '');
  const [rules, setRules] = useState(community.rules ?? '');
  const [visibility, setVisibility] = useState<CommunityVisibility>(community.visibility);
  const [bannerFile, setBannerFile] = useState<File | null>(null);
  const [bannerPreview, setBannerPreview] = useState<string | null>(community.banner_url || null);
  const [pendingCrop, setPendingCrop] = useState<PendingCrop | null>(null);
  const bannerRef = useRef<HTMLInputElement>(null);

  function handleBannerPick(file: File | null) {
    if (!file) return;
    setPendingCrop({ src: URL.createObjectURL(file), filename: file.name });
    if (bannerRef.current) bannerRef.current.value = '';
  }

  function onCropComplete(file: File, preview: string) {
    setBannerFile(file);
    setBannerPreview(preview);
    setPendingCrop(null);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    try {
      await mutateAsync({
        name: name.trim(),
        description: description.trim(),
        rules: rules.trim(),
        visibility,
        banner: bannerFile,
        categoryId,
      });
      onSuccess();
    } catch {
      // error shown via mutation error state
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
          <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)' }}>Edit Community</span>
          <button
            onClick={onClose}
            style={{ background: 'none', border: 'none', color: 'var(--text-tertiary)', fontSize: 20, cursor: 'pointer' }}
          >
            ✕
          </button>
        </div>

        {/* Banner */}
        <div
          onClick={() => bannerRef.current?.click()}
          style={{
            height: 100,
            background: bannerPreview ? `url(${bannerPreview}) center/cover no-repeat` : 'var(--bg-subtle)',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {!bannerPreview && <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Click to change banner</span>}
          <input ref={bannerRef} type="file" accept="image/*" style={{ display: 'none' }}
            onChange={(e) => handleBannerPick(e.target.files?.[0] ?? null)} />
        </div>

        <form onSubmit={handleSubmit} style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>Name *</label>
            <input style={inputStyle} value={name} onChange={(e) => setName(e.target.value)} placeholder="Community name" maxLength={100} required />
          </div>

          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>Category</label>
            <select style={{ ...inputStyle, appearance: 'none' }} disabled>
              {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              <option value="">{community.category_name || 'Current category'}</option>
            </select>
            <div style={{ fontSize: 11, color: 'var(--text-tertiary)', marginTop: 4 }}>Category cannot be changed after creation</div>
          </div>

          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>Visibility</label>
            <div style={{ display: 'flex', gap: 8 }}>
              {(['PUBLIC', 'PRIVATE'] as CommunityVisibility[]).map((v) => (
                <button key={v} type="button" onClick={() => setVisibility(v)}
                  style={{
                    flex: 1, padding: '8px', borderRadius: 8,
                    border: `1px solid ${visibility === v ? 'var(--text)' : 'var(--border)'}`,
                    background: visibility === v ? 'var(--text)' : 'transparent',
                    color: visibility === v ? 'var(--bg)' : 'var(--text)',
                    fontSize: 13, fontWeight: 600, cursor: 'pointer',
                  }}
                >
                  {v === 'PUBLIC' ? 'Public' : 'Private'}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>Description</label>
            <textarea style={{ ...inputStyle, minHeight: 72, resize: 'vertical' }} value={description} onChange={(e) => setDescription(e.target.value)} placeholder="What is this community about?" />
          </div>

          <div>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>Rules</label>
            <textarea style={{ ...inputStyle, minHeight: 72, resize: 'vertical' }} value={rules} onChange={(e) => setRules(e.target.value)} placeholder="Community rules and guidelines" />
          </div>

          {error && (
            <div style={{ fontSize: 13, color: 'var(--error, #e53e3e)' }}>
              {error instanceof Error ? error.message : 'Something went wrong'}
            </div>
          )}

          <button
            type="submit"
            disabled={isPending || !name.trim()}
            style={{
              width: '100%', padding: '12px', borderRadius: 999, border: 'none',
              background: 'var(--text)', color: 'var(--bg)', fontSize: 15, fontWeight: 700,
              cursor: isPending || !name.trim() ? 'default' : 'pointer',
              opacity: isPending || !name.trim() ? 0.5 : 1,
            }}
          >
            {isPending ? 'Saving…' : 'Save Changes'}
          </button>
        </form>
      </div>
    </div>

    {pendingCrop && (
      <ImageCropModal
        imageSrc={pendingCrop.src}
        filename={pendingCrop.filename}
        aspect={3}
        onComplete={onCropComplete}
        onClose={() => setPendingCrop(null)}
      />
    )}
    </>
  );
}
