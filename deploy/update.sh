#!/usr/bin/env bash
# Run by cron on the server, from a clone of this repo: when origin/main moved, fast-forwards
# the clone, rebuilds the image and restarts the site. deploy/.deployed-commit only moves after
# a healthy deploy, so a failed run is retried next time. See deploy/README.md.
#
#   DOCKER="sudo -n docker" PROXY_CONTAINER=masjid-proxy-caddy-1 deploy/update.sh
set -euo pipefail

# The braces make bash read the whole script before running it: the merge below may
# replace this very file.
{
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE="$REPO/deploy/.deployed-commit"
LOG="$REPO/deploy/update.log"
read -r -a DOCKER <<<"${DOCKER:-docker}"
# the reverse proxy's container, used to check the site answers on the shared network
PROXY_CONTAINER="${PROXY_CONTAINER:-masjid-proxy-caddy-1}"

exec >>"$LOG" 2>&1
log() { echo "$(date -u '+%F %T UTC') $*"; }
trap 'log "FAILED at line $LINENO"' ERR

exec 9>"$REPO/deploy/.update.lock"
flock -n 9 || { log "another run is active, skipping"; exit 0; }

# keep the log small
tail -n 1000 "$LOG" >"$LOG.tmp" && cat "$LOG.tmp" >"$LOG" && rm -f "$LOG.tmp"

git -C "$REPO" fetch --quiet origin main
target="$(git -C "$REPO" rev-parse origin/main)"
deployed="$(cat "$STATE" 2>/dev/null || true)"
[[ "$target" != "$deployed" ]] || { log "up to date (${target:0:7})"; exit 0; }

git -C "$REPO" merge --ff-only --quiet origin/main
log "building ${target:0:7} (deployed: ${deployed:0:7})"

cd "$REPO"
previous="$("${DOCKER[@]}" image inspect -f '{{.Id}}' masjid-public:latest 2>/dev/null || true)"
"${DOCKER[@]}" compose build --quiet web
"${DOCKER[@]}" compose up -d web

healthy=
for _ in $(seq 15); do
	if "${DOCKER[@]}" exec "$PROXY_CONTAINER" wget -q -O /dev/null http://public-web/; then healthy=1; break; fi
	sleep 2
done
[[ -n "$healthy" ]] || { log "site did not answer after deploy"; exit 1; }

echo "$target" >"$STATE"
current="$("${DOCKER[@]}" image inspect -f '{{.Id}}' masjid-public:latest)"
if [[ -n "$previous" && "$previous" != "$current" ]]; then
	"${DOCKER[@]}" image rm "$previous" >/dev/null 2>&1 || true
fi
log "deployed ${target:0:7}"
exit 0
}
