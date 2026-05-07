#!/bin/sh
set -eu

echo "[run.sh] Starting service"

echo "[run.sh] Starting Caddy"
caddy run --config /etc/caddy/Caddyfile &

echo "[run.sh] Starting Go app"
PORT=8080 exec /app/bin/app
