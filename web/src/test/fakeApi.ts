/*
 * Test fakes for the service modules, wired in src/test/setup.ts. Stateful and strict like the Go
 * API (validation, duplicates, own-role rule, error codes) so page tests exercise real flows,
 * independent of the quick mocks inside src/services.
 */
import { ApiError } from '@/api/client'
import type { LoginRequest, Role } from '@/types/auth'
import type { MosqueProfile, UpdateMosqueProfileRequest } from '@/types/mosque'
import type {
  CreatePrayerOverrideRequest,
  PrayerSettings,
  PrayerTimeOverride,
} from '@/types/prayer'
import type { IqamahSettings } from '@/types/tv'
import type { CreateUserRequest, UpdateUserRequest, User } from '@/types/user'
import { loadSession } from '@/utils/auth/token'
import { normalizePhone } from '@/utils/format/phone'

export const FAKE_PASSWORD = 'masjid123'

const SEED_TIME = '2026-09-01T08:00:00.000Z'

function seed(): User[] {
  const user = (
    u: Pick<User, 'id' | 'name' | 'phone' | 'email' | 'role' | 'status' | 'created_by'>,
  ) => ({ ...u, last_login_at: null, created_at: SEED_TIME, updated_at: SEED_TIME }) as User
  return [
    user({
      id: 1,
      name: 'Ahmad Fauzi',
      phone: '+6281234561000',
      email: 'ketua@masjid.id',
      role: 'super_admin',
      status: 'active',
      created_by: 0,
    }),
    user({
      id: 2,
      name: 'Hasan Basri',
      phone: '+6281234561001',
      email: '',
      role: 'takmir',
      status: 'active',
      created_by: 1,
    }),
    user({
      id: 3,
      name: 'Siti Aminah',
      phone: '+6281234561002',
      email: 'bendahara@masjid.id',
      role: 'bendahara',
      status: 'active',
      created_by: 1,
    }),
    user({
      id: 4,
      name: 'Umar Hadi',
      phone: '+6281234561003',
      email: '',
      role: 'takmir',
      status: 'disabled',
      created_by: 1,
    }),
  ]
}

let users = seed()

/** How often each fake was called, to assert what refetches. */
export const fakeCalls = { listUsers: 0 }

export function resetFakeApi() {
  users = seed()
  fakeCalls.listUsers = 0
  mosqueProfile = SEED_MOSQUE
  fakeServerClock.aheadMs = 0
  mosqueLogo = null
  prayerSettings = SEED_PRAYER_SETTINGS()
  prayerOverrides = SEED_PRAYER_OVERRIDES()
}

const fail = (status: number, code: string, message: string): never => {
  throw new ApiError(status, { code, message })
}

function requireRole(...roles: Role[]) {
  const session = loadSession() ?? fail(401, 'UNAUTHENTICATED', 'silakan login terlebih dahulu')
  if (!roles.includes(session.role)) fail(403, 'FORBIDDEN', 'anda tidak memiliki akses')
  return session
}

const requireSuperAdmin = () => requireRole('super_admin')

const findUser = (id: number) =>
  users.find((u) => u.id === id) ?? fail(404, 'NOT_FOUND', 'user tidak ditemukan')

function assertUnique(phone: string, email: string, exceptId?: number) {
  const taken = users.some(
    (u) => u.id !== exceptId && (u.phone === phone || (email !== '' && u.email === email)),
  )
  if (taken) fail(409, 'ALREADY_EXISTS', 'nomor HP atau email sudah terdaftar')
}

function encode(value: object) {
  return btoa(JSON.stringify(value)).replace(/=+$/, '')
}

/** Unsigned JWT with the API's claims, as a logged-in session for renderRoute. */
export function fakeToken(userId: number, role: Role) {
  const exp = Math.floor(Date.now() / 1000) + 3600
  return `${encode({ alg: 'none' })}.${encode({ sub: String(userId), role, exp })}.fake`
}

