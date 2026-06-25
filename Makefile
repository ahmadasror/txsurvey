.PHONY: run dev build fe-dev fe-build test lint tidy migrate-new up down docker logs

# --- Go ---
run: ## Run the API (auto-runs migrations at boot)
	go run ./cmd/server

dev: ## Run with hot-reload if 'air' is installed, else plain run
	@command -v air >/dev/null 2>&1 && air || go run ./cmd/server

build: fe-build ## Build the server binary with the SPA embedded
	CGO_ENABLED=0 go build -tags embedspa -o bin/server ./cmd/server

test: ## Run all Go tests
	go test ./...

lint: ## Vet + gofmt check
	go vet ./...
	@test -z "$$(gofmt -l . | grep -v '^frontend/')" || (echo "gofmt issues:" && gofmt -l . | grep -v '^frontend/' && exit 1)

tidy:
	go mod tidy

# --- Migrations ---
migrate-new: ## Scaffold the next migration pair: make migrate-new name=add_foo
	@test -n "$(name)" || (echo "usage: make migrate-new name=<snake_case>" && exit 1)
	@next=$$(ls internal/database/migrations/*.up.sql 2>/dev/null | sed -E 's/.*\/([0-9]+)_.*/\1/' | sort -n | tail -1); \
	next=$$(printf '%03d' $$((10#$${next:-0}+1))); \
	up=internal/database/migrations/$${next}_$(name).up.sql; \
	down=internal/database/migrations/$${next}_$(name).down.sql; \
	echo "-- $${next} — $(name)" > $$up; echo "-- down for $(name)" > $$down; \
	echo "created $$up and $$down"

# --- Frontend ---
fe-dev: ## Run the Vite dev server (proxies to API on :8080)
	npm --prefix frontend run dev

fe-build: ## Build the SPA and stage it at internal/web/dist for embedding
	@test -d frontend && ( \
		npm --prefix frontend ci && npm --prefix frontend run build && \
		rm -rf internal/web/dist && cp -r frontend/dist internal/web/dist \
	) || echo "no frontend/ yet, skipping"

# --- Docker ---
up: ## Start postgres + app
	docker compose up -d --build

down:
	docker compose down

docker:
	docker compose build

logs:
	docker compose logs -f app
