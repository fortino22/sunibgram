'use client';

import { useState, useRef, useEffect } from 'react';
import { toast } from 'sonner';
import { useCategories } from '@/hooks/useCommunities';
import { useUpdateCommunity } from '@/hooks/useCommunityMutations';
import { INACTIVE_COMMUNITY_TOAST, isInactiveCommunityError } from '@/lib/api';
import { ImageCropModal } from '@/components/profile/ImageCropModal';
import type { CommunityData, CommunityVisibility } from '@/lib/services/communities';

type PendingCrop = { src: string; filename: string };

interface EditCommunityModalProps {
  community: CommunityData;
  onClose: () => void;
  onSuccess: () => void;
  onDelete: () => void;
  deletePending: boolean;
}

export function EditCommunityModal({ community, onClose, onSuccess, onDelete, deletePending }: EditCommunityModalProps) {
  const { data: categoriesData } = useCategories();
  const categories = categoriesData?.data ?? [];
  const { mutateAsync, isPending, error } = useUpdateCommunity(community.community_id);

  const [name, setName] = useState(community.name);
  const [description, setDescription] = useState(community.description ?? '');
  const [selectedCategoryId, setSelectedCategoryId] = useState('');
  const [keepRuleIds, setKeepRuleIds] = useState<number[]>((community.rules ?? []).map((r) => r.id));
  const [newRules, setNewRules] = useState<string[]>([]);
  const [visibility, setVisibility] = useState<CommunityVisibility>(community.visibility);

  useEffect(() => {
    if (!selectedCategoryId && categories.length > 0) {
      const found = categories.find((c) => c.name === community.category_name)?.id ?? '';
      setSelectedCategoryId(found);
    }
  }, [categories, community.category_name, selectedCategoryId]);
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
    if (description.length > 2000) {
      toast.error('Description cannot exceed 2000 characters.');
      return;
    }
    const badRuleIdx = newRules.findIndex((r) => r.trim().length > 500);
    if (badRuleIdx !== -1) {
      toast.error(`New rule ${badRuleIdx + 1} exceeds 500 characters.`);
      return;
    }
    if (keepRuleIds.length + newRules.length > 10) {
      toast.error('A community can have at most 10 rules.');
      return;
    }
    try {
      await mutateAsync({
        name: name.trim(),
        description: description.trim(),
        keep_rule_ids: keepRuleIds,
        new_rules: newRules.map((r) => r.trim()).filter(Boolean),
        visibility,
        banner: bannerFile,
        categoryId: selectedCategoryId,
      });
      onSuccess();
    } catch (err) {
      if (isInactiveCommunityError(err)) {
        toast.error(INACTIVE_COMMUNITY_TOAST);
      }
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
            aspectRatio: '2.31',
            width: '100%',
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
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: 6 }}>Category *</label>
            <select
              style={{ ...inputStyle, appearance: 'none' }}
              value={selectedCategoryId}
              onChange={(e) => setSelectedCategoryId(e.target.value)}
              required
            >
              <option value="">Select category</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
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
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
              <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}>Description</label>
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
              <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}>Rules</label>
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>
                {keepRuleIds.length + newRules.length}/10
              </span>
            </div>
            {(community.rules ?? []).filter((rule) => keepRuleIds.includes(rule.id)).map((rule) => (
              <div key={rule.id} style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6, padding: '8px 10px', borderRadius: 8, background: 'var(--bg-subtle)', border: '1px solid var(--border)' }}>
                <span style={{ fontSize: 13, color: 'var(--text)', lineHeight: 1.5, flex: 1 }}>
                  {rule.text}
                </span>
                <button
                  type="button"
                  onClick={() => setKeepRuleIds(keepRuleIds.filter((id) => id !== rule.id))}
                  style={{
                    background: 'none', border: 'none', color: 'var(--text-tertiary)',
                    cursor: 'pointer', padding: '2px 4px', fontSize: 16, lineHeight: 1,
                    flexShrink: 0, display: 'flex', alignItems: 'center',
                  }}
                  title="Remove rule"
                >
                  ✕
                </button>
              </div>
            ))}
            {newRules.map((rule, i) => (
              <div key={i} style={{ marginBottom: 6 }}>
                <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                  <input
                    style={{ ...inputStyle, flex: 1 }}
                    value={rule}
                    onChange={(e) => {
                      const next = [...newRules];
                      next[i] = e.target.value;
                      setNewRules(next);
                    }}
                    placeholder={`New rule ${i + 1}`}
                    maxLength={500}
                  />
                  <button
                    type="button"
                    onClick={() => setNewRules(newRules.filter((_, j) => j !== i))}
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
            {keepRuleIds.length + newRules.length < 10 && (
              <button
                type="button"
                onClick={() => setNewRules([...newRules, ''])}
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
              {isInactiveCommunityError(error)
                ? INACTIVE_COMMUNITY_TOAST
                : error instanceof Error
                  ? error.message
                  : 'Something went wrong'}
            </div>
          )}

          <button
            type="button"
            onClick={onDelete}
            disabled={deletePending}
            style={{
              width: '100%', padding: '12px', borderRadius: 999,
              border: '1px solid var(--error, #e53e3e)', background: 'transparent',
              color: 'var(--error, #e53e3e)', fontSize: 15, fontWeight: 700,
              cursor: deletePending ? 'default' : 'pointer',
              opacity: deletePending ? 0.5 : 1,
              transition: 'background 0.12s ease',
            }}
            onMouseEnter={(e) => { if (!deletePending) e.currentTarget.style.background = 'rgba(239,68,68,0.08)'; }}
            onMouseLeave={(e) => { if (!deletePending) e.currentTarget.style.background = 'transparent'; }}
          >
            {deletePending ? 'Deleting…' : 'Delete Community'}
          </button>

          <button
            type="submit"
            className="btn-solid"
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
        aspect={2.31}
        onComplete={onCropComplete}
        onClose={() => setPendingCrop(null)}
      />
    )}
    </>
  );
}
