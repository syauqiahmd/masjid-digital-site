# Deploy

The site runs as the compose project `masjid-public` (`../docker-compose.yml`): `web` on a
Docker network it shares with the reverse proxy that holds the domain and TLS, and `api`
(`server/`), which only `web` can reach. The proxy reaches it as
`public-web:80`, e.g. with Caddy:

```
example.com {
	encode zstd gzip
	reverse_proxy public-web:80
}
```

The `web` image is the Vite build served by Caddy ([`Caddyfile`](Caddyfile)), which forwards
`/api/*` to the `api` container. Its settings (Telegram bot and chat for the form, IDCloudHost key
and billing account for the shared server's numbers) come from `.env`. Every proxy in front must
append to `X-Forwarded-For` from a private network, so the form's rate limit sees the visitor's IP.

## First setup on the server

```sh
git clone <this repo> ~/site/masjid-digital-public
cd ~/site/masjid-digital-public
cp .env.example .env
#   VITE_DEMO_URL=https://<demo domain>   (empty hides the demo)
#   TELEGRAM_BOT_TOKEN=... TELEGRAM_CHAT_ID=...                      (the Daftar form)
#   IDCLOUDHOST_API_KEY=... IDCLOUDHOST_BILLING_ACCOUNT_ID=...       (the shared server's numbers)
#   TRANSPARENCY_VM_NAMES=masjidyuk, TRANSPARENCY_EXTRA_DEPOSIT       (see the README)
chmod 600 .env   # holds the IDCloudHost key, which can change the hosting account
docker compose up -d --build
git rev-parse HEAD > deploy/.deployed-commit
```

The demo links are baked in at build time, so change `.env` and rebuild to move them. The
API's settings are read when its container starts: after changing them, `docker compose up -d api`.

## Updates

`update.sh` checks `origin/main`; when it moved it fast-forwards the clone, rebuilds, restarts both
containers and checks the proxy can reach the site and the site can reach its API. It logs to `deploy/update.log` and remembers the
deployed commit in `deploy/.deployed-commit`, which only moves after a healthy deploy, so a
failed run is retried next time. Run it from cron, e.g. at 06:00 and 18:00 WIB on a server in UTC:

```
0 11,23 * * * DOCKER="sudo -n docker" /home/<user>/site/masjid-digital-public/deploy/update.sh
```

`DOCKER` is how the script calls Docker (default `docker`); `PROXY_CONTAINER` is the proxy's
container name for the health check (default `masjid-proxy-caddy-1`).

## Moving the API into this repo (once)

Before, the Daftar form went to a mosque's masjid-digital API (`API_UPSTREAM`). Now the site has its
own `api` container. `update.sh` deploys whatever lands on main, so fill in `.env` **before** the
change is pushed; otherwise the API starts without its settings and the form answers 503.

1. On the server, add to `.env` (keep `API_UPSTREAM` for now, the old compose file needs it if you
   roll back), then `chmod 600 .env`:

   ```
   TELEGRAM_BOT_TOKEN=...            # same bot and chat as the mosque's app.inquiry
   TELEGRAM_CHAT_ID=...
   INQUIRY_ENV_TAG=                  # empty in production
   IDCLOUDHOST_API_KEY=...
   IDCLOUDHOST_BILLING_ACCOUNT_ID=...
   TRANSPARENCY_VM_NAMES=masjidyuk
   TRANSPARENCY_EXTRA_DEPOSIT=0
   VITE_DEMO_URL=https://test.masjidyuk.com
   VITE_DEMO_PHONE=...               # a demo account: every visitor sees it
   VITE_DEMO_PASSWORD=...
   ```
2. Push to main. The next `update.sh` run builds both images (the first API build pulls the Go
   image, so it takes a few minutes) and only records the deploy once the API answers too. Or run
   it by hand: `DOCKER="sudo -n docker" deploy/update.sh; tail deploy/update.log`.
3. Check: `curl https://<site>/api/public/v1/transparency` shows the numbers, and one test
   submission of the form reaches Telegram.
4. Then the old route can go: drop `API_UPSTREAM` from `.env`, the Telegram settings from the
   mosque's stack in masjid-digital-deployment, and the inquiry code from masjid-digital.

To roll back, revert the change on main and push: `update.sh` deploys it on its next run (run it
by hand to go faster), and the form goes back to `API_UPSTREAM`. Then `docker compose rm -sf api`
removes the API container, which the old compose file no longer lists.
