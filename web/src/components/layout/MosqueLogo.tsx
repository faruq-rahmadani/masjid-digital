// Placeholder mark from the mockup until each mosque uploads its own logo.
export function MosqueLogo({ size = 72, className }: { size?: number; className?: string }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 72 72"
      fill="none"
      role="img"
      aria-label="Logo masjid"
      className={className}
    >
      <circle cx="36" cy="36" r="34" className="fill-brand-2" />
      <g className="stroke-gold" strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 50h32M24 50V38c0-8 5-14 12-17 7 3 12 9 12 17v12" strokeWidth="3" />
        <path d="M36 21v-7" strokeWidth="2.5" />
        <path d="M32 50v-7a4 4 0 0 1 8 0v7" strokeWidth="3" />
      </g>
    </svg>
  )
}
