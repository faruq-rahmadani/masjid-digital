# Screens

The 17 mockup screens, what builds them, and which components they use. Open the HTML files in
[mockups/](./mockups/) in a browser; bracketed text is placeholder data.

## Admin panel (390px, mobile first)

✓ = built (routes `/login`, `/`, `/users`, `/profile`; all paths in `web/src/route.ts`). Menu items without a screen open a "Segera hadir" page.

Navigation: **Beranda is the hub** (menu grid filtered by role). Every other screen opens from it
and has a back button in `PageHeader`. There is no tab bar.

| Screen           | Mockup                    | Task(s) | Roles that see it                    | Key components                                                                                                                                      |
| ---------------- | ------------------------- | ------- | ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| Login ✓          | `admin-login.html`        | E2-06   | all                                  | `MosqueLogo`, `Field`+`Input size="lg"`, `Checkbox`, `Button xl` (primary, outline), `Divider`                                                      |
| Beranda ✓        | `admin-beranda.html`      | E3-02   | all                                  | `HeroHeader`, `Stat`, `NavCard` + `StatusDot`, `MenuTile`, `SectionHeader`, `List`                                                                  |
| Profil masjid ✓  | `admin-profil.html`       | E3-03   | super admin, takmir                  | `UploadTile`, `Field`, `InfoTile`, map pin (Leaflet), `ActionBar`                                                                                   |
| Pengguna ✓       | `admin-user.html`         | E2-07   | super admin                          | `Avatar`, `Badge`, `ActionMenu`, `Sheet` + `RadioCard`, `ConfirmDialog`, `Field`, `Select` (form follows the API, see PRD)                          |
| Layar TV         | `admin-layar.html`        | E9-06   | super admin, takmir                  | `Card brand` + QR + `CopyField`, `Badge` (Online/Offline/BARU), `Callout danger`, `Select` (Paksa mode), `Sheet` (nama, layar diblokir), `ConfirmDialog` on every action |
| Waktu sholat     | `admin-waktu.html`        | E4-05   | super admin, takmir                  | Context card, 7-day preview (today highlighted), `Stepper sm` corrections, Pengaturan khusus (mingguan, Hijri, jadwal khusus), `ConfirmDialog` on every save/delete, `LeaveGuard` |
| Iqamah & alarm   | `admin-iqamah.html`       | E5-04   | super admin, takmir                  | Brand "Alur di layar" card, iqamah & mode sholat `Stepper sm` table, Jumat group (settings + `NavCard` to Petugas Jumat), Layar & alarm (`RadioCard`, `Checkbox`, `Select`, `Slider`), `ConfirmDialog`, `LeaveGuard` |
| Slide ✓          | `admin-slide.html`        | E6-04   | super admin, takmir                  | `UploadTile` (one file, PNG/JPG/WebP ≤ 20 MB, checked client-side) opens a "Slide baru" `Sheet` with a mandatory crop to the TV panel's shape (1312:582, react-easy-crop: drag, pinch, zoom `Slider`, caption bar laid over it; only the framed part is uploaded, as JPEG ≤ 2624 px wide), then `Checkbox` tampilkan judul & keterangan, judul from the file name, keterangan, `Select` durasi 5–30 dtk); nothing is uploaded until Simpan. `SegmentedControl pill` filter, drag to reorder, `Thumbnail` + meta (durasi · sedang tampil/nonaktif), `Button sm` Matikan / Tampilkan, edit `Sheet` (same fields) + `ConfirmDialog`. No "Tampil di": every screen shows the same slides; the image can't be replaced |
| Running text ✓   | `admin-running-text.html` | E6-05   | super admin, takmir                  | `Textarea showCount`, `SegmentedControl pill` filter (Aktif default · Nonaktif · Semua), drag to reorder (dnd-kit), `Button sm` Matikan / Tampilkan per row (action, not a checkbox), meta line (pembuat · sedang/terakhir tampil), `Sheet` (edit) + `ConfirmDialog`, `SegmentedControl split`, `Ticker preview`. No "penting" flag: order decides |
| Keuangan         | `admin-keuangan.html`     | E7-06   | super admin, bendahara (takmir view) | `Card brand` + `Stat`, `Button income/expense`, `SearchInput`, `SegmentedControl pill`, `Select pill`, `List` + `Money`, `Callout`, `ConfirmDialog` |
| Tambah transaksi | `admin-transaksi.html`    | E7-07   | super admin, bendahara               | `SegmentedControl track`, `MoneyInput`, `Field`, `UploadTile tile`, `Callout`, `ActionBar`                                                          |
| Kategori         | `admin-kategori.html`     | E7-08   | super admin, bendahara               | `Tabs`, `Input`, `List`, `Button danger`, `ConfirmDialog`                                                                                           |
| Akun             | no mockup yet             | E7-10   | super admin, bendahara (takmir view) | Accounts with balances (Tunai, bank), Mulai pembukuan with the opening balance per account                                                          |
| Tutup buku       | no mockup yet             | E7-11   | super admin, bendahara               | `Sheet`: per account computed vs counted balance, difference recorded before closing, `ConfirmDialog`                                               |
| Petugas Jumat    | `admin-petugas.html`      | E5-05   | super admin, takmir                  | Own menu. Brand "Jumat terdekat" card, next 8 Fridays (`Badge notice` "Belum diisi"), `Sheet` form, confirm before save/delete |

