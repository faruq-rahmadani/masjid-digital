# The API as one container, for hosts other than the VPS setup in deploy/ (see docs/deploy.md).
# The web clients (web/, web_public/) are static and deploy separately, e.g. to Vercel.
#
#   docker build -t masjid-digital-api .
#   docker run -p 3000:3000 -e APP_AUTH_JWT_SECRET=... -e APP_DATABASE_DRIVER=postgres \
#     -e APP_DATABASE_DSN=postgres://... masjid-digital-api
#
# Configuration comes from environment variables (see config.yaml.example), or from a file
# mounted at /etc/masjid-digital-config/config.yaml.

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY migrations ./migrations
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/http ./cmd/http \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/admin ./cmd/admin

FROM alpine:3.22
# tzdata: the mosque's time zone; ca-certificates: TLS to the database
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=build /out/http /out/migrate /out/admin ./
USER app
ENV ENV=production
EXPOSE 3000
CMD ["/app/http"]
