/*
 * Test fakes for the service modules, wired in src/test/setup.ts. Stateful and strict like the Go
 * API (validation, duplicates, own-role rule, error codes) so page tests exercise real flows,
 * independent of the quick mocks inside src/services.
 */
import { ApiError } from '@/api/client'
import type { LoginRequest, Role } from '@/types/auth'
import type { FridayDuty, SaveFridayDutyRequest } from '@/types/fridayDuty'
import type { IqamahSettings } from '@/types/iqamah'
import type { MosqueProfile, UpdateMosqueProfileRequest } from '@/types/mosque'
import type {
  CreatePrayerOverrideRequest,
  PrayerSettings,
  PrayerTimeOverride,
} from '@/types/prayer'
import type { RunningText, RunningTextSetting, UpdateRunningTextRequest } from '@/types/runningText'
import type { Screen, ScreenCommand, UpdateScreenRequest } from '@/types/screen'
import type { CreateSlideRequest, Slide, UpdateSlideRequest } from '@/types/slide'
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
export const fakeCalls = {
  listUsers: 0,
  /** Commands sent to the TVs, e.g. ['reload:1', 'test_sound:all']. */
  screenCommands: [] as string[],
}

export function resetFakeApi() {
  users = seed()
  fakeCalls.listUsers = 0
  mosqueProfile = SEED_MOSQUE
  fakeServerClock.aheadMs = 0
  mosqueLogo = null
  prayerSettings = SEED_PRAYER_SETTINGS()
  prayerOverrides = SEED_PRAYER_OVERRIDES()
  runningTexts = SEED_RUNNING_TEXTS()
  runningTextSetting = { speed: 'medium' }
  iqamahSettings = SEED_IQAMAH_SETTINGS()
  fridayDuties = SEED_FRIDAY_DUTIES()
  slides = SEED_SLIDES()
  screens = SEED_SCREENS()
  fakeCalls.screenCommands = []
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

const SEED_RUNNING_TEXTS = (): RunningText[] =>
  [
    'Mohon matikan nada dering HP selama di dalam masjid',
    "Kajian Tafsir setiap Kamis ba'da Maghrib",
    'Infaq via QRIS tersedia di dekat pintu utama',
    'Pendaftaran qurban dibuka sampai 20 Mei',
  ].map((text, i) => ({
    id: i + 1,
    text,
    active: i < 3,
    position: i,
    created_by: 1,
    created_by_name: 'Ahmad Fauzi',
    deactivated_at: i < 3 ? null : SEED_TIME,
    created_at: SEED_TIME,
    updated_at: SEED_TIME,
  }))

let runningTexts = SEED_RUNNING_TEXTS()
let runningTextSetting: RunningTextSetting = { speed: 'medium' }

/** Read or replace the fake's stored running texts (in display order) and scroll speed. */
export const fakeRunningText = {
  get texts() {
    return runningTexts
  },
  set texts(value: RunningText[]) {
    runningTexts = value
  },
  get setting() {
    return runningTextSetting
  },
}

function validRunningText(text: string) {
  const trimmed = text.trim()
  if (trimmed === '') fail(400, 'INVALID_INPUT', 'teks wajib diisi')
  if ([...trimmed].length > 200) fail(400, 'INVALID_INPUT', 'teks maksimal 200 karakter')
  return trimmed
}

const findRunningText = (id: number) =>
  runningTexts.find((t) => t.id === id) ?? fail(404, 'NOT_FOUND', 'running text tidak ditemukan')

export const fakeRunningTextService = {
  listRunningTexts: async () => {
    requireRole('super_admin', 'takmir', 'bendahara')
    return runningTexts
  },
  createRunningText: async (text: string) => {
    const session = requireRole('super_admin', 'takmir')
    const now = new Date().toISOString()
    const created: RunningText = {
      id: Math.max(0, ...runningTexts.map((t) => t.id)) + 1,
      text: validRunningText(text),
      active: true,
      position: Math.max(-1, ...runningTexts.map((t) => t.position)) + 1,
      created_by: session.userId,
      created_by_name: findUser(session.userId).name,
      deactivated_at: null,
      created_at: now,
      updated_at: now,
    }
    runningTexts = [...runningTexts, created]
    return created
  },
  updateRunningText: async (id: number, body: UpdateRunningTextRequest) => {
    requireRole('super_admin', 'takmir')
    const current = findRunningText(id)
    const now = new Date().toISOString()
    const saved = {
      ...current,
      text: validRunningText(body.text),
      active: body.active,
      // stamped when turned off, kept while only edited, cleared when turned on
      deactivated_at: body.active ? null : (current.deactivated_at ?? now),
      updated_at: now,
    }
    runningTexts = runningTexts.map((t) => (t.id === id ? saved : t))
    return saved
  },
  deleteRunningText: async (id: number) => {
    requireRole('super_admin', 'takmir')
    findRunningText(id)
    runningTexts = runningTexts.filter((t) => t.id !== id)
    return { message: 'running text dihapus' }
  },
  reorderRunningTexts: async (ids: number[]) => {
    requireRole('super_admin', 'takmir')
    const sameSet =
      ids.length === runningTexts.length &&
      new Set(ids).size === ids.length &&
      ids.every((id) => runningTexts.some((t) => t.id === id))
    if (!sameSet)
      fail(
        400,
        'INVALID_INPUT',
        'daftar running text sudah berubah, muat ulang halaman lalu urutkan lagi',
      )
    runningTexts = ids.map((id, position) => ({ ...findRunningText(id), position }))
    return runningTexts
  },
  getRunningTextSetting: async () => {
    requireRole('super_admin', 'takmir', 'bendahara')
    return runningTextSetting
  },
  updateRunningTextSetting: async (body: RunningTextSetting) => {
    requireRole('super_admin', 'takmir')
    if (!['slow', 'medium', 'fast'].includes(body.speed))
      fail(400, 'INVALID_INPUT', 'kecepatan gulir tidak valid')
    runningTextSetting = body
    return runningTextSetting
  },
  getPublicRunningTexts: async () => ({
    speed: runningTextSetting.speed,
    texts: runningTexts.filter((t) => t.active).map((t) => t.text),
  }),
}

const SEED_SLIDES = (): Slide[] =>
  (
    [
      ['Kajian Rutin Ahad Pagi', 'Ahad, 08.00 WIB · Ustadz Abdullah', 10, true],
      ['Infaq Renovasi Atap', 'Transfer ke BSI 7123456789 a.n. Masjid Al-Ikhlas', 15, true],
      ['Foto Kegiatan Santri TPA', '', 10, true],
      ['Poster Idul Adha 1447', '', 10, false],
    ] as const
  ).map(([title, caption, duration_seconds, active], i) => ({
    id: i + 1,
    title,
    caption,
    duration_seconds,
    show_caption: true,
    active,
    position: i,
    created_by: 1,
    created_by_name: 'Ahmad Fauzi',
    created_at: SEED_TIME,
    updated_at: SEED_TIME,
  }))

let slides = SEED_SLIDES()

/** Read or replace the fake's stored slides, in display order. */
export const fakeSlide = {
  get slides() {
    return slides
  },
  set slides(value: Slide[]) {
    slides = value
  },
}

const findSlide = (id: number) =>
  slides.find((s) => s.id === id) ?? fail(404, 'NOT_FOUND', 'slide tidak ditemukan')

export const fakeSlideService = {
  slideImageUrl: (id: number) => `/fake/slides/${id}/image`,
  listSlides: async () => {
    requireRole('super_admin', 'takmir', 'bendahara')
    return slides
  },
  createSlide: async (file: File, body: CreateSlideRequest) => {
    const session = requireRole('super_admin', 'takmir')
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type))
      fail(400, 'INVALID_INPUT', 'format gambar harus PNG, JPG, atau WebP')
    if (file.size > 5 * 1024 * 1024) fail(400, 'INVALID_INPUT', 'ukuran gambar maksimal 5 MB')
    const now = new Date().toISOString()
    const created: Slide = {
      id: Math.max(0, ...slides.map((s) => s.id)) + 1,
      title: body.title.trim() || file.name.replace(/\.[^.]*$/, '').slice(0, 100),
      caption: body.caption.trim(),
      duration_seconds: body.duration_seconds,
      show_caption: body.show_caption,
      active: true,
      position: Math.max(-1, ...slides.map((s) => s.position)) + 1,
      created_by: session.userId,
      created_by_name: findUser(session.userId).name,
      created_at: now,
      updated_at: now,
    }
    slides = [...slides, created]
    return created
  },
  updateSlide: async (id: number, body: UpdateSlideRequest) => {
    requireRole('super_admin', 'takmir')
    const title = body.title.trim()
    if (title === '') fail(400, 'INVALID_INPUT', 'judul wajib diisi')
    if (![5, 10, 15, 20, 30].includes(body.duration_seconds))
      fail(400, 'INVALID_INPUT', 'durasi harus 5, 10, 15, 20, atau 30 detik')
    const saved = {
      ...findSlide(id),
      ...body,
      title,
      caption: body.caption.trim(),
      updated_at: new Date().toISOString(),
    }
    slides = slides.map((s) => (s.id === id ? saved : s))
    return saved
  },
  deleteSlide: async (id: number) => {
    requireRole('super_admin', 'takmir')
    findSlide(id)
    slides = slides.filter((s) => s.id !== id)
    return { message: 'slide dihapus' }
  },
  reorderSlides: async (ids: number[]) => {
    requireRole('super_admin', 'takmir')
    const sameSet =
      ids.length === slides.length &&
      new Set(ids).size === ids.length &&
      ids.every((id) => slides.some((s) => s.id === id))
    if (!sameSet)
      fail(400, 'INVALID_INPUT', 'daftar slide sudah berubah, muat ulang halaman lalu urutkan lagi')
    slides = ids.map((id, position) => ({ ...findSlide(id), position }))
    return slides
  },
  getPublicSlides: async () =>
    slides
      .filter((s) => s.active)
      .map(({ id, title, caption, duration_seconds, show_caption }) => ({
        id,
        title,
        caption,
        duration_seconds,
        show_caption,
      })),
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
  listQuotes: async () => [
    { text: 'Maka sesungguhnya bersama kesulitan ada kemudahan.', source: 'QS. Al-Insyirah: 5' },
  ],
}

