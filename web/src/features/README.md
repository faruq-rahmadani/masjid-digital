# features

One folder per domain area, mirroring `internal/domain` on the Go side. Add subfolders only when
they have files. Planned features and their epics ([backlog](../../../docs/product/backlog.md)):

| Folder           | Epic   | Screens                                     |
| ---------------- | ------ | ------------------------------------------- |
| `auth`           | E2     | Login                                       |
| `users`          | E2     | Pengguna                                    |
| `mosque-profile` | E3     | Profil masjid                               |
| `prayer-times`   | E4, E5 | Waktu sholat, Iqamah & alarm, Petugas Jumat |
| `content`        | E6     | Slide, Running text                         |
| `finance`        | E7     | Keuangan, Tambah transaksi, Kategori        |
| `screens`        | E9     | Layar TV                                    |
| `tv`             | E8     | The `/tv/{slug}` display                    |

```
features/<name>/
  api/          # fetchers + query keys/options, built on `@/lib/api-client`
  components/   # UI used only by this feature
  hooks/        # useQuery / useMutation wrappers
  routes/       # page components mounted in `src/app/routes.tsx`
  types.ts      # shapes of the backend payloads
  index.ts      # public API: the only file other modules may import
```

## Rules

- Import another feature only through its `index.ts`, never a deep path.
- Features may import from `components`, `lib`, `config`; those never import from features.
- Server state lives in TanStack Query. Reach for local `useState` before any global store.
- Query keys are defined once per feature, e.g.:

  ```ts
  export const announcementKeys = {
    all: ['announcements'] as const,
    detail: (id: string) => [...announcementKeys.all, id] as const,
  }
  ```
