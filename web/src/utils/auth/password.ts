import { z } from 'zod'

/** Same limits as the Go API (internal/domain/auth.go: 8..128). */
export const passwordSchema = z
  .string()
  .min(8, 'Password minimal 8 karakter')
  .max(128, 'Password maksimal 128 karakter')
