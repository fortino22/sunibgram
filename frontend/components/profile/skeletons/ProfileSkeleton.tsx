function Bone({ w, h, radius = 8 }: { w: number | string; h: number; radius?: number }) {
  return (
    <div
      style={{
        width: w,
        height: h,
        borderRadius: radius,
        background: 'var(--bg-subtle)',
        animation: 'pulse 1.5s ease-in-out infinite',
      }}
    />
  );
}

export function ProfileSkeleton() {
  return (
    <div>
      <style>{`@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.4} }`}</style>

      {/* Header */}
      <div style={{ height: 57, borderBottom: '1px solid var(--border)', background: 'var(--bg-card)' }} />

      {/* Banner */}
      <div style={{ height: 200, background: 'var(--bg-subtle)' }} />

      {/* Avatar + button row */}
      <div style={{ position: 'relative', padding: '0 20px' }}>
        <div
          style={{
            position: 'absolute',
            top: -44,
            left: 20,
            width: 84,
            height: 84,
            borderRadius: '50%',
            background: 'var(--bg-subtle)',
            border: '4px solid var(--bg-card)',
            animation: 'pulse 1.5s ease-in-out infinite',
          }}
        />
        <div style={{ display: 'flex', justifyContent: 'flex-end', paddingTop: 14 }}>
          <Bone w={110} h={36} radius={20} />
        </div>

        <div style={{ marginTop: 46, paddingBottom: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
          <Bone w={160} h={22} />
          <Bone w={100} h={16} />
          <div style={{ marginTop: 4, display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Bone w="90%" h={15} />
            <Bone w="70%" h={15} />
          </div>
          <div style={{ display: 'flex', gap: 20, marginTop: 4 }}>
            <Bone w={80} h={16} />
            <Bone w={80} h={16} />
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div style={{ height: 49, borderBottom: '1px solid var(--border)', borderTop: '1px solid var(--border)' }} />

      {/* Thread skeletons */}
      {[1, 2, 3].map((i) => (
        <div key={i} style={{ padding: '16px 20px', display: 'flex', gap: 12, borderBottom: '1px solid var(--border)' }}>
          <div style={{ width: 40, height: 40, borderRadius: '50%', background: 'var(--bg-subtle)', flexShrink: 0, animation: 'pulse 1.5s ease-in-out infinite' }} />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Bone w={120} h={14} />
            <Bone w="85%" h={14} />
            <Bone w="60%" h={14} />
          </div>
        </div>
      ))}
    </div>
  );
}
