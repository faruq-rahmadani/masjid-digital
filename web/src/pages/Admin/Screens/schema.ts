import { z } from 'zod'

// Mirrors the Go validation (internal/domain/screen.go), so most errors show before a request.

export const screenNameSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, 'Nama layar wajib diisi')
    .max(50, 'Nama layar maksimal 50 karakter'),
})

export type ScreenNameForm = z.infer<typeof screenNameSchema>
