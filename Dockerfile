# The public site: built by Vite, served by Caddy on port 80 (TLS is left to the proxy in front).
#
#   docker compose build   # reads VITE_DEMO_URL / VITE_DEMO_TV_URL from .env, see deploy/README.md

FROM node:24-alpine AS build
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY . ./
# baked into the pages at build time (see .env.example); empty hides the demo
ARG VITE_DEMO_URL
ARG VITE_DEMO_TV_URL
ARG VITE_DEMO_PHONE
ARG VITE_DEMO_PASSWORD
ARG VITE_INQUIRY_URL
ARG VITE_TRANSPARENCY_URL
ARG VITE_SITE_URL
ARG VITE_GOOGLE_SITE_VERIFICATION
RUN npm run build

FROM caddy:2-alpine
COPY deploy/Caddyfile /etc/caddy/Caddyfile
COPY --from=build /src/dist /srv
