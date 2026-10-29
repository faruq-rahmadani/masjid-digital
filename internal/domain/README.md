# domain

Core business entities of masjid-digital and the rules that belong to them.

## What goes here

- Entity structs (e.g. `Masjid`, `PrayerSchedule`, `Announcement`)
- Request/response payload structs shared across layers
- Domain errors (e.g. `ErrMasjidNotFound`)
- Pure validation / helper methods on entities

## Rules

- No dependency on other `internal/` packages; every other layer depends on `domain`, never the other way around
- No framework, database, or HTTP client imports; keep it plain Go
- One file per entity, with its test next to it (`masjid.go`, `masjid_test.go`)
