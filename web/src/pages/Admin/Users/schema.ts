import { z } from 'zod'
import { passwordSchema } from '@/utils/auth/password'
import { ROLES } from '@/utils/auth/roles'
import { normalizePhone } from '@/utils/format/phone'

// Same limits as internal/domain/user.go, so most mistakes are caught before the request.

/** Name, phone and email: what "Ubah data" edits. */
export const userDetailsSchema = z.object({
  name: z.string().trim().min(1, 'Nama wajib diisi').max(100, 'Nama maksimal 100 karakter'),
  phone: z
    .string()
    .trim()
    .min(1, 'Nomor HP wajib diisi')
    .refine(
      (value) => normalizePhone(value) !== null,
      'Nomor HP tidak valid, contoh 0812-3456-7890',
    ),
  email: z
    .string()
    .trim()
    .toLowerCase()
    .refine((value) => value === '' || z.email().safeParse(value).success, 'Email tidak valid'),
})

export const createUserSchema = userDetailsSchema.extend({
  role: z.enum(ROLES),
  password: passwordSchema,
})

export type UserDetailsForm = z.input<typeof userDetailsSchema>
export type CreateUserForm = z.input<typeof createUserSchema>