export const fakeAuthService = {
  login: async ({ identifier, password }: LoginRequest) => {
    const id = identifier.trim()
    const user = id.includes('@')
      ? users.find((u) => u.email === id.toLowerCase())
      : users.find((u) => u.phone === normalizePhone(id))
    if (!user || password !== FAKE_PASSWORD) {
      fail(401, 'UNAUTHENTICATED', 'nomor HP/email atau password salah')
    }
    if (user?.status !== 'active') fail(403, 'FORBIDDEN', 'akun dinonaktifkan')
    return { token: fakeToken((user as User).id, (user as User).role), expire_in: 3600 }
  },
  logout: async () => ({ message: 'logout berhasil' }),
  getProfile: async () => {
    const session = loadSession() ?? fail(401, 'UNAUTHENTICATED', 'silakan login terlebih dahulu')
    return findUser(session.userId)
  },
  changePassword: async (currentPassword: string, newPassword: string) => {
    loadSession() ?? fail(401, 'UNAUTHENTICATED', 'silakan login terlebih dahulu')
    if (currentPassword !== FAKE_PASSWORD) fail(400, 'INVALID_INPUT', 'password lama salah')
    if (newPassword === currentPassword) {
      fail(400, 'INVALID_INPUT', 'password baru harus berbeda dari password lama')
    }
    return { message: 'password berhasil diganti' }
  },
}

const SEED_MOSQUE: MosqueProfile = {
  name: 'Masjid Nur Husen',
  address: 'Jl. Merdeka No. 1, Jakarta',
  contact_phone: '+6281234561000',
  latitude: -6.2088,
  longitude: 106.8456,
  timezone: 'Asia/Jakarta',
}

let mosqueProfile: MosqueProfile | null = SEED_MOSQUE
let mosqueLogo: string | null = null

/** Read or replace the fake's stored mosque profile and logo, e.g. null for a fresh install. */
export const fakeMosque = {
  get profile() {
    return mosqueProfile
  },
  set profile(value: MosqueProfile | null) {
    mosqueProfile = value
  },
  get logo() {
    return mosqueLogo
  },
  set logo(value: string | null) {
    mosqueLogo = value
  },
}

export const fakeMosqueProfileService = {
  getMosqueProfile: async () =>
    mosqueProfile ?? fail(404, 'NOT_FOUND', 'profil masjid belum diisi'),
  updateMosqueProfile: async (body: UpdateMosqueProfileRequest) => {
    requireRole('super_admin', 'takmir')
    if (body.name.trim() === '') fail(400, 'INVALID_INPUT', 'nama wajib di isi')
    if (body.latitude === 0 && body.longitude === 0) {
      fail(400, 'INVALID_INPUT', 'lokasi masjid wajib di isi')
    }
    const contact_phone =
      body.contact_phone === ''
        ? ''
        : (normalizePhone(body.contact_phone) ??
          fail(400, 'INVALID_INPUT', 'kontak pengurus harus nomor HP yang valid'))
    mosqueProfile = { ...body, name: body.name.trim(), contact_phone }
    return mosqueProfile
  },
  getMosqueLogo: async () => ({
    url: mosqueLogo ?? fail(404, 'NOT_FOUND', 'logo masjid belum diunggah'),
  }),
  // No login needed, like /api/public/v1.
  getPublicMosqueProfile: async () =>
    mosqueProfile ?? fail(404, 'NOT_FOUND', 'profil masjid belum diisi'),
  getPublicMosqueLogo: async () => ({
    url: mosqueLogo ?? fail(404, 'NOT_FOUND', 'logo masjid belum diunggah'),
  }),
  uploadMosqueLogo: async (file: File) => {
    requireRole('super_admin', 'takmir')
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) {
      fail(400, 'INVALID_INPUT', 'format logo harus PNG, JPG, atau WebP')
    }
    mosqueLogo = `data:${file.type};base64,AAAA`
    return { message: 'logo masjid berhasil disimpan' }
  },
  deleteMosqueLogo: async () => {
    requireRole('super_admin', 'takmir')
    if (!mosqueLogo) fail(404, 'NOT_FOUND', 'logo masjid belum diunggah')
    mosqueLogo = null
    return { message: 'logo masjid berhasil dihapus' }
  },
}

/** Any query finds Masjid Raya Makassar (WITA), except "tidak ada" which finds nothing. */
export const fakeGeocodingService = {
  searchPlaces: async (query: string) =>
    query === 'tidak ada'
      ? []
      : [
          {
            name: 'Masjid Raya Makassar, Jalan Masjid Raya',
            latitude: -5.1336,
            longitude: 119.4183,
          },
        ],
}

