import { z } from 'zod'

// Mirrors the Go validation (internal/domain/iqamah.go), so most errors show before a request.

const name = z.string().trim().max(100, 'Nama maksimal 100 karakter')

export const fridayDutySchema = z
  .object({
    khatib: name,
    imam: name,
    muadzin: name,
    theme: z.string().trim().max(200, 'Tema maksimal 200 karakter'),
  })
  .refine((d) => d.khatib || d.imam || d.muadzin, {
    message: 'Isi minimal satu petugas',
    path: ['khatib'],
  })

export type FridayDutyForm = z.infer<typeof fridayDutySchema>