// Seeded for the TV's fixed timelines: Maghrib iqamah 7 min after a 1-min adzan screen, 10 min of
// mode sholat; Friday 2026-10-30 has its petugas Jumat.
const SEED_IQAMAH_SETTINGS = (): IqamahSettings => ({
  iqamah_min: { subuh: 10, dzuhur: 7, ashar: 7, maghrib: 7, isya: 7 },
  sholat_min: { subuh: 10, dzuhur: 10, ashar: 10, maghrib: 10, isya: 10, jumat: 10 },
  adzan_display_min: 1,
  sholat_display: 'dim',
  jumat_before_min: 60,
  khutbah_min: 30,
  alarm: { enabled: true, at_adzan: true, before_iqamah: true, sound: 'beep', volume: 70 },
})
const SEED_FRIDAY_DUTIES = (): FridayDuty[] => [
  {
    date: '2026-10-30',
    khatib: 'Ustadz Abdullah',
    imam: 'Ustadz Hasan',
    muadzin: 'Bilal Ahmad',
    theme: 'Menjaga amanah',
    updated_by: 1,
    created_at: SEED_TIME,
    updated_at: SEED_TIME,
  },
]

let iqamahSettings = SEED_IQAMAH_SETTINGS()
let fridayDuties = SEED_FRIDAY_DUTIES()

