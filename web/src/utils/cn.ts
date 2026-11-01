import { type ClassValue, clsx } from 'clsx'
import { extendTailwindMerge } from 'tailwind-merge'

// Teach tailwind-merge the custom tokens from styles/index.css that it can't infer.
const twMerge = extendTailwindMerge({
  extend: {
    theme: {
      radius: ['sheet', 'tv', 'tv-lg'],
      // TV sizes are named by px (text-tv-26); without this they read as colors and clash.
      text: [(value: string) => /^tv-\d+$/.test(value)],
    },
  },
})

// Merge class lists so later Tailwind utilities override earlier conflicting ones.
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
