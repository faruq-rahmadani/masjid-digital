import { z } from 'zod'

// Mirrors the Go validation (internal/domain/prayer_time.go), so most errors show before a request.

const clock = z.string().regex(/^([01]\d|2[0-3]):[0-5]\d$/, 'Isi jam, contoh 12:00')
const remarks = z.string().trim().max(100, 'Keterangan maksimal 100 karakter')
const prayer = z.enum(['subuh', 'dzuhur', 'ashar', 'maghrib', 'isya'])

export const overrideSchema = z.object({
  date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Pilih tanggal'),
  prayer,
  time: clock,
  remarks,
})

export type OverrideForm = z.infer<typeof overrideSchema>

export const weeklyRuleSchema = z.object({
  weekday: z.enum(['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday']),
  prayer,
  time: clock,
  remarks,
})

export type WeeklyRuleForm = z.infer<typeof weeklyRuleSchema>
