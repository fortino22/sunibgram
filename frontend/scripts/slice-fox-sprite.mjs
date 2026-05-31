// Slices the fox sprite sheets into individual frame PNGs by detecting
// transparent-column gutters in the actual pixel data, so the popup animation
// can cycle clean frames instead of relying on a (mis-)assumed frame width.
//
// Run:  node scripts/slice-fox-sprite.mjs
// Writes: public/assets/fox-sleep/<white|black>/NN.png

import sharp from 'sharp';
import { mkdir, rm, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const ASSETS = join(ROOT, 'public', 'assets');
const OUT_BASE = join(ASSETS, 'fox-sleep');

const ALPHA_GUTTER = 8; // a column is a gutter if every pixel alpha <= this

const SHEETS = [
  { name: 'white', file: join(ASSETS, 'white_fox.png') },
  { name: 'black', file: join(ASSETS, 'black_fox.png') },
];

/** Find frame [start,end) column ranges from transparent-gutter detection. */
function detectFrames(data, width, height, channels) {
  const colHasInk = new Array(width).fill(false);
  for (let x = 0; x < width; x++) {
    for (let y = 0; y < height; y++) {
      const alpha = channels === 4 ? data[(y * width + x) * channels + 3] : 255;
      if (alpha > ALPHA_GUTTER) { colHasInk[x] = true; break; }
    }
  }
  // Contiguous runs of ink columns = frame content spans.
  const spans = [];
  let runStart = -1;
  for (let x = 0; x <= width; x++) {
    const ink = x < width && colHasInk[x];
    if (ink && runStart === -1) runStart = x;
    else if (!ink && runStart !== -1) { spans.push([runStart, x]); runStart = -1; }
  }
  if (spans.length === 0) return [];
  // Cut at gutter midpoints so each cell keeps its surrounding padding.
  const cuts = [0];
  for (let i = 0; i < spans.length - 1; i++) {
    cuts.push(Math.round((spans[i][1] + spans[i + 1][0]) / 2));
  }
  cuts.push(width);
  const frames = [];
  for (let i = 0; i < cuts.length - 1; i++) frames.push([cuts[i], cuts[i + 1]]);
  return frames;
}

const counts = {};

for (const sheet of SHEETS) {
  const img = sharp(sheet.file);
  const meta = await img.metadata();
  const { width, height, channels } = meta;
  const raw = await img.ensureAlpha().raw().toBuffer();
  const detected = detectFrames(raw, width, height, channels === 4 ? channels : 4);

  // The leading walk poses touch in the source (no transparent gutter) so they
  // fuse into one over-wide cell. Drop any frame wider than 1.4x the median —
  // that's the un-sliceable fused walk frame — keeping the clean settle→sleep
  // single poses (sit, curl, sleeping balls).
  const widths = detected.map((f) => f[1] - f[0]).sort((a, b) => a - b);
  const median = widths[Math.floor(widths.length / 2)] || 1;
  const frames = detected.filter((f) => f[1] - f[0] <= median * 1.4);

  console.log(`\n[${sheet.name}] ${width}x${height} ch=${channels} -> ${detected.length} detected, ${frames.length} kept (median=${median})`);
  frames.forEach((f, i) => console.log(`  frame ${String(i).padStart(2, '0')}: x ${f[0]}..${f[1]} (w=${f[1] - f[0]})`));

  const outDir = join(OUT_BASE, sheet.name);
  await rm(outDir, { recursive: true, force: true });
  await mkdir(outDir, { recursive: true });

  for (let i = 0; i < frames.length; i++) {
    const [x0, x1] = frames[i];
    const name = `${String(i).padStart(2, '0')}.png`;
    await sharp(sheet.file)
      .ensureAlpha()
      .extract({ left: x0, top: 0, width: x1 - x0, height })
      .toFile(join(outDir, name));
  }
  console.log(`  wrote ${frames.length} files -> ${outDir}`);
  counts[sheet.name] = frames.length;
}

// Emit per-theme frame counts for the component.
await writeFile(
  join(OUT_BASE, 'frames.json'),
  JSON.stringify(counts, null, 2) + '\n',
);
console.log(`\ncounts = ${JSON.stringify(counts)}`);
