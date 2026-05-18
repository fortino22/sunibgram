export interface CompressOptions {
  maxDimension?: number;
  maxSizeBytes?: number;
  quality?: number;
}

const DEFAULTS = {
  maxDimension: 1920,
  maxSizeBytes: 800 * 1024, // 800 KB
  quality: 0.85,
};

/**
 * Compress an image file using canvas, reducing dimensions and/or quality
 * until the output is under maxSizeBytes. Non-image files are returned as-is.
 * PNG/JPEG/WebP inputs → JPEG output (WebP stays WebP).
 */
export async function compressImage(
  file: File,
  opts: CompressOptions = {},
): Promise<File> {
  if (!file.type.startsWith('image/') || file.type === 'image/gif') return file;

  const maxDim = opts.maxDimension ?? DEFAULTS.maxDimension;
  const maxBytes = opts.maxSizeBytes ?? DEFAULTS.maxSizeBytes;
  const startQ = opts.quality ?? DEFAULTS.quality;

  return new Promise((resolve, reject) => {
    const img = new Image();
    const src = URL.createObjectURL(file);

    img.onload = () => {
      URL.revokeObjectURL(src);

      let w = img.naturalWidth;
      let h = img.naturalHeight;

      if (w > maxDim || h > maxDim) {
        const scale = Math.min(maxDim / w, maxDim / h);
        w = Math.round(w * scale);
        h = Math.round(h * scale);
      }

      // Already fits — no work needed
      if (file.size <= maxBytes && img.naturalWidth === w && img.naturalHeight === h) {
        resolve(file);
        return;
      }

      const canvas = document.createElement('canvas');
      canvas.width = w;
      canvas.height = h;
      const ctx = canvas.getContext('2d');
      if (!ctx) { reject(new Error('Canvas unavailable')); return; }
      ctx.drawImage(img, 0, 0, w, h);

      const outMime = file.type === 'image/webp' ? 'image/webp' : 'image/jpeg';
      const outExt = outMime === 'image/jpeg' ? '.jpg' : '.webp';
      const outName = file.name.replace(/\.[^.]+$/, outExt);

      let q = startQ;

      const attempt = () => {
        canvas.toBlob(
          (blob) => {
            if (!blob) { reject(new Error('Compression failed')); return; }
            if (blob.size <= maxBytes || q <= 0.3) {
              resolve(new File([blob], outName, { type: outMime }));
            } else {
              q = Math.max(0.3, q - 0.1);
              attempt();
            }
          },
          outMime,
          q,
        );
      };

      attempt();
    };

    img.onerror = () => {
      URL.revokeObjectURL(src);
      resolve(file); // fall back to original on load failure
    };

    img.src = src;
  });
}

export async function compressImages(
  files: File[],
  opts?: CompressOptions,
): Promise<File[]> {
  return Promise.all(files.map((f) => compressImage(f, opts)));
}
