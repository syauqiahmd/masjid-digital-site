# Deploy

The site runs as one container, `masjid-public` (`../docker-compose.yml`), on a Docker network
it shares with the reverse proxy that holds the domain and TLS. The proxy reaches it as
`public-web:80`, e.g. with Caddy:

```
example.com {
	encode zstd gzip
	reverse_proxy public-web:80
}
```

The image is the Vite build served by Caddy ([`Caddyfile`](Caddyfile)), which also forwards the
Daftar form, `POST /api/public/v1/inquiries`, to a masjid-digital backend (`API_UPSTREAM`). That
backend needs `app.inquiry` set (the Telegram bot and chat), and every proxy in between must
append to `X-Forwarded-For` from a private network, so its rate limit sees the visitor's IP.

## First setup on the server

```sh
git clone <this repo> ~/site/masjid-digital-public
cd ~/site/masjid-digital-public
cp .env.example .env
#   VITE_DEMO_URL=https://<demo domain>   (empty hides the demo)
#   API_UPSTREAM=<mosque>-web:80          (a mosque's web container on the proxy network)
docker compose up -d --build
git rev-parse HEAD > deploy/.deployed-commit
```

The demo links are baked in at build time, so change `.env` and rebuild to move them.

## Updates

`update.sh` checks `origin/main`; when it moved it fast-forwards the clone, rebuilds, restarts the
container and checks the proxy can reach it. It logs to `deploy/update.log` and remembers the
deployed commit in `deploy/.deployed-commit`, which only moves after a healthy deploy, so a
failed run is retried next time. Run it from cron, e.g. at 06:00 and 18:00 WIB on a server in UTC:

```
0 11,23 * * * DOCKER="sudo -n docker" /home/<user>/site/masjid-digital-public/deploy/update.sh
```

`DOCKER` is how the script calls Docker (default `docker`); `PROXY_CONTAINER` is the proxy's
container name for the health check (default `masjid-proxy-caddy-1`).
