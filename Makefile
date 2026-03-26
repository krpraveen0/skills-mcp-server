.PHONY: all build test lint clean docker-up docker-down dev-backend dev-frontend

# ── Variables ─────────────────────────────────────────────────
GOPATH ?= $(shell go env GOPATH)
BACKEND_DIR = ./backend
FRONTEND_DIR = ./frontend

# ── Default ───────────────────────────────────────────────────
all: build

# ── Build ─────────────────────────────────────────────────────
build: build-backend build-frontend

build-backend:
	@echo "→ Building Go backend..."
	cd $(BACKEND_DIR) && go build -v ./cmd/server/...
	cd $(BACKEND_DIR) && go build -v ./cmd/worker/...

build-frontend:
	@echo "→ Building React frontend..."
	cd $(FRONTEND_DIR) && npm ci && npm run build

# ── Test ──────────────────────────────────────────────────────
test: test-backend test-frontend

test-backend:
	@echo "→ Running Go tests..."
	cd $(BACKEND_DIR) && go test -v -race ./...

test-frontend:
	@echo "→ Running frontend tests..."
	cd $(FRONTEND_DIR) && npm run test

# ── Lint ──────────────────────────────────────────────────────
lint: lint-backend lint-frontend

lint-backend:
	cd $(BACKEND_DIR) && go vet ./...

lint-frontend:
	cd $(FRONTEND_DIR) && npm run lint

# ── Docker ────────────────────────────────────────────────────
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

# ── Local Dev ─────────────────────────────────────────────────
dev-backend:
	@echo "→ Starting backend (requires Redis on :6379)..."
	cd $(BACKEND_DIR) && go run ./cmd/server/...

dev-worker:
	@echo "→ Starting background worker..."
	cd $(BACKEND_DIR) && go run ./cmd/worker/...

dev-frontend:
	@echo "→ Starting React dev server on :3000..."
	cd $(FRONTEND_DIR) && npm run dev

# ── Clean ─────────────────────────────────────────────────────
clean:
	rm -rf $(BACKEND_DIR)/data
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(FRONTEND_DIR)/node_modules
