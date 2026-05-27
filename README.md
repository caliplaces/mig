# Muscle Group Image Generator

[![CI](https://github.com/caliplaces/mig/actions/workflows/ci.yml/badge.svg)](https://github.com/caliplaces/mig/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev)

HTTP API that renders an anatomical body silhouette with one or more muscle
groups highlighted in arbitrary colors. Drop-in compatible with the original
PHP API surface; reimplemented in Go for ~20× faster rendering, a ~10 MB
static binary, and zero runtime dependencies.

## Quick start

```bash
make run            # listens on :8080
curl http://localhost:8080/getMuscleGroups
curl -o out.png 'http://localhost:8080/getImage?muscleGroups=chest,triceps&color=FF0000'
```

Or with Docker:

```bash
make docker
docker run --rm -p 8080:8080 mig:dev
```

## API

All endpoints are `GET`. The legacy camelCase paths are preserved for
backward compatibility with existing clients.

| Path | Description |
| --- | --- |
| `/getMuscleGroups` | JSON list of supported muscle group names |
| `/getBaseImage` | The blank silhouette PNG |
| `/getImage` | Highlight one or more groups in a single color |
| `/getMulticolorImage` | Highlight two sets of groups in two different colors |
| `/getIndividualColorImage` | Assign a distinct color to each group |
| `/healthz` | Liveness probe (returns `ok`) |

### Query parameters

| Parameter | Where | Format | Notes |
| --- | --- | --- | --- |
| `muscleGroups` | `/getImage`, `/getIndividualColorImage` | comma-separated names | required |
| `primaryMuscleGroups`, `secondaryMuscleGroups` | `/getMulticolorImage` | comma-separated names | required |
| `color`, `primaryColor`, `secondaryColor` | image endpoints | hex (`#FF0000` / `FF0000` / `F00`) or RGB (`255,0,0`) | required where listed |
| `colors` | `/getIndividualColorImage` | comma-separated hex or RGB | count must equal `muscleGroups` count |
| `transparentBackground` | image endpoints | `0` or `1` (default `0`) | optional |

Errors are returned as `{"error": "..."}` JSON with an appropriate 4xx status.

### Examples

```
/getImage?muscleGroups=chest,triceps&color=FF0000&transparentBackground=1
/getMulticolorImage?primaryMuscleGroups=chest&secondaryMuscleGroups=triceps&primaryColor=FF0000&secondaryColor=0000FF
/getIndividualColorImage?muscleGroups=biceps,triceps,abs&colors=FF0000,00FF00,0000FF
```

## Architecture

```
cmd/server/             # entry point: config, lifecycle, graceful shutdown
internal/
  assets/               # //go:embed PNG library — single binary, no disk I/O at runtime
  compositor/           # color parsing, asset library, image composition
  httpapi/              # HTTP handlers, router, middleware (logging, CORS, recover)
```

**Design highlights**

- **Zero external dependencies.** Standard library only — `net/http` with Go
  1.22+ pattern-matched routing, `log/slog` for structured logging,
  `image/png` + `image/draw` for compositing.
- **Assets are embedded** via `//go:embed` and decoded once at startup. No
  disk I/O on the request path.
- **Recoloring is O(1)** per muscle group. Source PNGs are 8-bit paletted;
  the highlight palette entry is swapped per-request without touching pixel
  data, so concurrent requests share the same `Pix` slice safely.
- **Graceful shutdown** on `SIGINT` / `SIGTERM` with a 15s drain window.
- **Distroless final image** (`gcr.io/distroless/static`), runs as
  non-root, ~10 MB total.

## Development

```bash
make test           # race detector + count=1
make cover          # coverage report
make lint           # golangci-lint
make build          # produces bin/mig
make docker         # local image
```

CI (`.github/workflows/ci.yml`) runs vet, golangci-lint, tests with race,
build, and Docker build on every push and pull request.

## Configuration

| Env var | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | TCP port to listen on |

## Deployment (Railway)

1. Connect this repository to a new Railway service.
2. Railway autodetects the `Dockerfile`.
3. Under **Settings → Networking**, generate a public domain and set the
   **target port to `8080`** (or whatever `PORT` you configure).

That's it — no Apache, no MPM, no PHP, no surprises.

## Credits

Forked from [MertenD/gym-api](https://github.com/MertenD/gym-api) (original
PHP implementation). All PNG assets © their original authors.
