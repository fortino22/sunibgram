interface ProfileBannerProps {
  bannerUrl?: string | null;
}

export function ProfileBanner({ bannerUrl }: ProfileBannerProps) {
  return (
    <div
      style={{
        height: 200,
        background: bannerUrl
          ? undefined
          : 'linear-gradient(135deg, oklch(28% 0.09 260), oklch(18% 0.05 290), oklch(22% 0.07 270))',
        flexShrink: 0,
        overflow: 'hidden',
      }}
    >
      {bannerUrl && (
        <img
          src={bannerUrl}
          alt=""
          style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
        />
      )}
    </div>
  );
}
