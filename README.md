# masjid-digital-public

The public site for Masjid Digital: what the product does, its documentation, the live demo, the
plans, the shared server's numbers and the "Daftar" form. The app itself (admin panel and TV
display) lives in the masjid-digital repository. The site has its own small Go API in `server/`
for the form and the shared server's numbers.

- `/` (`index.html`): the product page. The hero's TV switches between the real screenshots of
  each mode; the admin panel is shown as its phone screens.
- `/docs/` (`docs/index.html`): the user guide, from installing a TV to every admin menu.

Plain HTML pages, Tailwind CSS v4 and a small script (`src/main.ts`), built with Vite. No React:
the content is in the HTML, so search engines read it without running JavaScript.

## Commands

```sh
cp .env.example .env   # demo links point at the local admin app, http://localhost:5173
npm install
npm run dev            # http://localhost:5174
npm run check          # typecheck + Biome + tests
npm run build          # static files in dist/
```

## Demo links

Every "Coba demo" and "Lihat layar TV demo" button reads its URL from `.env`:

| Variable           | What it opens        | When empty                  |
| ------------------ | -------------------- | --------------------------- |
| `VITE_DEMO_URL`    | the demo admin panel | the demo buttons are hidden |
| `VITE_DEMO_TV_URL` | the demo TV display  | `VITE_DEMO_URL` + `/tv`     |

Locally, run the admin app too (`cd web && npm run dev`, port 5173) so the demo works; both ports
are fixed (`strictPort`), so neither shifts onto the other. The values are baked in at build
time, so rebuild after changing them. Without JavaScript the
buttons open the documentation instead.

The "Coba demo" section runs the demo inside the page, side by side: the admin panel in a 390 px
phone frame (on a computer; on a phone a button opens it instead) and the TV in a 16:9 frame, so a
change saved on one shows on the other. Nothing loads until the visitor clicks "Mulai demo". For
this to work the demo host must allow being framed by this site: send no `X-Frame-Options`, or a
`Content-Security-Policy: frame-ancestors` that lists the public site's domain.

`VITE_DEMO_PHONE` and `VITE_DEMO_PASSWORD` show the demo's login, with copy buttons, once the demo
starts: the panel runs on the demo's own domain, so the page cannot fill its form. Every visitor
sees them, so use an account made for the demo.

Every visitor who opens the TV demo registers as a new screen on the demo server (BARU, at most
50). Reset the demo's data now and then, e.g. nightly.

## The API (`server/`)

A Go service with two public routes, which the page calls on its own origin:

| Route                             | What it does                                                                                                                                                                         |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `POST /api/public/v1/inquiries`   | the Daftar form (`#daftar`): validated, rate limited (5 per visitor then one per 10 minutes, 30 in total then one per 2 minutes) and sent to the team's Telegram. Nothing is stored. |
| `GET /api/public/v1/transparency` | the "server bersama" numbers (`#transparansi`): credit and running cost of the shared server's IDCloudHost billing account. Cached for an hour.                                      |

Each answers 503 until its settings are filled in, and the page then shows a short note instead.
Layout follows clean architecture: `domain` (rules), `usecase` (application, against the
interfaces in `usecase/interface.go`), `repository` (Telegram, IDCloudHost), `delivery/http`
(router, controllers, rate limit), and `cmd/http` (`main.go` runs the server, `wire.go` builds the
dependencies).

```sh
cd server
cp .env.example .env   # development settings; read from the working directory
go run ./cmd/http      # http://localhost:8080, the Vite dev server forwards /api here
go test ./...
```

In production `docker compose` runs it as the `api` container with the repo root's `.env` as its
environment, and the site's Caddy forwards `/api/*` to it (see [deploy/README.md](deploy/README.md)).
`VITE_INQUIRY_URL` and `VITE_TRANSPARENCY_URL` point the page elsewhere, which then needs CORS.

### The shared server's numbers

- **Saldo deposit**: `running_totals.ongoing` of the billing account
  (`GET /v1/payment/billing_account`), the credit left after this month's charges so far
  (`credit_available` is the credit before them), plus `TRANSPARENCY_EXTRA_DEPOSIT`, money held for
  the server but not topped up yet.
- **Biaya server / bulan**: the hourly `price` of each resource billed this month
  (`GET /v1/charging/usage`: the VM, its disk, each backup replica) for the VMs in
  `TRANSPARENCY_VM_NAMES` (all when empty), times the hours in the month, VAT added. IDCloudHost's
  own "Est. monthly cost" is the same figure before VAT. One-off products (`price_unit` `pc`) do
  not count.
- **Aman sampai**: today plus deposit ÷ the monthly cost _before_ VAT, since IDCloudHost charges
  the VAT when credit is topped up, not when it is used.

IDCloudHost has no read-only API key: the key can also create or delete servers and pay invoices,
so it only lives in the server's `.env`. The credit belongs to the whole billing account: if the
account also pays for other projects, its credit counts toward the deposit too. A billing account
of its own for the shared server keeps the numbers exact.

## Screenshots

`public/screenshots/` holds the app's screenshots, served at `/screenshots/`. They are copies of
the ones in the masjid-digital README (`docs/screenshots/`): retake one there, copy it here.

## Deploy

A Docker image (Vite build served by Caddy), deployed from a clone of this repo by a cron job:
see [deploy/README.md](deploy/README.md).

## Copy

All visible text is Bahasa Indonesia, like the app. Describe only what the app does today; a
feature that is not built yet is labelled "segera hadir". Keep the docs in step with
`docs/product/prd.md` and the README in masjid-digital when a feature changes.
