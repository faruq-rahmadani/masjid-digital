// Single place that reads import.meta.env, so the rest of the app gets typed values.
export const env = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? '/api',
} as const
