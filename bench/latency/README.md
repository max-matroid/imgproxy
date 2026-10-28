# imgproxy latency vs direct fetch

Measures how much latency imgproxy adds compared to fetching the same image
straight from the origin.

The benchmark starts a local origin HTTP server (serving generated test images,
or your own via `-images`), then for each image runs these scenarios:

| scenario   | request                                                  |
|------------|----------------------------------------------------------|
| `direct`   | `GET origin/<img>`, the baseline                         |
| `raw`      | `/insecure/raw:1/plain/<src>`: proxy only, no processing |
| `reencode` | `/insecure/plain/<src>`: decode + re-encode, same format |
| `resize`   | `/insecure/rs:fit:300:300/plain/<src>`                   |
| `webp`     | `/insecure/rs:fit:300:300/plain/<src>@webp`              |
| `avif`     | `/insecure/rs:fit:300:300/plain/<src>@avif`              |

It reports TTFB and total time (including reading the body) at p50/p90/p99/max,
throughput, response size, and p50 overhead vs `direct`.

## Run

Requires Docker and Go.

```sh
./run.sh                                  # imgproxy:latest, 200 req, concurrency 1
./run.sh -n 500 -c 8 -json results.json   # more requests, under load
./run.sh -images ~/photos -scenarios direct,raw,webp
IMGPROXY_IMAGE=ghcr.io/imgproxy/imgproxy:v3.28 ./run.sh
IMGPROXY_DOCKER_ARGS="--cpus=2 -e IMGPROXY_WORKERS=2" ./run.sh
```

To benchmark an imgproxy you're already running, use
`go run . -imgproxy http://host:port -origin-imgproxy http://<origin as seen by imgproxy>:9000`.

## Caveats

- On macOS, Docker Desktop runs containers in a VM, so `raw` includes two extra
  VM network hops (client→container, container→origin) that `direct` doesn't.
  Treat `raw - direct` as an upper bound on proxy overhead; run on Linux for
  numbers close to production.
- The origin is local and serves from the page cache, so `direct` is a
  best case. Real origins (S3, CDN) add network latency that applies to both
  paths, which shrinks the *relative* overhead.
- `-c 1` measures pure latency; increase `-c` to see queueing once requests
  exceed `IMGPROXY_WORKERS`.
