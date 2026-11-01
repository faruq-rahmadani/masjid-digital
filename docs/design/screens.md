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
| Layar TV         | `admin-layar.html`        | E9-06   | super admin, takmir                  | `Card brand` + `CopyField`, `Badge` (Online/Offline/BARU), `Callout danger`, `Select`, `Button`                                                     |
| Waktu sholat     | `admin-waktu.html`        | E4-05   | super admin, takmir                  | `Table`, `Stepper sm`, `SegmentedControl split` (Hijri), `Field`, `ActionBar`                                                                       |
| Iqamah & alarm   | `admin-iqamah.html`       | E5-04   | super admin, takmir                  | `Stepper md`, `RadioCard`, `Checkbox`, `Select`, `Slider`, `ActionBar`                                                                              |
| Slide            | `admin-slide.html`        | E6-04   | super admin, takmir                  | `UploadTile`, `List` + `Thumbnail`, `Sheet` (edit), `Checkbox`, `Select`                                                                            |
| Running text     | `admin-running-text.html` | E6-05   | super admin, takmir                  | `Textarea showCount`, `Checkbox`, `Badge caps`, `SegmentedControl split`                                                                            |
| Keuangan         | `admin-keuangan.html`     | E7-06   | super admin, bendahara (takmir view) | `Card brand` + `Stat`, `Button income/expense`, `SearchInput`, `SegmentedControl pill`, `Select pill`, `List` + `Money`, `Callout`, `ConfirmDialog` |
| Tambah transaksi | `admin-transaksi.html`    | E7-07   | super admin, bendahara               | `SegmentedControl track`, `MoneyInput`, `Field`, `UploadTile tile`, `Callout`, `ActionBar`                                                          |
| Kategori         | `admin-kategori.html`     | E7-08   | super admin, bendahara               | `Tabs`, `Input`, `List`, `Button danger`, `ConfirmDialog`                                                                                           |
| Petugas Jumat    | none yet                  | E5-05   | super admin, takmir                  | Reuse the Iqamah screen style in a `Sheet`                                                                                                          |

## TV display (1920×1080, `/tv`)

Built at `/tv` (`web/src/pages/Tv/`): one 1920×1080 stage scaled to the screen, `tv-*` tokens.
Public, no login: it reads `/api/public/v1/*` (read-only twins of the admin endpoints). Live:
mosque name, address, time zone, logo. The clock runs on server time (`GET /api/public/v1/time`,
resynced every 10 min) shown in the mosque's time zone, so a TV with a wrong clock or zone still
switches modes on the minute. Mocked in
`tvService`: prayer times, kas, slides, running text, quotes, iqamah settings, petugas.

| Mode   | Mockup           | Task  | Content                                                                                                                                                                                 |
| ------ | ---------------- | ----- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Normal | `tv-main.html`   | E8-01 | Header (logo, nama, alamat) + 4-column kas; date + Hijri, clock, zone; quote; slide with caption and dots; "Menuju {waktu}" countdown; 7 prayer tiles, next one gold; INFO running text |
| Adzan  | `tv-adzan.html`  | E8-02 | "Telah masuk waktu {waktu}", time, reminders                                                                                                                                            |
| Iqamah | `tv-iqamah.html` | E8-03 | Countdown with progress bar, "dari N menit", reminders, imam and muadzin                                                                                                                |
| Sholat | `tv-sholat.html` | E8-04 | Dim text + small clock, or black                                                                                                                                                        |
| Jumat  | `tv-jumat.html`  | E8-05 | Sholat Jumat time, adzan countdown, khatib / imam / muadzin, tema khutbah, INFO bar                                                                                                     |

Mode order per waktu: normal → adzan → iqamah → sholat → normal (E8-06).
