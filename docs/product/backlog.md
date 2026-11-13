# Backlog

Estimates are developer-days. Task IDs (`E7-06`) are referenced from PRs, commits and
[../design/screens.md](../design/screens.md). Roadmap: 7 weeks build + 1 week pilot, with one API
and one UI developer in parallel.

## Roadmap

| Weeks | Phase               | Epics      | Exit criterion                                          |
| ----- | ------------------- | ---------- | ------------------------------------------------------- |
| W1–2  | Foundation & access | E1, E2, E3 | Admin logs in and sees only their role's menus          |
| W3–5  | Time & kas core     | E4, E5, E7 | One real month re-entered; every account and every kas balance matches |
| W4–6  | TV display          | E8, E6     | All 5 modes switch on schedule                          |
| W6–7  | TV ops & QA         | E9, E10    | 12 h on a TV box, offline for 1 h, no reload            |
| W8    | Pilot               | —          | 3 mosques, 2 Fridays each, all announce kas from the TV |

## Epics

| Epic | Name                        | Summary                                                    |
| ---- | --------------------------- | ---------------------------------------------------------- |
| E1   | Fondasi & design system     | Repo, tokens and the shared component kit                  |
| E2   | Akun & peran                | Login, three roles, users and activity log                 |
| E3   | Beranda & profil masjid     | Role-based home and the mosque profile                     |
| E4   | Waktu sholat & Hijriah      | On-device Kemenag engine, corrections, overrides           |
| E5   | Iqamah, mode sholat & alarm | Per-waktu durations, alarm and petugas Jumat               |
| E6   | Konten tayang               | Slides, running text and quotes                            |
| E7   | Keuangan (akun × kas)       | Kas (kantong) with syariah rules, accounts, categories, transactions, pindah dana, tutup buku, TV summary |
| E8   | Layar TV: tampilan & mode   | The 1920×1080 display and its five modes                   |
| E9   | Layar TV: operasional       | URL registration, realtime, offline, admin screen list     |
| E10  | QA & pilot                  | States, device tests, isolation tests, pilot               |
| E11  | Web publik & laporan        | Public pages and reports (P1)                              |
| E12  | Konten & jadwal lanjutan    | Quotes, agenda, scheduling, Ramadhan (P1)                  |
| E13  | Keuangan lanjutan           | Programs, donors, rekening, dana terikat, donation (P1–P2) |
| E14  | Platform & SaaS             | OTP, billing, monitoring, warga, i18n (P1–P2)              |

## P0 tasks

