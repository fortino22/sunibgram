'use client';

import { useState, useCallback } from 'react';
import Cropper from 'react-easy-crop';
import type { Area, Point } from 'react-easy-crop';
import { X, Check } from 'lucide-react';
import { getCroppedFile } from '@/lib/cropImage';
import { compressImage, IMAGE_MAX_BYTES } from '@/lib/compressImage';

interface ImageCropModalProps {
  imageSrc: string;
  filename: string;
  aspect: number;
  maxSizeBytes?: number;
  onComplete: (file: File, preview: string) => void;
  onClose: () => void;
}

export function ImageCropModal({ imageSrc, filename, aspect, maxSizeBytes = IMAGE_MAX_BYTES.avatar, onComplete, onClose }: ImageCropModalProps) {
  const [crop, setCrop] = useState<Point>({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [croppedAreaPixels, setCroppedAreaPixels] = useState<Area | null>(null);
  const [applying, setApplying] = useState(false);
  const [sizeError, setSizeError] = useState<string | null>(null);

  const onCropComplete = useCallback((_: Area, croppedPixels: Area) => {
    setCroppedAreaPixels(croppedPixels);
  }, []);

  async function handleConfirm() {
    if (!croppedAreaPixels) return;
    setApplying(true);
    setSizeError(null);
    try {
      const cropped = await getCroppedFile(imageSrc, croppedAreaPixels, filename);
      const compressed = await compressImage(cropped, {
        maxDimension: 1200,
        maxSizeBytes,
      });
      if (compressed.size > maxSizeBytes) {
        setSizeError('Image is still too large after compression. Try a smaller or less detailed image.');
        return;
      }
      const preview = URL.createObjectURL(compressed);
      onComplete(compressed, preview);
    } finally {
      setApplying(false);
    }
  }

  return (
    <div
      style={{ position: 'fixed', inset: 0, zIndex: 200, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.75)', backdropFilter: 'blur(4px)' }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div style={{ background: 'var(--bg-card)', borderRadius: 16, width: '100%', maxWidth: 480, overflow: 'hidden', display: 'flex', flexDirection: 'column', boxShadow: 'var(--shadow-dropdown)' }}>
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 20px', borderBottom: '1px solid var(--border)', flexShrink: 0 }}>
          <button
            type="button"
            onClick={onClose}
            style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: 6, fontSize: 14, fontFamily: 'inherit', padding: '4px 0' }}
          >
            <X size={16} strokeWidth={2} />
            Cancel
          </button>
          <span style={{ color: 'var(--text)', fontWeight: 700, fontSize: 15, letterSpacing: '-0.02em' }}>Crop photo</span>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={applying}
            style={{ background: 'none', border: 'none', cursor: applying ? 'not-allowed' : 'pointer', color: 'var(--accent)', display: 'flex', alignItems: 'center', gap: 6, fontSize: 14, fontWeight: 700, fontFamily: 'inherit', padding: '4px 0', opacity: applying ? 0.6 : 1 }}
          >
            <Check size={16} strokeWidth={2.5} />
            Apply
          </button>
        </div>

        {/* Crop area — fixed height so the modal stays compact */}
        <div style={{ position: 'relative', height: 300, background: '#111' }}>
          <Cropper
            image={imageSrc}
            crop={crop}
            zoom={zoom}
            aspect={aspect}
            onCropChange={setCrop}
            onZoomChange={setZoom}
            onCropComplete={onCropComplete}
            style={{
              containerStyle: { background: '#111' },
              cropAreaStyle: { border: '2px solid rgba(255,255,255,0.8)' },
            }}
          />
        </div>

        {/* Zoom slider */}
        <div style={{ padding: '14px 24px 16px', borderTop: '1px solid var(--border)', display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span style={{ fontSize: 16, color: 'var(--text-tertiary)', lineHeight: 1, userSelect: 'none' }}>−</span>
            <input
              type="range"
              min={1}
              max={3}
              step={0.01}
              value={zoom}
              onChange={(e) => { setZoom(Number(e.target.value)); setSizeError(null); }}
              style={{ flex: 1, accentColor: 'var(--accent)', cursor: 'pointer' }}
            />
            <span style={{ fontSize: 16, color: 'var(--text-tertiary)', lineHeight: 1, userSelect: 'none' }}>+</span>
          </div>
          {sizeError && (
            <p style={{ margin: 0, fontSize: 12, color: 'var(--coral)', textAlign: 'center' }}>{sizeError}</p>
          )}
        </div>
      </div>
    </div>
  );
}
