// Sections of the /kit page, in reading order.
export const GROUPS = [
  { id: 'foundations', label: 'Foundations' },
  { id: 'actions', label: 'Actions' },
  { id: 'forms', label: 'Forms' },
  { id: 'data-display', label: 'Data display' },
  { id: 'feedback', label: 'Feedback' },
  { id: 'overlays', label: 'Overlays' },
  { id: 'navigation', label: 'Navigation' },
  { id: 'layout', label: 'Layout' },
] as const

export type GroupId = (typeof GROUPS)[number]['id']
