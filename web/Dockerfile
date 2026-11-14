# Admin panel and TV display image: the built SPA served as static files, with
# /api/* forwarded to the backend at $API_UPSTREAM (default api:3000), so the app
# is one origin on port 80 and one image fits every domain.
#
#   docker build -t masjid-digital-web web/

FROM node:24-alpine AS build
WORKDIR /src

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
# Shown on the Layar TV page, e.g. a release tag or the commit
ARG VITE_APP_VERSION=dev
ENV VITE_APP_VERSION=$VITE_APP_VERSION VITE_API_BASE_URL=/api
RUN npm run build

FROM caddy:2-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=build /src/dist /srv
EXPOSE 80
