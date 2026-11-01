import { z } from 'zod'
import type { MosqueProfile } from '@/types/mosque'
import { formatPhone, normalizePhone } from '@/utils/format/phone'

// Same limits as internal/domain/mosque_profile.go, so most mistakes are caught before the request.

export const profileSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, 'Nama masjid wajib diisi')
    .max(100, 'Nama masjid maksimal 100 karakter'),
  address: z.string().trim().max(500, 'Alamat maksimal 500 karakter'),
  contact_phone: z
    .string()
    .trim()
    .refine(
      (value) => value === '' || normalizePhone(value) !== null,
      'Nomor HP tidak valid, contoh 0812-3456-7890',
    ),
  // null until a point is picked: a new profile has none, and the API rejects 0,0.
  location: z
    .object({ latitude: z.number(), longitude: z.number() })
    .nullable()
    .refine((value) => value !== null, 'Tentukan lokasi masjid di peta'),
  timezone: z.enum(['Asia/Jakarta', 'Asia/Makassar', 'Asia/Jayapura']),
})

export type ProfileForm = z.input<typeof profileSchema>

export function toFormValues(profile: MosqueProfile | null): ProfileForm {
  if (!profile) {
    return { name: '', address: '', contact_phone: '', location: null, timezone: 'Asia/Jakarta' }
  }
  return {
    name: profile.name,
    address: profile.address,
    contact_phone: profile.contact_phone && formatPhone(profile.contact_phone),
    location: { latitude: profile.latitude, longitude: profile.longitude },
    timezone: profile.timezone,
  }
}
