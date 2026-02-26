.PHONY: dev-backend dev-frontend dev build tidy

# ── Development ───────────────────────────────────────────────────────────────

dev-backend:
	cd backend && go run ./cmd/server ../config.yaml

dev-frontend:
	cd frontend && npm run dev

# Run both in parallel (requires a POSIX shell with job control or tmux/foreman).
dev:
	$(MAKE) -j2 dev-backend dev-frontend

# ── Build ─────────────────────────────────────────────────────────────────────

build:
	cd backend  && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../dist/aperture ./cmd/server
	cd frontend && npm run build

# ── Docker ────────────────────────────────────────────────────────────────────

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

# ── Maintenance ───────────────────────────────────────────────────────────────

tidy:
	cd backend && go mod tidy