| ID     | Type | Days | Task                                                                                                                       | Screen             |
| ------ | ---- | ---- | -------------------------------------------------------------------------------------------------------------------------- | ------------------ |
| E1-01  | API  | 1    | Repo, CI/CD, env config, lint, test runner                                                                                 |                    |
| E1-02  | API  | 1    | DB schema, migrations and seed (demo mosque, default categories, quotes)                                                   |                    |
| E1-03  | API  | 0    | Dropped: single-tenant, the TV is `/tv` (no slug resolver)                                                                 |                    |
| E1-04  | API  | 1    | Object storage for images with resize to 1920 px; logging, error tracking, daily DB backup                                 |                    |
| E1-05  | Ops  | 1    | Staging + production deploy                                                                                                |                    |
| E1-06  | UI   | 1.5  | Design tokens from the mockup + Rupiah formatter                                                                           |                    |
| E1-07  | UI   | 2    | Component kit: buttons, inputs, segmented control, − / + stepper, toggle, chips, list row with ⋯ menu, bottom sheet, toast |                    |
| E1-08  | UI   | 1    | Admin shell 390 px: header, back nav, "Simpan dan kirim ke layar" bar                                                      |                    |
| E2-01  | API  | 1    | Login with email or HP + password, remember me, sessions                                                                   |                    |
| E2-02  | API  | 1    | Forgot password and reset link by email / WhatsApp link                                                                    |                    |
| E2-03  | API  | 1.5  | Roles super admin / takmir / bendahara + permission guard per menu                                                         |                    |
| E2-04  | API  | 1.5  | Users: invite, change role, send reset link, deactivate; keep one active super admin                                       |                    |
| E2-05  | API  | 1    | Activity log written on every change                                                                                       |                    |
| E2-06  | UI   | 0.5  | Login screen                                                                                                               | admin-login        |
| E2-07  | UI   | 1.5  | Pengguna screen: list, status, ⋯ actions, add user form                                                                    | admin-user         |
| E3-01  | API  | 1    | Mosque profile CRUD: logo ≤2 MB square, nama, alamat, kontak, koordinat, zona waktu                                        |                    |
| E3-02  | UI   | 1.5  | Beranda: greeting + role, next prayer card, TV status card, role menu grid, aktivitas terakhir                             | admin-beranda      |
| E3-03  | UI   | 1.5  | Profil masjid with Leaflet map pin and save-to-screens                                                                     | admin-profil       |
| E4-01  | API  | 1.5  | Shared prayer-time engine (adhan JS, Kemenag 20°/18°) used by API and TV                                                   |                    |
| E4-02  | API  | 0.5  | Settings: per-waktu correction −10..+10, Imsak offset                                                                      |                    |
| E4-03  | API  | 0.5  | Date overrides (e.g. Jumat Dzuhur 11:58)                                                                                   |                    |
| E4-04  | API  | 0.5  | Hijri correction −2..+2 with effective date                                                                                |                    |
| E4-05  | UI   | 2    | Waktu sholat screen: table with steppers, 7-day preview, overrides, Hijri control                                          | admin-waktu        |
| E5-01  | API  | 1    | Iqamah 1–30, mode sholat 5–30 (Jumat ≤60), adzan display time, sholat display, alarm, sound, volume                        |                    |
| E5-02  | API  | 0.5  | "Uji bunyi" command to all screens (`POST /screens/commands`, button on Iqamah & Alarm)                                   |                    |
| E5-03  | API  | 0.5  | Petugas Jumat per date: khatib, imam, muadzin, tema                                                                        |                    |
| E5-04  | UI   | 1.5  | Iqamah & Alarm screen                                                                                                      | admin-iqamah       |
| E5-05  | UI   | 0.5  | Petugas Jumat sheet (no mockup yet; reuse Iqamah screen style)                                                             |                    |
| E6-01  | API  | 1.5  | Slides: multi upload JPG/PNG/WebP ≤5 MB, judul, keterangan, durasi, target screens, aktif, order                           |                    |
| E6-02  | API  | 0.5  | Running text ≤200 chars, penting flag, order, scroll speed                                                                 |                    |
| E6-03  | API  | 0.5  | Built-in quote list (ayat/hadits) served to TV                                                                             |                    |
| E6-04  | UI   | 2    | Slide screen: upload, list, drag reorder, edit sheet, delete                                                               | admin-slide        |
| E6-05  | UI   | 1.5  | Running text screen: add with counter, penting toggle, reorder, speed, preview                                             | admin-running-text |
| E7-13  | API  | 1.5  | Kas model: kinds and their rules in code; seed Kas Umum, Zakat Mal, Zakat Fitrah, Fidyah (locked); users create only Terikat kas with nama, peruntukan, target, rencana sisa dana, tampil di TV; archive only at Rp 0 |  |
| E7-01  | API  | 1    | Categories scoped by kas kind; system categories in code (zakat in, 8 asnaf, fidyah in, fakir/miskin for fidyah, Pindah dana); own categories only for Umum and Terikat; rename, delete if unused, archive / restore |  |
| E7-20  | API  | 0.5  | Syariah warnings: amil spending above 1/8 of zakat income this Hijri year; Zakat Fitrah balance left from 1 Syawal |  |
| E7-02  | API  | 2    | Transactions in / out with account + kas: integer Rp, date, category, note, photo; edit/delete need reason (revisions); "Diedit"; soft delete; correction entries; no account or kas below Rp 0 |  |
| E7-15  | API  | 1    | Split income: one receipt saved as linked entries in several kas (group_id, edited and deleted together) |  |
| E7-14  | API  | 1    | Pindah dana as a linked out + in pair (group_id, system category, out of the totals); nothing into or out of zakat/fidyah, terikat needs reason + donor consent, Umum → terikat free |  |
| E7-16  | API  | 1    | Target & close: progress, "Target tercapai"; "Selesaikan program" posts the remainder to Kas Umum (consent + note when rencana sisa dana is ask_donor) and archives the kas |  |
| E7-09  | API  | 1    | Accounts (Tunai / bank): add, rename, archive; Mulai pembukuan with the opening balance per account, split over kas |  |
| E7-03  | API  | 1    | Tutup buku in month order with a counted balance per account; keeps month totals and per-kas balances; locks the month |  |
| E7-04  | API  | 1    | Summaries computed in the usecase: per account, per kas, rekap; masuk/keluar today and period; TV payload with the kas marked tampil di TV |  |
| E7-05  | API  | 0.5  | Server validation: max Rp 10.000.000.000 per transaction, confirm above Rp 10.000.000 |  |
| E7-06  | UI   | 2    | Keuangan screen: kas chips, balance card with type badge and purpose, + / − / Pindah dana (zakat note), filters incl. Pindah, month, list, show deleted, tutup buku | kas-simulation |
| E7-07  | UI   | 2    | Form transaksi: kas with purpose hint, akun, categories by kas kind, clean amount input, photo; "Bagi ke beberapa kas" split mode | kas-simulation |
| E7-17  | UI   | 1    | One Pindah dana form (dari kas + akun, ke kas + akun; covers setor tunai): saldo hints, blocked zakat options with reason, donor-consent checkbox for terikat | kas-simulation |
| E7-18  | UI   | 1.5  | Kas list (built-in kas locked) and Kas baru / Ubah kas sheets for Terikat: peruntukan, target, rencana sisa dana, tampil di TV, archive at Rp 0 | kas-simulation |
| E7-21  | UI   | 0.5  | Warning banners for the amil cap and the Zakat Fitrah deadline on the kas card and Beranda |  |
| E7-19  | UI   | 1    | Target progress bar and "Selesaikan program" confirm showing the remainder going to Kas Umum (no mockup yet) |  |
| E7-10  | UI   | 1    | Akun screen with balances, Mulai pembukuan with per-kas opening split (no mockup yet) |  |
| E7-11  | UI   | 1    | Tutup buku sheet: computed vs counted balance per account, kas balances listed (no mockup yet) |  |
| E7-08  | UI   | 1    | Kategori screen per kas kind incl. asnaf: tabs, add, counts, rename / delete / archive, restore | kas-simulation |
| E8-01  | UI   | 3    | TV normal layout: header + kas, date & Hijri, clock, quote, slides, countdown, 7 tiles, INFO bar                           | tv-main            |
| E8-02  | UI   | 0.5  | Mode adzan                                                                                                                 | tv-adzan           |
| E8-03  | UI   | 0.5  | Iqamah countdown with reminders and petugas                                                                                | tv-iqamah          |
| E8-04  | UI   | 0.5  | Mode sholat: dim text or black                                                                                             | tv-sholat          |
| E8-05  | UI   | 1    | Mode Jumat with khatib, imam, muadzin and tema                                                                             | tv-jumat           |
| E8-06  | UI   | 1.5  | Mode scheduler (normal → adzan → iqamah → sholat), Jumat rules, alarm with tap-to-enable audio                             | tv-*               |
| E8-07  | UI   | 0.5  | Scale for 720p, 1080p and 4K                                                                                               |                    |
| E9-01  | API  | 0    | Dropped: the TV's small public reads stay separate; the live connection makes refreshes cheap                              |                    |
| E9-02  | API  | 1    | Auto-register on first connect: `?layar=` name hint, version; presence in the cache; block / unblock; cap 50              |                    |
| E9-03  | API  | 1.5  | Live connection (SSE, in-memory hub): state, content refresh after saves, commands reload / test sound, forced mode       |                    |
| E9-04  | UI   | 0.5  | Offline mark after 2 min; runs on in-memory data and cached images (service worker deferred until the pilot shows a need) |                    |
| E9-05  | UI   | 0.5  | Kiosk behaviour: wake lock, fullscreen on tap, auto-reload on error only when online                                       |                    |
| E9-06  | UI   | 1.5  | Layar TV admin: URL + Salin + QR, list with status, force mode, reload, rename, block, BARU, blocked list with unblock     | admin-layar        |
| E10-01 | UI   | 1    | Empty, loading and error states on every P0 screen                                                                         |                    |
| E10-02 | QA   | 1    | Device tests: Android TV box, Smart TV browser, low-end Android phone                                                      |                    |
| E10-03 | QA   | 1    | Permission tests per role and route                                                                                        |                    |
| E10-04 | Ops  | 0    | Pilot with 3 mosques for 2 Fridays (week 8)                                                                                |                    |