/** Read or replace the fake's stored iqamah settings. */
export const fakeIqamah = {
  get settings() {
    return iqamahSettings
  },
  set settings(value: IqamahSettings) {
    iqamahSettings = value
  },
}

/** Read or replace the fake's stored petugas Jumat. */
export const fakeFridayDuties = {
  get list() {
    return fridayDuties
  },
  set list(value: FridayDuty[]) {
    fridayDuties = value
  },
}

const between = (v: number, lo: number, hi: number) => v >= lo && v <= hi

export const fakeIqamahService = {
  getIqamahSettings: async () => {
    requireRole('super_admin', 'takmir', 'bendahara')
    return iqamahSettings
  },
  updateIqamahSettings: async (body: IqamahSettings) => {
    requireRole('super_admin', 'takmir')
    if (!Object.values(body.iqamah_min).every((v) => between(v, 1, 30)))
      fail(400, 'INVALID_INPUT', 'iqamah harus 1 sampai 30 menit')
    const { jumat, ...sholat } = body.sholat_min
    if (!Object.values(sholat).every((v) => between(v, 5, 30)))
      fail(400, 'INVALID_INPUT', 'mode sholat harus 5 sampai 30 menit')
    if (!between(jumat, 5, 60))
      fail(400, 'INVALID_INPUT', 'mode sholat Jumat harus 5 sampai 60 menit')
    iqamahSettings = body
    return iqamahSettings
  },
  getPublicIqamahSettings: async () => iqamahSettings,
}