export const fakeUserService = {
  listUsers: async () => {
    fakeCalls.listUsers++
    requireSuperAdmin()
    return users
  },
  createUser: async ({ password: _password, ...body }: CreateUserRequest) => {
    const session = requireSuperAdmin()
    const phone = normalizePhone(body.phone) ?? fail(400, 'INVALID_INPUT', 'nomor HP tidak valid')
    assertUnique(phone, body.email)
    const id = Math.max(...users.map((u) => u.id)) + 1
    const now = new Date().toISOString()
    users = [
      ...users,
      {
        ...body,
        id,
        phone,
        status: 'active',
        last_login_at: null,
        created_by: session.userId,
        created_at: now,
        updated_at: now,
      },
    ]
    return { message: 'user berhasil dibuat' }
  },
  updateUser: async (id: number, body: UpdateUserRequest) => {
    const session = requireSuperAdmin()
    const user = findUser(id)
    if (session.userId === id && body.role !== user.role) {
      fail(403, 'FORBIDDEN', 'tidak bisa mengubah role sendiri')
    }
    assertUnique(body.phone, body.email, id)
    const updated = { ...user, ...body }
    users = users.map((u) => (u.id === id ? updated : u))
    return updated
  },
  resetUserPassword: async (id: number) => {
    requireSuperAdmin()
    findUser(id)
    return { message: 'password berhasil direset' }
  },
}

const minutesAgo = (m: number) => new Date(Date.now() - m * 60_000).toISOString()

// Seeded so the TV's timeline is fixed: Wednesday 2026-10-28 Maghrib at 18:15 (a date override)
// and Friday Dzuhur at 12:09 (a weekly rule); every other time is computed for the seed mosque.
const SEED_PRAYER_SETTINGS = (): PrayerSettings => ({
  corrections: { subuh: 0, terbit: 0, dzuhur: 0, ashar: 0, maghrib: 0, isya: 0 },
  imsak_offset_min: 10,
  weekly_overrides: [
    { weekday: 'friday', prayer: 'dzuhur', time: '12:09', remarks: 'Sholat Jumat' },
  ],
  hijri_offset_days: 0,
  hijri_offset_from: '',
})
const SEED_PRAYER_OVERRIDES = (): PrayerTimeOverride[] => [
  {
    id: 1,
    date: '2026-10-28',
    prayer: 'maghrib',
    time: '18:15',
    remarks: 'Pengajian akbar',
    created_by: 1,
    created_at: SEED_TIME,
    updated_at: SEED_TIME,
  },
]

let prayerSettings = SEED_PRAYER_SETTINGS()
let prayerOverrides = SEED_PRAYER_OVERRIDES()

/** Read or replace the fake's stored prayer settings and date overrides. */
export const fakePrayer = {
  get settings() {
    return prayerSettings
  },
  set settings(value: PrayerSettings) {
    prayerSettings = value
  },
  get overrides() {
    return prayerOverrides
  },
  set overrides(value: PrayerTimeOverride[]) {
    prayerOverrides = value
  },
}

const CLOCK = /^([01]\d|2[0-3]):[0-5]\d$/