## TV display (1920×1080, `/tv`)

Built at `/tv` (`web/src/pages/Tv/`): one 1920×1080 stage scaled to the screen, `tv-*` tokens.
Public, no login: it reads `/api/public/v1/*` (read-only twins of the admin endpoints). Live:
mosque name, address, time zone, logo. The clock runs on server time (`GET /api/public/v1/time`,
resynced every 10 min) shown in the mosque's time zone, so a TV with a wrong clock or zone still
switches modes on the minute. Prayer times are computed on the TV from
`GET /api/public/v1/prayer-time/config` (settings plus 30 days of date overrides), the INFO bar
from `GET /api/public/v1/running-texts` (active texts plus scroll speed), the iqamah settings from
`GET /api/public/v1/iqamah/config`, petugas Jumat from `GET /api/public/v1/friday-duties`
(the next 30 days), and the slides from `GET /api/public/v1/slides` (each on for its own duration,
timed off the clock so every TV shows the same one; images at `/slides/{id}/image`, cached for
good; the image fills the panel (`object-cover`) with the title and caption on a frosted-glass bar
over its bottom, or no bar when the slide's `show_caption` is off; the admin preview shows the same). Mocked in `tvService`: kas, quotes.
Every TV shows the same thing at the same moment: data reloads on the server clock's 5-minute
marks (`useAlignedRefresh`), the slide comes from the server clock (`slideIndexAt`), slide images
are preloaded, and the INFO bar's scroll is set from the server clock (`Ticker syncTo`).

The TV also holds one live connection, `GET /api/public/v1/screens/{device_id}/events` (SSE, E9),
with a random device ID kept in localStorage and `?layar=` as its name hint. It brings the
screen's forced mode (dark → black screen, normal, Jumat; until the mosque's midnight), a
`refresh` after any admin save of TV content (so changes show within seconds), and the reload and
test-sound commands. Offline over 2 minutes shows "Offline · menampilkan data tersimpan"; a
blocked TV shows "Layar ini diblokir". The alarm sounds on the adzan screen and in the last 10 s
before iqamah, after one tap unlocks audio ("Ketuk layar sekali untuk menyalakan suara alarm").

| Mode   | Mockup           | Task  | Content                                                                                                                                                                                 |
| ------ | ---------------- | ----- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Normal | `tv-main.html`   | E8-01 | Header (logo, nama, alamat) + 4-column kas; date + Hijri, clock, zone; quote; slide with caption and dots; "Menuju {waktu}" countdown; 7 prayer tiles, next one gold; INFO running text |
| Adzan  | `tv-adzan.html`  | E8-02 | "Telah masuk waktu {waktu}", time, reminders                                                                                                                                            |
| Iqamah | `tv-iqamah.html` | E8-03 | Countdown with progress bar, "dari N menit", reminders, imam and muadzin                                                                                                                |
| Sholat | `tv-sholat.html` | E8-04 | Dim text + small clock, or black                                                                                                                                                        |
| Jumat  | `tv-jumat.html`  | E8-05 | Sholat Jumat time, adzan countdown, khatib / imam / muadzin, tema khutbah, INFO bar                                                                                                     |

Mode order per waktu: normal → adzan → iqamah → sholat → normal (E8-06).
