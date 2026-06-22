# ============================================================
#  NF-e Processor — Makefile
# ============================================================

.PHONY: help up down build logs \
        backend-dev backend-tidy backend-swagger backend-test \
        frontend-dev frontend-install frontend-build \
        db-shell mq-ui lint clean

# Default target
help:
	@echo ""
	@echo "  NF-e Processor"
	@echo ""
	@echo "  Infrastructure"
	@echo "    make up              Start all services (Docker Compose)"
	@echo "    make down            Stop and remove containers"
	@echo "    make build           Rebuild all Docker images"
	@echo "    make logs            Follow logs from all containers"
	@echo "    make logs-backend    Follow backend logs only"
	@echo ""
	@echo "  Backend"
	@echo "    make backend-dev     Run backend locally (needs postgres + rabbitmq running)"
	@echo "    make backend-tidy    go mod tidy"
	@echo "    make backend-swagger Regenerate Swagger docs"
	@echo "    make backend-test    Run tests"
	@echo ""
	@echo "  Frontend"
	@echo "    make frontend-install  npm install"
	@echo "    make frontend-dev      Vite dev server (http://localhost:5173)"
	@echo "    make frontend-build    Production build"
	@echo ""
	@echo "  Utilities"
	@echo "    make db-shell        psql into the running PostgreSQL container"
	@echo "    make mq-ui           Open RabbitMQ management UI in browser"
	@echo "    make clean           Remove containers, volumes and build artifacts"
	@echo ""

# ── Infrastructure ────────────────────────────────────────────

up:
	@cp -n .env.example .env 2>/dev/null || true
	docker compose up -d --build
	@echo ""
	@echo "  Frontend  → http://localhost:3000"
	@echo "  API       → http://localhost:8080/api/v1"
	@echo "  Swagger   → http://localhost:8080/swagger/index.html"
	@echo "  RabbitMQ  → http://localhost:15672  (guest / guest)"
	@echo ""

down:
	docker compose down

build:
	docker compose build --no-cache

logs:
	docker compose logs -f

logs-backend:
	docker compose logs -f backend

# ── Backend ───────────────────────────────────────────────────

backend-dev:
	@echo "Starting backend locally — ensure postgres and rabbitmq are running."
	cd backend && go run ./cmd/server

backend-tidy:
	cd backend && go mod tidy

backend-swagger:
	@which swag > /dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@latest
	cd backend && swag init -g cmd/server/main.go -o docs

backend-test:
	cd backend && go test ./... -v -race

# ── Frontend ──────────────────────────────────────────────────

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

# ── Utilities ─────────────────────────────────────────────────

db-shell:
	docker compose exec postgres psql -U $${DB_USER:-nfe} -d $${DB_NAME:-nfe_db}

mq-ui:
	@echo "Opening RabbitMQ management UI..."
	@open http://localhost:15672 2>/dev/null || xdg-open http://localhost:15672 2>/dev/null || \
		echo "Navigate to http://localhost:15672 (guest / guest)"

lint:
	cd backend && go vet ./...
	cd frontend && npx tsc --noEmit

clean:
	docker compose down -v --remove-orphans
	rm -rf backend/docs
	rm -rf frontend/dist frontend/node_modules
	@echo "Clean complete."