export const fakePrayerTimeService = {
  getPrayerSettings: async () => {
    requireRole('super_admin', 'takmir', 'bendahara')
    return prayerSettings
  },
  updatePrayerSettings: async (body: PrayerSettings) => {
    requireRole('super_admin', 'takmir')
    const corrections = Object.values(body.corrections)
    if (corrections.some((c) => c < -10 || c > 10))
      fail(400, 'INVALID_INPUT', 'koreksi waktu harus antara -10 dan +10 menit')
    if (body.imsak_offset_min < 1 || body.imsak_offset_min > 30)
      fail(400, 'INVALID_INPUT', 'imsak harus 1 sampai 30 menit sebelum subuh')
    const keys = body.weekly_overrides.map((o) => `${o.weekday}/${o.prayer}`)
    if (new Set(keys).size !== keys.length)
      fail(
        400,
        'INVALID_INPUT',
        'satu waktu sholat hanya boleh punya satu jadwal mingguan per hari',
      )
    if (body.weekly_overrides.some((o) => !CLOCK.test(o.time)))
      fail(400, 'INVALID_INPUT', 'format jam harus HH:MM')
    prayerSettings = body
    return prayerSettings
  },
  listPrayerOverrides: async (from: string, to: string) => {
    requireRole('super_admin', 'takmir', 'bendahara')
    return prayerOverrides
      .filter((o) => o.date >= from && o.date <= to)
      .sort((a, b) => a.date.localeCompare(b.date) || a.prayer.localeCompare(b.prayer))
  },
  createPrayerOverride: async (body: CreatePrayerOverrideRequest) => {
    const session = requireRole('super_admin', 'takmir')
    if (!CLOCK.test(body.time)) fail(400, 'INVALID_INPUT', 'format jam harus HH:MM')
    if (body.date < new Date().toISOString().slice(0, 10))
      fail(400, 'INVALID_INPUT', 'tanggal sudah lewat')
    if (prayerOverrides.some((o) => o.date === body.date && o.prayer === body.prayer))
      fail(409, 'ALREADY_EXISTS', 'waktu sholat ini sudah punya jadwal khusus di tanggal tersebut')
    const now = new Date().toISOString()
    const created: PrayerTimeOverride = {
      ...body,
      remarks: body.remarks.trim(),
      id: Math.max(0, ...prayerOverrides.map((o) => o.id)) + 1,
      created_by: session.userId,
      created_at: now,
      updated_at: now,
    }
    prayerOverrides = [...prayerOverrides, created]
    return created
  },
  deletePrayerOverride: async (id: number) => {
    requireRole('super_admin', 'takmir')
    if (!prayerOverrides.some((o) => o.id === id))
      fail(404, 'NOT_FOUND', 'jadwal khusus tidak ditemukan')
    prayerOverrides = prayerOverrides.filter((o) => o.id !== id)
    return { message: 'jadwal khusus dihapus' }
  },
  getPublicPrayerConfig: async () => ({ settings: prayerSettings, overrides: prayerOverrides }),
}

/** How far the fake server's clock runs ahead of the (fake) device clock, in ms. */
export const fakeServerClock = { aheadMs: 0 }

export const fakeTimeService = {
  getServerTime: async () => Date.now() + fakeServerClock.aheadMs,
}

export const fakeTvService = {
  getCashSummary: async () => ({
    as_of: '2026-10-28',
    opening_balance: 91_934_781,
    income: 1_250_000,
    expense: 1_126_667,
    closing_balance: 92_058_114,
  }),
  listSlides: async () => [
    { id: 1, title: 'Kajian Rutin Ahad Pagi', caption: 'Ahad, 08.00 WIB', image_url: '' },
    { id: 2, title: 'Kajian Tafsir', caption: "Kamis ba'da Maghrib", image_url: '' },
  ],
  listRunningTexts: async () => ['Infaq via QRIS tersedia di dekat pintu utama'],
  listQuotes: async () => [
    { text: 'Maka sesungguhnya bersama kesulitan ada kemudahan.', source: 'QS. Al-Insyirah: 5' },
  ],
  getIqamahSettings: async (): Promise<IqamahSettings> => ({
    iqamah_minutes: { subuh: 10, dzuhur: 7, ashar: 7, maghrib: 7, isya: 7 },
    adzan_minutes: 1,
    prayer_minutes: 10,
    jumat_before_minutes: 60,
    khutbah_minutes: 30,
  }),
  getDutyRoster: async () => ({
    imam: 'Ustadz Hasan',
    muadzin: 'Bilal Ahmad',
    khatib: 'Ustadz Abdullah',
    khutbah_theme: 'Menjaga amanah',
  }),
}

export const fakeScreenService = {
  listScreens: async () => [
    { id: 1, name: 'TV Ruang Utama', status: 'online' as const, last_seen_at: minutesAgo(1) },
    { id: 2, name: 'TV Teras', status: 'offline' as const, last_seen_at: minutesAgo(125) },
  ],
}

export const fakeActivityService = {
  listRecentActivity: async () => [
    {
      id: 1,
      actor_name: 'Siti Aminah',
      summary: 'menambah pemasukan Rp 750.000',
      created_at: minutesAgo(5),
    },
  ],
}
