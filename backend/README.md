# Aperture — Backend

Go HTTP server that runs health checks and exposes a JSON API consumed by the frontend.

## Requirements

- Go 1.22 or later
- Docker socket access (`/var/run/docker.sock`) — only needed if you have `docker`-type services in your config

## Development

```bash
# Install dependencies
go mod download

# Run against the root config.yaml
go run ./cmd/server ../config.yaml

# Or point at a custom config
go run ./cmd/server /path/to/config.yaml
```

The server starts on the port defined in `config.yaml` (default `8888`).

## Building

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o ../dist/aperture ./cmd/server
```

The binary is fully static and runs in a `scratch` container.

## API reference

All endpoints return `application/json`.

| Method | Path                    | Description                                              |
|--------|-------------------------|----------------------------------------------------------|
| `GET`  | `/api/health`           | Liveness probe — returns `{"status":"ok"}`               |
| `GET`  | `/api/config`           | Dashboard-level config for the frontend                  |
| `GET`  | `/api/services`         | Latest health-check result for all configured services   |
| `GET`  | `/api/system/resources` | Host CPU, memory, and load average                       |
| `GET`  | `/api/ollama/models`    | Proxied list of models from the Ollama `/api/tags` endpoint |

### `GET /api/services` response shape

```json
{
  "services": [
    {
      "name": "Portainer",
      "type": "http",
      "url": "http://portainer:9000",
      "status": "healthy",
      "statusCode": 200,
      "responseTime": 42,
      "checkedAt": "2024-01-01T12:00:00Z",
      "category": "Management",
      "size": "s"
    }
  ],
  "updatedAt": "2024-01-01T12:00:00Z"
}
```

Status values: `healthy` | `degraded` | `unhealthy` | `unknown`

### `GET /api/system/resources` response shape

```json
{
  "cpu":    { "percent": 12.4, "cores": 8 },
  "memory": { "total": 17179869184, "used": 9663676416, "free": 7516192768, "percent": 56.2 },
  "load":   { "load1": 1.23, "load5": 0.98, "load15": 0.75 },
  "updatedAt": "2024-01-01T12:00:00Z"
}
```

## Package layout

```
backend/
├── cmd/server/main.go              Entry point; wires config → worker → API server
└── internal/
    ├── config/config.go            Loads and validates config.yaml
    ├── checker/
    │   ├── types.go                Status enum + ServiceStatus struct
    │   ├── http.go                 HTTP health checker (10 s timeout, follows up to 5 redirects)
    │   ├── docker.go               Docker socket checker (raw HTTP, no SDK)
    │   └── worker.go               Fan-out goroutine pool + ticker-based scheduling
    ├── system/resources.go         CPU / RAM / load via gopsutil v3
    └── api/
        ├── handlers.go             HTTP handler functions
        └── router.go               Go 1.22 ServeMux wiring + CORS middleware
```

## Docker socket access

The `docker`-type checker dials `/var/run/docker.sock` directly using a custom `http.Transport`. No third-party Docker SDK is required. When running in a container, mount the socket read-only:

```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock:ro
```
