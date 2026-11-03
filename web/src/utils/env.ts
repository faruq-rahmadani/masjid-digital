// Single place that reads import.meta.env, so the rest of the app gets typed values.
export const env = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? '/api',
  /** Shown per TV on the Layar TV page; set by the build, e.g. a release tag or commit. */
  appVersion: import.meta.env.VITE_APP_VERSION ?? 'dev',
} as const
