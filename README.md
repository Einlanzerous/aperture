# Aperture

A self-hosted, open-source server dashboard. Monitor the health of your services, Docker containers, local AI models, and host resources — all from a single clean UI.

> Inspired by Homer, built for more.

---

## Features

- **Service health checks** — HTTP status polling with response-time tracking
- **Docker container monitoring** — live state and healthcheck status via the Docker socket (no SDK overhead)
- **System resources** — CPU usage, RAM, and load averages sampled every 5 seconds
- **Ollama widget** — lists locally available models pulled from your Ollama instance
- **Flexible grid layout** — widgets come in three sizes: S (1/3), M (2/3), and L (full-width)
- **Zero-dependency frontend** — Vue 3 + TypeScript + Tailwind, no heavy UI frameworks
- **Single config file** — everything lives in `config.yaml`

---

## Quick start

### Docker Compose (recommended)

```bash
git clone https://github.com/aperture-dashboard/aperture
cd aperture

# Start from the provided example and fill in your own services.
cp config.example.yaml config.yaml

docker compose up -d
```

The dashboard is available at `http://localhost:4000`. The backend API runs on port `8888`.

### Manual

```bash
# Backend (requires Go 1.22+)
cd backend
go mod download
go run ./cmd/server ../config.yaml

# Frontend (separate terminal, requires Node 20+)
cd frontend
npm install
npm run dev
```

See [`backend/README.md`](backend/README.md) and [`frontend/README.md`](frontend/README.md) for full development guides.

---

## Configuration

All configuration lives in [`config.yaml`](config.yaml) at the project root.

```yaml
title: "Aperture"
port: 8888
check_interval: 30        # seconds between health-check cycles

services:
  - name: "Portainer"
    type: http             # http | docker
    url: "http://portainer:9000"
    category: "Management"
    size: s                # s | m | l

  - name: "Nginx"
    type: docker
    container: "nginx"
    size: s

ollama:
  url: "http://ollama:11434"   # omit or leave blank to hide the Ollama widget

system:
  enabled: true                # set to false to hide the resource widget
```

| Field            | Type     | Description                                          |
|------------------|----------|------------------------------------------------------|
| `title`          | string   | Dashboard name shown in the header                   |
| `port`           | int      | Port the backend HTTP server listens on              |
| `check_interval` | int      | Seconds between each full round of health checks     |
| `services[].type`| string   | `http` performs a GET request; `docker` inspects the socket |
| `services[].size`| string   | Grid column span: `s` = 1/3, `m` = 2/3, `l` = full  |

---

## Project structure

```
aperture/
├── config.yaml
├── docker-compose.yaml
├── Makefile
├── backend/            Go API server + health-check worker
└── frontend/           Vue 3 + TypeScript SPA
```

---

## Roadmap

- [ ] Persistent history / uptime graphs
- [ ] Custom widget icons
- [ ] Drag-and-drop layout editor
- [ ] Notification webhooks (Slack, Discord)
- [ ] Multi-host support

---

## Contributing

Pull requests are welcome. Please open an issue first for anything beyond small bug fixes so we can agree on the approach.

## License

MIT
