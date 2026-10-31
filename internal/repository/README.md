# repository

Data access and external integrations. Implements the repository interfaces declared in `internal/usecase/interface.go`.

## What goes here

One subfolder per data source / client, for example:

- `postgres/`: database storage
- `memory/`: in-memory store (cache, local dev, tests)
- `<service>/`: third-party API client (e.g. a prayer time API)

## Rules

- Each subpackage exposes a `NewRepository(...)` constructor that returns a type satisfying a `usecase.*RepositoryItf`
- Accept and return `domain` types; convert DB rows / API payloads inside this layer
- No business logic: that belongs in `internal/usecase`
- Client config lives in `internal/config/schema_client.go` and is passed in from `cmd/http/main.go`
