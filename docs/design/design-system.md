# Design system

The mockup is the spec: every token and component here was taken from the 17 screens in
[mockups/](./mockups/). Tokens live in `web/src/styles/index.css`, components in
`web/src/components/` (import from `@/components`). In dev, **`/kit`** renders this whole document
live, in the same order.

**Contents:** [Principles](#principles) · [Foundations](#foundations) ·
[Components](#components) · [Patterns](#patterns) · [Content](#content) ·
[Accessibility](#accessibility) · [Changing the system](#changing-the-system)

## Principles

1. **Calm and legible.** Pengurus of every age use this on a phone, often in a hurry after
   sholat. Big targets, plain words, one primary action per screen.
2. **Money is never ambiguous.** Pemasukan is green with +, pengeluaran is rust with −, always.
3. **Gold is rare.** It marks the one thing to look at: the next prayer, a countdown, focus.
4. **Flat and quiet.** White cards on cream, no gradients, no decorative shadows, no emoji.
5. **Every save reaches the TV.** The wording says so: "Simpan dan kirim ke layar".

## Foundations

### Color

Use colors by **role**, through the token class (`bg-brand`, `text-ink-muted`, `border-line`).
Tailwind's default palette is removed, so `bg-gray-100` or a hex in a component fails the build.

| Group          | Token            | Hex       | Role                                     |
| -------------- | ---------------- | --------- | ---------------------------------------- |
| Grounds        | `canvas`         | `#F4F1E8` | Admin page background (cream)            |
|                | `card`           | `#FFFFFF` | Cards, inputs                            |
|                | `track`          | `#E6E0CE` | Segmented-control and tab track          |
|                | `highlight`      | `#F7F4EC` | Row currently being edited               |
|                | `media`          | `#DDE5E0` | Image placeholder                        |
| Text           | `ink`            | `#0F3D33` | Primary text                             |
|                | `ink-soft`       | `#4F6158` | Field labels                             |
|                | `ink-muted`      | `#5B6B64` | Secondary text, meta, help               |
| Lines          | `line`           | `#D9D2BE` | Input and button borders                 |
|                | `line-soft`      | `#ECE7DA` | Dividers between rows                    |
|                | `dash`           | `#9FB3AA` | Dashed upload borders                    |
| Brand          | `brand`          | `#0F3D33` | Primary buttons, hero header, saldo card |
|                | `brand-2`        | `#164A3E` | Cards sitting on brand                   |
|                | `on-brand`       | `#F4F1E8` | Text on brand                            |
|                | `on-brand-muted` | `#C9D6CF` | Secondary text on brand                  |
| Accent         | `gold`           | `#C9A227` | Next prayer, countdown, focus ring       |
|                | `on-gold`        | `#0F2A23` | Text on gold                             |
|                | `gold-soft`      | `#F5EBCB` | Gold badge ground                        |
|                | `gold-ink`       | `#6B5210` | Gold badge text, "Diedit" tag            |
|                | `notice`         | `#FBF6E6` | Notice callout ground                    |
|                | `notice-ink`     | `#5C4A12` | Notice callout text, Tutup buku button   |
| Money & status | `income`         | `#1C6B3A` | Pemasukan; success; Online               |
|                | `income-soft`    | `#DCEFE2` | Success ground                           |
|                | `expense`        | `#9A3B1E` | Pengeluaran; danger; Offline             |
|                | `expense-soft`   | `#F6E1D8` | Danger ground                            |
|                | `expense-line`   | `#E8C6B8` | Border of destructive buttons            |
|                | `expense-ink`    | `#6E2A15` | Danger text, error toast                 |

Contrast: `ink` on `canvas`/`card` and `on-brand` on `brand` are well above WCAG AA. `ink-muted`
on `card` is AA for body text. Never put `gold` text on `card`; use `gold-ink`.

### Typography

Plus Jakarta Sans (variable, self-hosted so the TV works offline). **Tabular numerals everywhere**
(set on `body`) so times and Rupiah don't jitter. The scale is project-specific: `text-sm` is
13px here, not Tailwind's 14px.

| Role           | Classes                                                     | Size / weight | Example                   |
| -------------- | ----------------------------------------------------------- | ------------- | ------------------------- |
| Amount         | `text-4xl font-extrabold`                                   | 36 / 800      | Rp 750.000 (entry)        |
| Display        | `text-3xl font-extrabold`                                   | 30 / 800      | Saldo                     |
| Stat           | `text-2xl font-extrabold`                                   | 22 / 800      | Maghrib 18:15             |
| Page title     | `text-xl font-extrabold`                                    | 20 / 800      | Keuangan                  |
| Primary action | `text-lg font-bold`                                         | 16 / 700      | Simpan dan kirim ke layar |
| Section title  | `text-md font-bold`                                         | 15 / 700      | Aktivitas terakhir        |
| Body strong    | `text-base font-semibold`                                   | 14 / 600      | List titles               |
| Body           | `text-base`                                                 | 14 / 400      | Default                   |
| Small          | `text-sm font-semibold`                                     | 13 / 600      | Small buttons, chips      |
| Label          | `text-xs font-semibold text-ink-soft`                       | 12 / 600      | Field labels              |
| Meta           | `text-xs text-ink-muted`                                    | 12 / 400      | 28 Okt · Infaq transfer   |
| Caption caps   | `text-2xs font-bold uppercase tracking-wide text-ink-muted` | 11 / 700      | Table headers             |

**TV display** (E8): sizes are px of the 1920×1080 stage, as `text-tv-<px>` tokens (clock
`text-tv-196`, prayer times `text-tv-54`, labels `text-tv-26`); cards `rounded-tv` (20), slide
`rounded-tv-lg` (24); extra colors `tv-tile-label`, `tv-media`, `tv-caption`, `tv-dot`, `tv-dark`,
`tv-dark-ink`; `font-arabic` (Amiri) for the adzan screen.

### Spacing

4px grid (Tailwind steps). Choose by relationship, not by eye.

| Step  | px  | Use                         |
| ----- | --- | --------------------------- |
| `1`   | 4   | Icon to text in tight spots |
| `1.5` | 6   | Label to control            |
| `2`   | 8   | Chips, menu tiles           |
| `2.5` | 10  | Buttons side by side        |
| `3`   | 12  | Fields inside a card        |
| `3.5` | 14  | Cards on a page             |
| `4`   | 16  | Card padding                |
| `5`   | 20  | Page gutter, page sections  |
| `6`   | 24  | Hero padding                |

`Stack` encodes the common gaps: `xs` 6, `sm` 10, `md` 14, `lg` 20.

### Radius

| Class             | px  | Use                                |
| ----------------- | --- | ---------------------------------- |
| `rounded-sm`      | 8   | Tags, small stepper buttons        |
| `rounded-md`      | 10  | Inputs, small and medium buttons   |
| `rounded-lg`      | 12  | Large buttons, inner stat cards    |
| `rounded-xl`      | 14  | Cards, primary action, callouts    |
| `rounded-2xl`     | 16  | Cards on brand                     |
| `rounded-3xl`     | 18  | Saldo card                         |
| `rounded-t-sheet` | 24  | Bottom sheet top corners           |
| `rounded-full`    |     | Pills, avatars, round icon buttons |

### Elevation

Cards are **flat**. Only layers above the page cast a shadow.

| Token            | Use                 |
| ---------------- | ------------------- |
| (none)           | Cards, page content |
| `shadow-overlay` | Bottom sheet        |
| `shadow-toast`   | Toast               |

Stacking: page content → `ActionBar` (`z-10`) → kit section nav (`z-20`) → toasts (`z-50`).
Banners are in the page flow, not floating.
Sheets use the native `<dialog>` top layer, above everything.

### Iconography

[lucide-react](https://lucide.dev). 20px in buttons and headers, 22px with stroke 1.8 in menu
tiles. Icons are `aria-hidden`; the text or the button's `label` carries the meaning. Common set:
ChevronLeft (back), X (close), Search, Settings, EllipsisVertical (row menu), Plus/Minus, Check,
Copy, Upload, Image, AlignLeft (running text), CreditCard (keuangan), Clock (waktu), Bell (iqamah),
House (profil), Users (pengguna), Monitor (layar TV).

### Motion

| Duration       | Use                         |
| -------------- | --------------------------- |
| 150ms          | Color and opacity (default) |
| 160ms ease-out | Toast enters                |
| 220ms ease-out | Sheet slides up             |

All motion is removed under `prefers-reduced-motion`.

### Layout

- One column drawn at **390px**, capped at `max-w-app` (480px) and centered on wider screens.
- Page gutter 20px. Safe areas (`pt-safe-<n>`, `pb-safe-<n>`: spacing n or the notch inset,
  whichever is larger; `px-safe`) are built into `PageHeader`,
  `HeroHeader`, `ActionBar` and `Sheet`.
- Touch targets 44px (`size-touch`, `min-h-touch`). 36–40px buttons only inline in rows and cards.

## Components

Grouped as in `web/src/components/` and `/kit`. Every component takes `className`.

### Actions

| Component        | Use                                               | API                                                                                                                                                                  |
| ---------------- | ------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Button`         | Any press that does something.                    | `variant`: primary · secondary · outline · ghost · danger · income · expense · gold · notice. `size`: sm 36 · md 40 · lg 48 · xl 52. `block`, `loading`, `disabled`. |
| `buttonStyles()` | A router `<Link>` that should look like a button. | Same options as `Button`.                                                                                                                                            |
| `IconButton`     | Icon-only press.                                  | `label` (required, becomes `aria-label`), `icon`, `variant`: card (on cream) · brand (on green) · ghost (⋯ in rows).                                                 |
| `CopyField`      | Read-only value with a Salin button (the TV URL). | `value`, `tone`: default · brand.                                                                                                                                    |

**Choosing a button:** one `primary` per screen (usually `xl` in the `ActionBar`); `secondary`
for everything else; `danger` for destructive actions, always confirmed; `income`/`expense` only
for the two Keuangan entry buttons; `gold` for "act on this new thing" (Beri nama, Salin); `notice`
only inside a notice `Callout`.

### Forms

| Component          | Use                                           | API                                                                                                       |
| ------------------ | --------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `Field`            | Wraps one control with label, hint and error. | `label`, `hint`, `error`. Wires `id`, `aria-describedby`, `aria-invalid`.                                 |
| `Input`            | Single-line text, date, number.               | `size`: md 44 · lg 48 (login only).                                                                       |
| `PasswordInput`    | Every password field.                         | Masked by default, eye button to show/hide. Same props as `Input` (no `type`). |
| `Textarea`         | Multi-line text.                              | `rows`, `maxLength` + `showCount` ("57 / 200").                                                           |
| `Select`           | Choose one value from a list.                 | `options`, `value`, `onChange`, `size`: md · lg · pill. Styled listbox (not the native menu): opens under the field at full width in the top layer, flips up near the bottom, keyboard and screen-reader ready. With react-hook-form, wrap it in `Controller`. |
| `SearchInput`      | Filter a list in place.                       | `label` (required).                                                                                       |
| `MoneyInput`       | Rupiah amount entry.                          | `value: number \| null`, `onChange`, `flow`: in · out, `max`. Digits only, grouped, selects all on focus. |
| `SegmentedControl` | One value from 2–5 options.                   | `variant`: track (binary type) · split (scale, Hijri −2..+2) · pill (list filters). `tone` per option.    |
| `Stepper`          | Small bounded integer.                        | `min`, `max`, `step`, `format`, `size`: sm (tables) · md. Disables at bounds.                             |
| `Checkbox`         | On/off setting.                               | `label`, `size`. **The mockup uses checkboxes, not switches.**                                            |
| `RadioCard`        | One of a few choices that need a description. | `title`, `description`; group in a `<fieldset>` with a `<legend>`.                                        |
| `Slider`           | Continuous value (volume).                    | Native range; `minLabel`, `maxLabel`.                                                                     |
| `UploadTile`       | Pick files.                                   | `variant`: dropzone · tile (72px); `onFiles`, `accept`, `multiple`, `hint`.                               |

**Validation:** validate on submit, then on change for fields already in error. Error text says
what to do ("Format email atau nomor HP tidak dikenali"), in rust, under the control.

### Data display

| Component                                      | Use                                                    | API                                                                                                               |
| ---------------------------------------------- | ------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------- |
| `Card`                                         | Group related content.                                 | `tone`: default · brand (one hero figure per screen) · brand-2 (on green). `padding`: none · list · md · lg · xl. |
| `List`, `ListRow`                              | Records with a title, meta line and trailing value.    | `leading`, `title`, `description`, `trailing`. Wrap in `Card padding="list"`.                                     |
| `Table` (+ `THead`, `TBody`, `Tr`, `Th`, `Td`) | Real tabular data with column headers (7-day preview). | `density`: default · compact; `caption`. `Th scope="row"` for row labels. Scrolls sideways inside its card.       |
| `Badge`                                        | Short status or role.                                  | `tone`: success · danger · gold · notice · neutral; `size`: sm (tag) · md (status); `caps` (BARU, PENTING).                |
| `Avatar`                                       | A person, as initials.                                 | `name`, `tone` (follows role), `size`: md 40 · lg 44.                                                             |
| `Stat`                                         | A labelled figure.                                     | `label`, `value`, `size`: md · lg · xl, `onBrand`, `align`.                                                       |
| `Money`                                        | Any Rupiah amount.                                     | `amount`, `flow` (signed + colored), `withSymbol`.                                                                |
| `InfoTile`                                     | Read-only label/value, two per row.                    | `label`, `value`.                                                                                                 |
| `StatusDot`                                    | Compact status next to text.                           | `tone`: success · danger · warning · neutral; `label` for screen readers.                                         |
| `Thumbnail`                                    | Image preview.                                         | `src`, `alt`, `shape`: slide (72×40) · square (72), `muted`.                                                      |
| `Ticker`                                       | The TV's INFO bar of running texts.                    | `texts`, `speed`: slow 5 · medium 8 · fast 12 characters/s, `size`: tv · preview. One unbroken stream: texts joined by •, ۞ marks the return to the first.                |
| `SectionHeader`                                | Title over a group, with an optional link.             | `title`, `description`, `action`.                                                                                 |
| `Divider`                                      | Separate content; "atau" between alternatives.         | `label`.                                                                                                          |

**Drag to reorder.** `@dnd-kit/sortable` (Running text sets the pattern; Slide follows). Drag
only by a grip handle (`GripVertical`, `touch-none`) so the rest of the row still scrolls and
taps; keyboard users lift with space, move with arrows. The new order saves on drop.

**List or Table?** One record per row with a title (transactions, users, slides) → `List`.
Values compared across columns (times per day) → `Table`.

### Feedback

Four tones shared by Toast and Banner, each with its own icon: **info** (Info), **success**
(CircleCheck), **warning** (TriangleAlert), **danger** (CircleAlert).

| Component     | Use                                                                                        | API                                                                                                                                                  |
| ------------- | ------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `useToast`    | Passing feedback after an action.                                                          | `toast({ message, description?, tone?, action?, duration? })` returns an id; `toast.dismiss(id)`. Needs `ToastProvider` (already in `AppProviders`). |
| `Banner`      | Screen-wide status in a full-width strip.                                                  | `tone`, `title`, `action`, `onDismiss` (only for plain information), `live` (announce when it appears later).                                        |
| `Callout`     | Inline message inside the content: consequences before acting, or a problem with one item. | `tone`: notice · danger; `title`, `action`.                                                                                                          |
| `EmptyState`  | A list or screen with no data yet.                                                         | `title`, `description`, `action`.                                                                                                                    |
| `Skeleton`    | Loading placeholder when the layout is known.                                              | Size it with `className`.                                                                                                                            |
| `Spinner`     | Short inline wait.                                                                         | `label`.                                                                                                                                             |
| `ProgressBar` | Known progress (iqamah, program targets).                                                  | `value`, `max`, `label`, `tone`: gold · brand · income.                                                                                              |

**Toast vs Banner vs Callout**

| Use       | When                                 | Lifetime                  | Example                                                |
| --------- | ------------------------------------ | ------------------------- | ------------------------------------------------------ |
| `Toast`   | One action just finished (or failed) | Disappears by itself      | "Tersimpan dan dikirim ke layar"                       |
| `Banner`  | A condition affects the whole screen | While the condition lasts | "Anda sedang offline", "Mode lihat saja", month locked |
| `Callout` | Something about one card or form     | Part of the content       | Saldo preview, Tutup buku card, one TV offline         |

**Toast rules**

- Sits above the `ActionBar`, newest at the bottom, **at most 3** on screen.
- Durations: success and info 4s, warning 6s, danger 8s. Hover or focus pauses; `duration: null`
  keeps it until closed (use for anything that needs a decision).
- Toasts float, so they use dark surfaces: success `brand`, info `on-gold`, warning
  `notice-ink`, danger `expense-ink`.
- One action at most, usually **"Batalkan"** for a reversible change. Never put the only way to
  do something in a toast.
- Danger toasts are announced immediately (`role="alert"`); the rest politely.
- Messages are a short past-tense result ("Slide dinonaktifkan"), with detail in `description`.

**Banner rules**

- Placed directly under the `PageHeader`, outside `PageBody`, so it spans the full width.
- Soft surfaces: info `track`, success `income-soft`, warning `notice`, danger `expense-soft`.
- One banner per screen; if two conditions apply, show the more severe.
- Not dismissible while the condition is real (offline, locked month). Dismissible only for
  plain information.
- Pass `live` when it can appear after the screen loaded (going offline), so screen readers hear it.

### Overlays

| Component       | Use                                              | API                                                                                          |
| --------------- | ------------------------------------------------ | -------------------------------------------------------------------------------------------- |
| `Sheet`         | Short form or detail without leaving the screen. | `open`, `onClose`, `title`, `footer`. Native `<dialog>`: focus trap, Escape, backdrop close. |
| `ActionMenu`    | The ⋯ at the end of a row.                       | `label` (row name), `actions: { label, onSelect, danger?, disabled? }[]`. Destructive last.  |
| `ConfirmDialog` | Confirm before anything that locks or deletes.   | `title`, `description`, `confirmLabel`, `danger`, `requireReason` → `onConfirm(reason)`.     |
| `LeaveGuard`    | A screen with a draft that is saved by a button. | `when` (unsaved changes). Asks before back or a link; the browser asks on tab close.         |

No centered modals and no tooltips: on a phone, sheets keep actions in thumb reach and tooltips
can't be hovered.

### Navigation

| Component    | Use                                                | API                                                                      |
| ------------ | -------------------------------------------------- | ------------------------------------------------------------------------ |
| `PageHeader` | Top of every screen except Beranda.                | `title`, `meta`, `actions`, `backTo`, `dismiss`: back · close.           |
| `MenuTile`   | Beranda "Kelola" grid, 4 per row.                  | `to`, `icon`, `label`. Render only tiles the role may open.              |
| `NavCard`    | A card that opens a screen, with a status summary. | `to`, `title`, `description`, `trailing`.                                |
| `Tabs`       | Switch between views of content (Kategori in/out). | `label`, `items: { value, label, content }[]`. Arrow keys, one tab stop. |
| `TextLink`   | Inline link ("Lihat semua", "Ubah").               | Router `Link` props, `size`.                                             |

**Tabs or SegmentedControl?** Changing what is shown → `Tabs`. Choosing a value that gets saved
→ `SegmentedControl`.

### Layout

| Component    | Use                                                          |
| ------------ | ------------------------------------------------------------ |
| `HeroHeader` | Green header that opens Beranda.                             |
| `PageBody`   | The content column under a header (20px gutters, 14px gaps). |
| `Stack`      | Vertical spacing from the scale: `gap` xs · sm · md · lg.    |
| `ActionBar`  | Sticky bottom bar for the screen's primary action.           |
| `MosqueLogo` | Placeholder mark until a mosque uploads its logo.            |

## Patterns

**Screen anatomy.** `PageHeader` → `PageBody` of cards → `ActionBar` with one `xl` primary
button. Beranda replaces the header with `HeroHeader` and has no action bar.

**Saving.** Settings screens save as a whole with "Simpan dan kirim ke layar", then
`toast({ message: 'Tersimpan dan dikirim ke layar' })`. The button shows `loading` while saving.
If some TVs are offline, save anyway and use a `warning` toast naming them. On failure, keep the
input and show a `danger` toast; never clear a form on error.

**Undo over confirm.** For reversible, low-stakes changes (deactivate a slide, reorder), act
immediately and offer "Batalkan" in the toast. Keep `ConfirmDialog` for what can't be undone or
needs a reason.

**What reaches the TV.** A save that goes straight to the screens (Waktu sholat settings, a
jadwal khusus) asks with `ConfirmDialog` first, and so does removing an item, even from a draft.
Draft edits (steppers, segmented controls) don't ask; the save does.

**Color roles on settings screens** (Waktu sholat sets the pattern; Iqamah and Layar follow).
Each color has one job, so the hierarchy reads without labels:

| Color                              | Means                               | e.g.                                                       |
| ---------------------------------- | ----------------------------------- | ---------------------------------------------------------- |
| `brand` (`Card tone="brand"`)      | Context from elsewhere, or "now"    | Method + location card; today's row in the 7-day preview   |
| `card` (white)                     | Settings edited on this screen      | Koreksi waktu, Jadwal mingguan, Koreksi Hijriah            |
| `gold` / `gold-ink`                | Differs from computed or from saved | Times replaced by an override; a changed stepper's value   |
| `highlight`                        | Row being edited                    | A correction row changed since the last save               |
| `notice` (`Badge tone="notice"`)   | Acts right away, take care          | "Langsung tersimpan"; "Ada perubahan belum disimpan"       |
| `expense`                          | Removes something                   | × on a rule or override                                    |

One big brand block per screen; elsewhere brand stays small (a row, a button). Gold never
decorates: it always means "not the computed or saved value". Income and expense greens and reds
stay with money.

**Unsaved changes.** A screen with a draft shows "Ada perubahan belum disimpan" above its save
button and mounts `LeaveGuard` while the draft differs from what is saved.

**Destructive and locking actions.** Always `ConfirmDialog`. Transaction edits and deletes use
`requireReason`. Tutup buku explains the lock in the description.

**States** (E10-01). Every list has: loading (`Skeleton` rows), empty (`EmptyState` with the
action that fills it), error (`Callout tone="danger"` with what to do). Offline TV state uses a
danger `Callout` that says the TV keeps running from cache.

**Role-based UI.** Hide what the role can't use ([permissions](../product/prd.md#roles-and-permissions)).
View-only roles see values as text, not disabled inputs, with an info `Banner` ("Mode lihat saja").

**Offline (admin).** When the phone loses connection, show a `live` warning `Banner` ("Anda sedang
offline") and disable saving; keep what was typed.

**Lists with row actions.** Title + meta on the left, value or `Badge` on the right, `ActionMenu`
last. Tapping a row opens its edit `Sheet`.

## Content

- **Bahasa Indonesia**, short and polite. Use the mockup's wording where it exists.
- **Verbs on buttons**: "Simpan", "Tambah", "Tutup buku", not "OK".
- **Money**: `formatRupiah` → "Rp 92.058.114"; in lists, signed without symbol: "+ 750.000".
- **Dates and times**: "Rabu, 28 Okt 2026", "28 Okt", "18:15" (colon, 24h; `id-ID` would
  give "18.15"). Zones are written WIB / WITA / WIT. Add formatters to `utils/format/` with the
  first screen that shows a date.
- **Separators**: " · " between meta parts.
- **Donor default**: "Hamba Allah".
- **Placeholders**: bracketed text in mockups (`[NAMA MASJID]`) is sample data, not copy.
- **Empty optional fields are hidden**, never shown blank.

## Accessibility

- Icon-only buttons need a `label`; icons are `aria-hidden`.
- Every control is inside a `Field` or has its own `aria-label`.
- Groups (`SegmentedControl`, `Stepper`) are `<fieldset>`s with an `aria-label`; `Tabs` use
  tablist semantics with arrow keys.
- Status is never color alone: `StatusDot` and `Badge` pair color with text.
- Focus ring: 3px gold outline, visible on cream and green.
- Toasts and counters are polite live regions; danger toasts use `role="alert"`. Banners that
  appear after load use `live`.
- Motion respects `prefers-reduced-motion`.

## Changing the system

1. Need a new color, size or radius? Add a token in `styles/index.css` and a row here first.
2. New component? Put it in the right folder, export it from `components/index.ts`, add a
   specimen to `/kit` (`src/pages/Kit/`), a row here, and a test if it has behavior.
3. Keep this file, `/kit` and the code in the same order; reviewers check all three.

### Not in the system yet

- **Map pin** for Profil masjid (E3-03): Leaflet, per the backlog.
- **QR code** for Layar TV (E9-06).
- **Dark mode**: not in the mockup; the TV has its own dark ground.
