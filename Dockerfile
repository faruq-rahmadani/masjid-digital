# API image: the Go backend with the migrate and create-admin tools. It connects
# to Postgres by default; set MASJID_APP_DATABASE_DSN (see deploy/config.yaml).
# The admin panel is a separate image, see web/Dockerfile.
#
#   docker build -t masjid-digital-api .

FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
COPY pkg ./pkg

# modernc.org/sqlite is pure Go, so the binaries need no cgo and run on distroless/static
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/masjid-digital ./cmd/http \
 && go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
 && go build -trimpath -ldflags="-s -w" -o /out/create-admin ./cmd/create-admin

# distroless/static ships tzdata and CA certificates, and runs as nonroot (65532)
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/masjid-digital /out/migrate /out/create-admin /usr/local/bin/
COPY deploy/config.yaml /etc/masjid-digital-config/config.yaml

ENV ENV=production
EXPOSE 3000

CMD ["masjid-digital"]
