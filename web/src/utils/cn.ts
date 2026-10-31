import { type ClassValue, clsx } from 'clsx'
import { extendTailwindMerge } from 'tailwind-merge'

// Teach tailwind-merge the custom tokens from styles/index.css that it can't infer.
const twMerge = extendTailwindMerge({
  extend: {
    theme: {
      radius: ['sheet'],
    },
  },
})

// Merge class lists so later Tailwind utilities override earlier conflicting ones.
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
