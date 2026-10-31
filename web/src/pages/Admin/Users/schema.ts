import { z } from 'zod'
import { ROLES } from '@/utils/auth/roles'
import { normalizePhone } from '@/utils/format/phone'

// Same limits as internal/domain/user.go, so most mistakes are caught before the request.

export const passwordSchema = z
  .string()
  .min(8, 'Password minimal 8 karakter')
  .max(128, 'Password maksimal 128 karakter')

export const createUserSchema = z.object({
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
  role: z.enum(ROLES),
  password: passwordSchema,
})

export type CreateUserForm = z.input<typeof createUserSchema>
