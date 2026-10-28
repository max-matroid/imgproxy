#!/usr/bin/env bash
# Starts imgproxy in Docker, runs the latency benchmark against it, and
# stops the container afterwards. Extra arguments are passed to the benchmark.
#
#   ./run.sh                      # defaults
#   ./run.sh -n 500 -c 4 -json results.json
#   IMGPROXY_IMAGE=ghcr.io/imgproxy/imgproxy:v3.28 ./run.sh
set -euo pipefail

cd "$(dirname "$0")"

IMAGE="${IMGPROXY_IMAGE:-ghcr.io/imgproxy/imgproxy:latest}"
PORT="${IMGPROXY_PORT:-8080}"
NAME="imgproxy-latency-bench"

docker rm -f "$NAME" >/dev/null 2>&1 || true
docker run -d --name "$NAME" \
  -p "127.0.0.1:${PORT}:8080" \
  --add-host=host.docker.internal:host-gateway \
  -e IMGPROXY_ALLOW_PRIVATE_SOURCE_ADDRESSES=true \
  -e IMGPROXY_ALLOW_LOOPBACK_SOURCE_ADDRESSES=true \
  -e IMGPROXY_LOG_LEVEL=warn \
  ${IMGPROXY_DOCKER_ARGS:-} \
  "$IMAGE" >/dev/null
trap 'docker rm -f "$NAME" >/dev/null 2>&1 || true' EXIT

for _ in $(seq 1 30); do
  curl -fs "http://127.0.0.1:${PORT}/health" >/dev/null && break
  sleep 1
done

go run . -imgproxy "http://127.0.0.1:${PORT}" "$@"