export const fakeFridayDutyService = {
  listFridayDuties: async (from: string, to: string) => {
    requireRole('super_admin', 'takmir', 'bendahara')
    return fridayDuties.filter((d) => d.date >= from && d.date <= to)
  },
  saveFridayDuty: async (date: string, body: SaveFridayDutyRequest) => {
    const session = requireRole('super_admin', 'takmir')
    if (new Date(`${date}T00:00:00Z`).getUTCDay() !== 5)
      fail(400, 'INVALID_INPUT', 'petugas Jumat hanya untuk hari Jumat')
    if (!body.khatib.trim() && !body.imam.trim() && !body.muadzin.trim())
      fail(400, 'INVALID_INPUT', 'isi minimal satu petugas')
    const now = new Date().toISOString()
    const existing = fridayDuties.find((d) => d.date === date)
    const saved: FridayDuty = {
      date,
      khatib: body.khatib.trim(),
      imam: body.imam.trim(),
      muadzin: body.muadzin.trim(),
      theme: body.theme.trim(),
      updated_by: session.userId,
      created_at: existing?.created_at ?? now,
      updated_at: now,
    }
    fridayDuties = [...fridayDuties.filter((d) => d.date !== date), saved].sort((a, b) =>
      a.date.localeCompare(b.date),
    )
    return saved
  },
  deleteFridayDuty: async (date: string) => {
    requireRole('super_admin', 'takmir')
    if (!fridayDuties.some((d) => d.date === date))
      fail(404, 'NOT_FOUND', 'petugas Jumat tidak ditemukan')
    fridayDuties = fridayDuties.filter((d) => d.date !== date)
    return { message: 'petugas Jumat dihapus' }
  },
  listUpcomingFridayDuties: async () => fridayDuties,
}

// One screen of each kind the Layar TV page shows: online, offline, new (BARU) and blocked.
const SEED_SCREENS = (): Screen[] => {
  const screen = (patch: Partial<Screen> & Pick<Screen, 'id'>): Screen => ({
    device_id: `device-000${patch.id}`,
    name: '',
    name_hint: '',
    user_agent: 'Mozilla/5.0 (Linux; Android 11; TV Box) Chrome/120.0 Safari/537.36',
    forced_mode: 'auto',
    forced_until: null,
    blocked_at: null,
    first_seen_at: SEED_TIME,
    updated_at: SEED_TIME,
    status: 'offline',
    last_seen_at: null,
    app_version: '1.4.0',
    ...patch,
  })
  return [
    screen({ id: 1, name: 'TV Ruang Utama', status: 'online', last_seen_at: minutesAgo(1) }),
    screen({ id: 2, name: 'TV Teras', last_seen_at: minutesAgo(125), app_version: '1.1.4' }),
    screen({ id: 3, name_hint: 'aula', first_seen_at: minutesAgo(5), status: 'online' }),
    screen({ id: 4, name: 'TV Lama', blocked_at: minutesAgo(60 * 24) }),
  ]
}

let screens = SEED_SCREENS()

/** Read or replace the fake's stored screens. */
export const fakeScreens = {
  get list() {
    return screens
  },
  set list(value: Screen[]) {
    screens = value
  },
}

const findScreen = (id: number) =>
  screens.find((s) => s.id === id) ?? fail(404, 'NOT_FOUND', 'layar tidak ditemukan')

const saveScreen = (id: number, patch: Partial<Screen>) => {
  const saved = { ...findScreen(id), ...patch, updated_at: new Date().toISOString() }
  screens = screens.map((s) => (s.id === id ? saved : s))
  return saved
}

/** The mosque's next midnight in WIB, as the API sets it for a forced mode. */
const nextWibMidnight = () => {
  const wib = new Date(Date.now() + 7 * 3600_000)
  wib.setUTCHours(24, 0, 0, 0)
  return new Date(wib.getTime() - 7 * 3600_000).toISOString()
}

export const fakeScreenService = {
  listScreens: async () => {
    requireRole('super_admin', 'takmir')
    return screens
  },
  updateScreen: async (id: number, body: UpdateScreenRequest) => {
    requireRole('super_admin', 'takmir')
    if (!body.name.trim()) fail(400, 'INVALID_INPUT', 'nama layar wajib diisi')
    return saveScreen(id, {
      name: body.name.trim(),
      forced_mode: body.forced_mode,
      forced_until: body.forced_mode === 'auto' ? null : nextWibMidnight(),
    })
  },
  blockScreen: async (id: number) => {
    requireRole('super_admin', 'takmir')
    return saveScreen(id, { blocked_at: findScreen(id).blocked_at ?? new Date().toISOString() })
  },
  unblockScreen: async (id: number) => {
    requireRole('super_admin', 'takmir')
    return saveScreen(id, { blocked_at: null })
  },
  sendScreenCommand: async (command: ScreenCommand, id?: number) => {
    requireRole('super_admin', 'takmir')
    if (id !== undefined) findScreen(id)
    fakeCalls.screenCommands.push(`${command}:${id ?? 'all'}`)
    return { message: 'perintah dikirim ke layar' }
  },
  screenEventsUrl: (deviceId: string) => `/fake/screens/${deviceId}/events`,
}

export const fakeActivityService = {
  listRecentActivity: async () => [
    {
      id: 1,
      actor_id: 3,
      actor_name: 'Siti Aminah',
      action: 'iqamah_settings.updated',
      entity_id: 0,
      summary: 'mengubah pengaturan iqamah',
      created_at: minutesAgo(5),
    },
  ],
}
