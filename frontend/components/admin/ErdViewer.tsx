'use client';

import { useEffect, useRef, useState } from 'react';
import { Maximize2, Minimize2 } from 'lucide-react';
import {
  ERD_DOMAINS,
  ERD_ENTITIES,
  ERD_ORDER,
  ERD_RELS,
  type ErdRel,
} from '@/lib/admin/erdData';

const NS = 'http://www.w3.org/2000/svg';

interface Pt {
  x: number;
  y: number;
  dx: number;
  dy: number;
  side: string;
}
interface EndType {
  many: boolean;
  zero: boolean;
  one: boolean;
}

/**
 * Interactive ERD canvas (Visual-Paradigm style). The render engine is the
 * original vanilla DOM/SVG implementation ported verbatim into a single
 * per-mount useEffect closure so all mutable state (pos/scale/locked) is
 * scoped to the mount and torn down cleanly on unmount.
 */
export function ErdViewer() {
  const [expanded, setExpanded] = useState(false);
  const stageRef = useRef<HTMLDivElement | null>(null);
  const worldRef = useRef<HTMLDivElement | null>(null);
  const linksRef = useRef<SVGSVGElement | null>(null);
  const searchRef = useRef<HTMLInputElement | null>(null);
  const datalistRef = useRef<HTMLDataListElement | null>(null);

  // Imperative toolbar handlers are wired by the engine into this ref so the
  // JSX buttons can call them without re-rendering.
  const apiRef = useRef<{
    zin: () => void;
    zout: () => void;
    fit: () => void;
    clear: () => void;
    reset: () => void;
    png: () => void;
  } | null>(null);

  useEffect(() => {
    const stage = stageRef.current;
    const world = worldRef.current;
    const svg = linksRef.current;
    const search = searchRef.current;
    const datalist = datalistRef.current;
    if (!stage || !world || !svg || !search || !datalist) return;

    const pos: Record<string, { x: number; y: number }> = {};
    const elById: Record<string, HTMLDivElement> = {};
    const connByName: Record<string, SVGGElement[]> = {};
    const adj: Record<string, Set<string>> = {};
    ERD_ORDER.forEach((n) => {
      connByName[n] = [];
      adj[n] = new Set();
    });
    ERD_RELS.forEach((r) => {
      adj[r.from].add(r.to);
      adj[r.to].add(r.from);
    });

    // Listener bookkeeping for clean teardown.
    const cleanups: Array<() => void> = [];
    const onDoc = (
      type: 'mousemove' | 'mouseup',
      fn: (e: MouseEvent) => void,
    ) => {
      document.addEventListener(type, fn);
      return () => document.removeEventListener(type, fn);
    };

    function icon(kind: string): string {
      const m: Record<string, string> = { PK: 'i-pk', FK: 'i-fk', UK: 'i-uk', COL: 'i-col' };
      return '<svg class="ico" viewBox="0 0 14 14"><use href="#' + m[kind] + '"/></svg>';
    }
    function rowKind(k: string): string {
      return k.includes('PK') ? 'PK' : k.includes('FK') ? 'FK' : k.includes('UK') ? 'UK' : 'COL';
    }

    function buildEntity(name: string) {
      const cols = ERD_ENTITIES[name] || [];
      const e = document.createElement('div');
      e.className = 'entity';
      e.dataset.name = name;
      let rows = '';
      for (const c of cols) {
        const kind = rowKind(c.key);
        const iskey = kind !== 'COL';
        const nn = c.null
          ? '<svg class="nn" viewBox="0 0 15 14"><use href="#i-nn"/></svg>'
          : '<span style="width:15px"></span>';
        rows +=
          '<div class="row ' + (iskey ? 'key' : '') + '">' +
          icon(kind) +
          '<span class="nm">' + c.name + '</span><span class="ty">' + c.type + '</span>' + nn +
          '</div>';
      }
      e.innerHTML =
        '<div class="hd"><svg class="tico" viewBox="0 0 13 11"><use href="#i-tbl"/></svg>' +
        name + '</div><div class="rows">' + rows + '</div>';
      world!.appendChild(e);
      elById[name] = e;
      makeDraggable(e, name);
      e.addEventListener('mouseenter', () => { if (!locked) focus(name); });
      e.addEventListener('mouseleave', () => { if (!locked) clearFocus(); });
      const hd = e.querySelector('.hd') as HTMLElement;
      hd.addEventListener('click', (ev) => {
        ev.stopPropagation();
        if (locked === name) { locked = null; clearFocus(); }
        else { locked = name; focus(name, true); }
      });
    }

    function layout() {
      let x = 50;
      const colGap = 130, rowGap = 50, top = 50;
      const placed = new Set<string>();
      for (const dom in ERD_DOMAINS) {
        let y = top, maxW = 0;
        for (const t of ERD_DOMAINS[dom]) {
          if (!ERD_ENTITIES[t]) continue;
          pos[t] = { x, y };
          const el = elById[t];
          el.style.left = x + 'px';
          el.style.top = y + 'px';
          y += el.offsetHeight + rowGap;
          maxW = Math.max(maxW, el.offsetWidth);
          placed.add(t);
        }
        x += maxW + colGap;
      }
      let y = top;
      for (const t of ERD_ORDER) {
        if (placed.has(t)) continue;
        pos[t] = { x, y };
        const el = elById[t];
        el.style.left = x + 'px';
        el.style.top = y + 'px';
        y += el.offsetHeight + rowGap;
      }
    }

    function anchors(name: string) {
      const el = elById[name], p = pos[name];
      const w = el.offsetWidth, h = el.offsetHeight;
      return { x: p.x, y: p.y, w, h, cx: p.x + w / 2, cy: p.y + h / 2 };
    }

    function endType(sym: string): EndType {
      return { many: /[{}]/.test(sym), zero: /o/.test(sym), one: /\|/.test(sym) };
    }
    function marker(p: { x: number; y: number; dx: number; dy: number }, t: EndType): SVGGElement {
      const g = document.createElementNS(NS, 'g') as SVGGElement;
      const px = p.x, py = p.y, dx = p.dx, dy = p.dy, perpx = -dy, perpy = dx;
      const ln = (x1: number, y1: number, x2: number, y2: number) => {
        const l = document.createElementNS(NS, 'line');
        l.setAttribute('x1', String(x1)); l.setAttribute('y1', String(y1));
        l.setAttribute('x2', String(x2)); l.setAttribute('y2', String(y2));
        l.setAttribute('class', 'mk');
        g.appendChild(l);
      };
      const foot = 12, spread = 7, barOff = 11, hw = 6;
      if (t.many) {
        const ax = px + dx * foot, ay = py + dy * foot;
        ln(ax, ay, px + perpx * spread, py + perpy * spread);
        ln(ax, ay, px - perpx * spread, py - perpy * spread);
        ln(ax, ay, px, py);
      }
      if (t.one && !t.many) {
        const ox = px + dx * barOff, oy = py + dy * barOff;
        ln(ox + perpx * hw, oy + perpy * hw, ox - perpx * hw, oy - perpy * hw);
      }
      if (t.zero) {
        const off = t.many ? foot + 6 : barOff + 5;
        const c = document.createElementNS(NS, 'circle');
        c.setAttribute('cx', String(px + dx * off));
        c.setAttribute('cy', String(py + dy * off));
        c.setAttribute('r', '4');
        c.setAttribute('class', 'mk circ');
        g.appendChild(c);
      }
      return g;
    }
    function sideDir(side: string): [number, number] {
      return side === 'R' ? [1, 0] : side === 'L' ? [-1, 0] : side === 'T' ? [0, -1] : [0, 1];
    }

    interface Item {
      r: ErdRel;
      self: boolean;
      A?: ReturnType<typeof anchors>;
      B?: ReturnType<typeof anchors>;
      sa?: string;
      sb?: string;
    }

    function drawLinks() {
      while (svg!.firstChild) svg!.removeChild(svg!.firstChild);
      const items: Item[] = [];
      for (const r of ERD_RELS) {
        if (!elById[r.from] || !elById[r.to]) continue;
        if (r.from === r.to) { items.push({ r, self: true }); continue; }
        const A = anchors(r.from), B = anchors(r.to);
        const dx = B.cx - A.cx;
        // Anchor on left/right faces only — keeps every edge horizontally
        // tangent so the bezier curves read cleanly and fan out per face.
        const sa = dx >= 0 ? 'R' : 'L';
        const sb = dx >= 0 ? 'L' : 'R';
        items.push({ r, A, B, sa, sb, self: false });
      }
      const groups: Record<string, { i: number; opp: number; end: string }[]> = {};
      items.forEach((it, i) => {
        if (it.self) return;
        const ka = it.r.from + '|' + it.sa;
        const kb = it.r.to + '|' + it.sb;
        (groups[ka] = groups[ka] || []).push({ i, opp: (it.sa === 'L' || it.sa === 'R') ? it.B!.cy : it.B!.cx, end: 'a' });
        (groups[kb] = groups[kb] || []).push({ i, opp: (it.sb === 'L' || it.sb === 'R') ? it.A!.cy : it.A!.cx, end: 'b' });
      });
      const ptOf: Record<string, Pt> = {};
      for (const key in groups) {
        const [name, side] = key.split('|');
        const a = anchors(name);
        const arr = groups[key];
        arr.sort((u, v) => u.opp - v.opp);
        arr.forEach((o, idx) => {
          const t = (idx + 1) / (arr.length + 1);
          let x: number, y: number;
          if (side === 'R') { x = a.x + a.w; y = a.y + a.h * t; }
          else if (side === 'L') { x = a.x; y = a.y + a.h * t; }
          else if (side === 'T') { x = a.x + a.w * t; y = a.y; }
          else { x = a.x + a.w * t; y = a.y + a.h; }
          const [dx, dy] = sideDir(side);
          ptOf[o.i + o.end] = { x, y, dx, dy, side };
        });
      }
      items.forEach((it, i) => {
        const g = document.createElementNS(NS, 'g') as SVGGElement;
        g.setAttribute('class', 'conn');
        g.dataset.a = it.r.from;
        g.dataset.b = it.r.to;
        if (it.self) {
          const a = anchors(it.r.from);
          const x = a.x + a.w;
          const y1 = a.y + a.h * 0.32, y2 = a.y + a.h * 0.68;
          const ext = 34;
          const p = document.createElementNS(NS, 'path');
          p.setAttribute('d', `M${x} ${y1} L${x + ext} ${y1} L${x + ext} ${y2} L${x} ${y2}`);
          p.setAttribute('class', 'link');
          g.appendChild(p);
          g.appendChild(marker({ x, y: y1, dx: 1, dy: 0 }, endType(it.r.card.split('--')[0])));
          g.appendChild(marker({ x, y: y2, dx: 1, dy: 0 }, endType(it.r.card.split('--')[1])));
          svg!.appendChild(g);
          connByName[it.r.from].push(g);
          return;
        }
        const pa = ptOf[i + 'a'], pb = ptOf[i + 'b'];
        if (!pa || !pb) return;
        // Smooth cubic bezier with horizontal control handles pushed off each
        // L/R anchor — hub edges fan out instead of stacking into a channel.
        const c = Math.max(40, Math.min(180, Math.abs(pb.x - pa.x) * 0.4));
        const d =
          `M${pa.x} ${pa.y} C${pa.x + pa.dx * c} ${pa.y}, ` +
          `${pb.x + pb.dx * c} ${pb.y}, ${pb.x} ${pb.y}`;
        const p = document.createElementNS(NS, 'path');
        p.setAttribute('d', d);
        p.setAttribute('class', 'link');
        g.appendChild(p);
        const parts = it.r.card.split('--');
        g.appendChild(marker(pa, endType(parts[0])));
        g.appendChild(marker(pb, endType(parts[1])));
        svg!.appendChild(g);
        connByName[it.r.from].push(g);
        connByName[it.r.to].push(g);
      });
      sizeSvg();
    }
    function sizeSvg() {
      let mx = 0, my = 0;
      for (const n in pos) { const a = anchors(n); mx = Math.max(mx, a.x + a.w); my = Math.max(my, a.y + a.h); }
      svg!.setAttribute('width', String(mx + 120));
      svg!.setAttribute('height', String(my + 120));
      world!.style.width = (mx + 120) + 'px';
      world!.style.height = (my + 120) + 'px';
    }

    let locked: string | null = null;
    function focus(name: string, pick?: boolean) {
      const set = new Set([name]);
      adj[name].forEach((n) => set.add(n));
      for (const n in elById) {
        const el = elById[n];
        el.classList.toggle('dim', !set.has(n));
        el.classList.toggle('nbr', set.has(n) && n !== name);
        el.classList.toggle('pick', !!pick && n === name);
      }
      svg!.classList.add('focus');
      svg!.querySelectorAll('.conn').forEach((g) => {
        const el = g as SVGGElement;
        el.classList.toggle('hot', el.dataset.a === name || el.dataset.b === name);
      });
    }
    function clearFocus() {
      locked = null;
      for (const n in elById) elById[n].classList.remove('dim', 'nbr', 'pick');
      svg!.classList.remove('focus');
      svg!.querySelectorAll('.conn.hot').forEach((g) => g.classList.remove('hot'));
    }

    function makeDraggable(el: HTMLDivElement, name: string) {
      const hd = el.querySelector('.hd') as HTMLElement;
      hd.addEventListener('mousedown', (ev) => {
        ev.stopPropagation();
        ev.preventDefault();
        const sx = ev.clientX, sy = ev.clientY, ox = pos[name].x, oy = pos[name].y;
        function mv(e: MouseEvent) {
          pos[name].x = ox + (e.clientX - sx) / scale;
          pos[name].y = oy + (e.clientY - sy) / scale;
          el.style.left = pos[name].x + 'px';
          el.style.top = pos[name].y + 'px';
          drawLinks();
          if (locked) focus(locked, true);
        }
        const offMv = onDoc('mousemove', mv);
        const offUp = onDoc('mouseup', () => { offMv(); offUp(); });
      });
    }

    // ---------- pan & zoom ----------
    let scale = 1, tx = 0, ty = 0;
    function apply() {
      world!.style.transform = 'translate(' + tx + 'px,' + ty + 'px) scale(' + scale + ')';
    }
    const onStageDown = (ev: MouseEvent) => {
      if ((ev.target as HTMLElement).closest('.entity')) return;
      if (locked) clearFocus();
      stage!.classList.add('panning');
      const sx = ev.clientX, sy = ev.clientY, otx = tx, oty = ty;
      function mv(e: MouseEvent) { tx = otx + (e.clientX - sx); ty = oty + (e.clientY - sy); apply(); }
      const offMv = onDoc('mousemove', mv);
      const offUp = onDoc('mouseup', () => {
        stage!.classList.remove('panning');
        offMv(); offUp();
      });
    };
    stage.addEventListener('mousedown', onStageDown);
    cleanups.push(() => stage.removeEventListener('mousedown', onStageDown));

    const onWheel = (ev: WheelEvent) => {
      ev.preventDefault();
      const r = stage!.getBoundingClientRect(), mxp = ev.clientX - r.left, myp = ev.clientY - r.top;
      const f = ev.deltaY < 0 ? 1.1 : 1 / 1.1, ns = Math.min(2.5, Math.max(0.15, scale * f));
      tx = mxp - (mxp - tx) * (ns / scale);
      ty = myp - (myp - ty) * (ns / scale);
      scale = ns;
      apply();
    };
    stage.addEventListener('wheel', onWheel, { passive: false });
    cleanups.push(() => stage.removeEventListener('wheel', onWheel));

    function zoomBtn(f: number) {
      const r = stage!.getBoundingClientRect(), mxp = r.width / 2, myp = r.height / 2;
      const ns = Math.min(2.5, Math.max(0.15, scale * f));
      tx = mxp - (mxp - tx) * (ns / scale);
      ty = myp - (myp - ty) * (ns / scale);
      scale = ns;
      apply();
    }
    function fit() {
      let mx = 0, my = 0;
      for (const n in pos) { const a = anchors(n); mx = Math.max(mx, a.x + a.w); my = Math.max(my, a.y + a.h); }
      const r = stage!.getBoundingClientRect();
      scale = Math.min(r.width / (mx + 100), r.height / (my + 100), 1);
      tx = 12; ty = 12;
      apply();
    }

    // datalist options
    ERD_ORDER.forEach((n) => {
      const o = document.createElement('option');
      o.value = n;
      datalist!.appendChild(o);
    });
    const onSearch = () => {
      const n = search!.value.trim();
      if (!elById[n]) return;
      locked = n;
      focus(n, true);
      const a = anchors(n);
      const r = stage!.getBoundingClientRect();
      scale = 1;
      tx = r.width / 2 - a.cx * scale;
      ty = r.height / 2 - a.cy * scale;
      apply();
    };
    search.addEventListener('change', onSearch);
    cleanups.push(() => search.removeEventListener('change', onSearch));

    async function png() {
      try {
        // Remote ESM via CDN — no bundling, no extra dependency.
        const mod = (await import(
          /* webpackIgnore: true */
          // @ts-expect-error remote module has no local types
          'https://cdn.jsdelivr.net/npm/html-to-image@1.11.11/+esm'
        )) as { toPng: (node: HTMLElement, opts: Record<string, unknown>) => Promise<string> };
        const prev = world!.style.transform;
        world!.style.transform = 'none';
        const themeBg =
          getComputedStyle(document.documentElement).getPropertyValue('--bg').trim() ||
          '#ffffff';
        const dataUrl = await mod.toPng(world!, {
          backgroundColor: themeBg,
          width: parseInt(world!.style.width),
          height: parseInt(world!.style.height),
          pixelRatio: 1.5,
        });
        world!.style.transform = prev;
        const a = document.createElement('a');
        a.href = dataUrl;
        a.download = 'sunibgram_erd.png';
        a.click();
      } catch (e) {
        alert('PNG export needs internet (CDN). ' + (e as Error).message);
      }
    }

    // resize / overflow observers
    const onResize = () => { /* layout is static; only fit-on-resize would change view */ };
    window.addEventListener('resize', onResize);
    cleanups.push(() => window.removeEventListener('resize', onResize));

    apiRef.current = {
      zin: () => zoomBtn(1.15),
      zout: () => zoomBtn(1 / 1.15),
      fit,
      clear: clearFocus,
      reset: () => { layout(); drawLinks(); fit(); },
      png,
    };

    // ---------- init ----------
    ERD_ORDER.forEach(buildEntity);
    layout();
    drawLinks();
    fit();

    return () => {
      cleanups.forEach((fn) => fn());
      apiRef.current = null;
      // wipe DOM the engine built so a remount starts clean (keep the svg node)
      world.querySelectorAll('.entity').forEach((el) => el.remove());
      while (svg.firstChild) svg.removeChild(svg.firstChild);
      datalist.innerHTML = '';
    };
  }, []);

  // Re-fit the diagram whenever the viewport size changes (expand/collapse).
  useEffect(() => {
    apiRef.current?.fit();
  }, [expanded]);

  // Esc collapses the fullscreen overlay.
  useEffect(() => {
    if (!expanded) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setExpanded(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [expanded]);

  return (
    <div className={'erd-root' + (expanded ? ' erd-expanded' : '')}>
      <style>{ERD_CSS}</style>

      {/* shared SVG defs (icons) — monochrome, inherit currentColor so they
          follow the app theme. PK/FK/UK distinguished by shape, not hue. */}
      <svg width="0" height="0" style={{ position: 'absolute' }} aria-hidden>
        <defs>
          <g id="i-pk">
            <circle cx="5" cy="6.5" r="2.4" fill="none" stroke="currentColor" strokeWidth="1.4" />
            <path d="M7 7 L12 12 M10.5 10.5 L12 9.5" stroke="currentColor" strokeWidth="1.6" fill="none" />
          </g>
          <g id="i-fk">
            <path d="M7 1.5 L12.5 7 L7 12.5 L1.5 7 Z" fill="none" stroke="currentColor" strokeWidth="1.3" />
          </g>
          <g id="i-uk">
            <rect x="2" y="2" width="10" height="10" rx="2" fill="none" stroke="currentColor" strokeWidth="1.2" />
            <text x="7" y="10.5" fontSize="8" textAnchor="middle" fill="currentColor" fontFamily="Segoe UI">U</text>
          </g>
          <g id="i-col">
            <rect x="2" y="2" width="10" height="10" rx="2" fill="none" stroke="currentColor" strokeWidth="1.1" />
            <line x1="4" y1="5.5" x2="10" y2="5.5" stroke="currentColor" strokeWidth="1" />
            <line x1="4" y1="8.5" x2="10" y2="8.5" stroke="currentColor" strokeWidth="1" />
          </g>
          <g id="i-tbl">
            <rect x="0.5" y="1" width="12" height="9" rx="1.5" fill="none" stroke="currentColor" strokeWidth="1" />
            <line x1="0.5" y1="4" x2="12.5" y2="4" stroke="currentColor" strokeWidth="1" />
            <line x1="4.5" y1="1" x2="4.5" y2="10" stroke="currentColor" strokeWidth="1" />
          </g>
          <g id="i-nn">
            <rect x="0.5" y="0.5" width="14" height="13" rx="2" fill="none" stroke="currentColor" strokeWidth="1" />
            <text x="7.5" y="10.5" fontSize="9" fontWeight="700" textAnchor="middle" fill="currentColor" fontFamily="Segoe UI">N</text>
          </g>
        </defs>
      </svg>

      <div className="erd-toolbar">
        <input
          ref={searchRef}
          className="erd-search"
          placeholder="find table…"
          list="erd-tnames"
        />
        <datalist id="erd-tnames" ref={datalistRef} />
        <button type="button" onClick={() => apiRef.current?.fit()}>Fit</button>
        <button type="button" onClick={() => apiRef.current?.reset()}>Reset</button>
        <button type="button" onClick={() => apiRef.current?.clear()}>Clear focus</button>
        <button type="button" onClick={() => apiRef.current?.zin()}>＋</button>
        <button type="button" onClick={() => apiRef.current?.zout()}>－</button>
        <button type="button" onClick={() => apiRef.current?.png()}>PNG</button>
        <button
          type="button"
          className="erd-expand-btn"
          aria-label={expanded ? 'Collapse' : 'Expand'}
          onClick={() => setExpanded((v) => !v)}
        >
          {expanded ? <Minimize2 size={14} strokeWidth={2.2} /> : <Maximize2 size={14} strokeWidth={2.2} />}
          {expanded ? 'Collapse' : 'Expand'}
        </button>
        <span className="erd-sp" />
        <div className="erd-legend">
          <span className="k key"><svg width="14" height="14"><use href="#i-pk" /></svg>PK</span>
          <span className="k key"><svg width="14" height="14"><use href="#i-fk" /></svg>FK</span>
          <span className="k"><svg width="14" height="14"><use href="#i-col" /></svg>Column</span>
          <span className="k"><svg width="15" height="14"><use href="#i-nn" /></svg>Nullable</span>
        </div>
      </div>

      <div className="erd-stage" ref={stageRef}>
        <div className="erd-world" ref={worldRef}>
          <svg className="erd-links" ref={linksRef} />
        </div>
      </div>
      <div className="erd-hint">
        Hover or click a table to spotlight its relationships • drag headers to move • drag canvas to pan • scroll to zoom
      </div>
    </div>
  );
}

const ERD_CSS = `
.erd-root{--erd-blue:#1d9bf0;--erd-blue-soft:color-mix(in oklab,#1d9bf0 14%,transparent);position:relative;height:calc(100vh - 230px);min-height:420px;border:1px solid var(--border);border-radius:10px;overflow:hidden;background:var(--bg);color:var(--text);font-family:inherit}
.erd-root.erd-expanded{position:fixed;inset:0;height:100vh;min-height:0;z-index:60;border:none;border-radius:0}
.erd-expand-btn{display:inline-flex;align-items:center;gap:5px}
.erd-toolbar{position:absolute;top:0;left:0;right:0;height:46px;background:var(--bg-card);border-bottom:1px solid var(--border);display:flex;align-items:center;gap:8px;padding:0 14px;z-index:5}
.erd-toolbar button{font-size:12.5px;border:1px solid var(--border);background:var(--bg-subtle);color:var(--text-secondary);border-radius:6px;padding:6px 10px;cursor:pointer;transition:background .12s,color .12s}
.erd-toolbar button:hover{background:var(--bg-hover);color:var(--text)}
.erd-search{font-size:12.5px;border:1px solid var(--border);background:var(--bg-subtle);color:var(--text);border-radius:6px;padding:6px 10px;width:150px}
.erd-search::placeholder{color:var(--text-tertiary)}
.erd-sp{flex:1}
.erd-legend{font-size:11.5px;color:var(--text-tertiary);display:flex;gap:12px;align-items:center}
.erd-legend .k{display:inline-flex;gap:4px;align-items:center;color:var(--text-secondary)}
.erd-legend .k.key{color:var(--erd-blue)}
.erd-stage{position:absolute;inset:46px 0 0 0;overflow:hidden;cursor:grab;background-image:radial-gradient(var(--border) 1px,transparent 1px);background-size:24px 24px}
.erd-stage.panning{cursor:grabbing}
.erd-world{position:absolute;top:0;left:0;transform-origin:0 0}
.erd-links{position:absolute;top:0;left:0;overflow:visible;pointer-events:none}
.erd-root .entity{position:absolute;width:230px;background:var(--bg-card);border:1px solid var(--border);border-radius:8px;box-shadow:0 1px 3px rgba(0,0,0,.18);font-size:12px;user-select:none;transition:opacity .15s,box-shadow .15s}
.erd-root .entity .hd{display:flex;align-items:center;gap:6px;padding:6px 9px;cursor:move;background:var(--erd-blue-soft);border-bottom:1px solid color-mix(in oklab,var(--erd-blue) 45%,var(--border));border-radius:7px 7px 0 0;font-weight:700;color:var(--text)}
.erd-root .entity .hd .tico{width:13px;height:11px;flex:0 0 auto;color:var(--erd-blue)}
.erd-root .entity .rows{padding:2px 0}
.erd-root .row{display:flex;align-items:center;gap:6px;padding:2.5px 9px;line-height:1.35}
.erd-root .row:nth-child(even){background:color-mix(in oklab,var(--text) 4%,transparent)}
.erd-root .row .ico{width:14px;height:14px;flex:0 0 auto;color:var(--text-tertiary)}
.erd-root .row.key .ico{color:var(--erd-blue)}
.erd-root .row .nm{flex:1;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;color:var(--text)}
.erd-root .row.key .nm{font-weight:700;font-style:italic}
.erd-root .row .ty{color:var(--text-tertiary);font-style:italic;margin-left:8px;white-space:nowrap}
.erd-root .row .nn{width:15px;height:14px;flex:0 0 auto;opacity:.85;color:var(--text-tertiary)}
.erd-root .entity.dim{opacity:.22}
.erd-root .entity.nbr{box-shadow:0 0 0 1.5px var(--erd-blue),0 2px 8px rgba(0,0,0,.22)}
.erd-root .entity.pick{box-shadow:0 0 0 2px var(--erd-blue),0 0 12px color-mix(in oklab,var(--erd-blue) 55%,transparent)}
.erd-root .link{stroke:var(--erd-blue);stroke-width:1.3;fill:none}
.erd-root .mk{stroke:var(--erd-blue);stroke-width:1.3;fill:none}
.erd-root .mk.circ{fill:var(--bg)}
.erd-links .conn{opacity:.16;transition:opacity .15s}
.erd-links.focus .conn{opacity:.07}
.erd-links.focus .conn.hot{opacity:1}
.erd-root .conn.hot{filter:drop-shadow(0 0 3px var(--erd-blue))}
.erd-root .conn.hot .link{stroke:var(--erd-blue);stroke-width:2.2}
.erd-root .conn.hot .mk{stroke:var(--erd-blue);stroke-width:2.2}
.erd-hint{position:absolute;bottom:10px;left:14px;font-size:11px;color:var(--text-tertiary);z-index:4;background:color-mix(in oklab,var(--bg) 80%,transparent);padding:3px 8px;border-radius:5px;pointer-events:none}
`;
