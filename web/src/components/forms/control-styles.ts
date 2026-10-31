// Shared look for text inputs, selects and textareas (mockup: 44px, radius 10, line border).
export const controlBase =
  'w-full min-w-0 rounded-md border border-line bg-card text-base text-ink placeholder:text-ink-muted ' +
  'disabled:opacity-60 aria-invalid:border-expense focus-visible:border-brand'

export const controlSizes = {
  md: 'h-11 px-3',
  lg: 'h-12 rounded-lg px-3.5 text-md', // login form
} as const

export type ControlSize = keyof typeof controlSizes