## P1 / P2 (after pilot)

| ID     | Pri | Type | Days | Task                                                           |
| ------ | --- | ---- | ---- | -------------------------------------------------------------- |
| E7-12  | P1  | API  | 1    | Reopen the latest tutup buku: super admin, with reason, logged |
| E11-01 | P1  | UI   | 4    | Public pages: Beranda, Laporan kas, Jadwal, Kontak             |
| E11-02 | P1  | API  | 3    | Reports + PDF export A4/Legal with signature blocks            |
| E11-03 | P1  | API  | 1    | Print kwitansi per transaction                                 |
| E11-04 | P1  | UI   | 1    | Full activity / audit log page ("Lihat semua")                 |
| E12-01 | P1  | UI   | 1.5  | Quote manager                                                  |
| E12-02 | P1  | UI   | 3    | Agenda & kajian schedule, petugas roster for all waktu         |
| E12-03 | P1  | API  | 1.5  | Content date ranges                                            |
| E12-04 | P1  | UI   | 3    | Ramadhan mode: imsakiyah, berbuka countdown, tarawih           |
| E12-05 | P1  | UI   | 2    | Theme / accent picker and portrait layout                      |
| E13-01 | P1  | API  | 3    | Programs with target and progress on TV and web                |
| E13-02 | P1  | API  | 1    | Rekening / QRIS info on TV and web                             |
| E13-03 | P1  | API  | 2    | Donor list and donor on income (default Hamba Allah)           |
| E14-01 | P1  | API  | 2    | Login with OTP WhatsApp                                        |
| E13-05 | P2  | API  | 6    | Online donation via QRIS / payment gateway + reconciliation    |
| E14-02 | P2  | API  | 6    | Super-admin SaaS console, plans and billing                    |
| E14-03 | P2  | API  | 2    | Remote monitoring with last screenshot per screen              |
| E14-04 | P2  | UI   | 1.5  | Tarhim / murottal audio before Subuh                           |
| E14-05 | P2  | UI   | 1.5  | Hari besar Islam and puasa sunnah reminders                    |
| E14-06 | P2  | API  | 4    | Warga database                                                 |
| E14-07 | P2  | UI   | 2    | Language switch id / en                                        |
