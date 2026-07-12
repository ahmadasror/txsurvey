.PHONY: run dev build fe-dev fe-build fe-test test lint check cover-check cover-update route-check tidy migrate-new up down docker logs

# --- Go ---
run: ## Run the API (loads .env if present; auto-runs migrations at boot)
	@bash -c 'set -a; [ -f .env ] && . ./.env; set +a; exec go run ./cmd/server'

dev: ## Run with hot-reload if 'air' is installed, else plain run
	@command -v air >/dev/null 2>&1 && air || go run ./cmd/server

build: fe-build ## Build the server binary with the SPA embedded
	CGO_ENABLED=0 go build -tags embedspa -o bin/server ./cmd/server

test: ## Run all Go tests
	go test ./...

lint: ## Vet + gofmt check
	go vet ./...
	@test -z "$$(gofmt -l . | grep -v '^frontend/')" || (echo "gofmt issues:" && gofmt -l . | grep -v '^frontend/' && exit 1)

fe-test: ## Frontend unit tests (vitest) — assumes frontend deps installed
	npm --prefix frontend run test

cover-check: ## Ratchet: fail if a tracked package's coverage regresses (scripts/coverage-baseline.json)
	python3 scripts/cover_check.py

cover-update: ## Re-bless the coverage baseline to current numbers
	python3 scripts/cover_check.py --update

route-check: ## Hard gate: every registered route has an FR endpoint or a waiver
	python3 scripts/route_check.py

check: ## Umbrella red/green: lint + unit tests + coverage + routes + docs + FE tests/build
	@set -e; \
		echo "== lint ==";             $(MAKE) -s lint; \
		echo "== go unit tests ==";    go test ./... -short; \
		echo "== coverage ratchet =="; $(MAKE) -s cover-check; \
		echo "== route -> FR ==";      $(MAKE) -s route-check; \
		echo "== docs ==";             $(MAKE) -s docs-check; \
		echo "== frontend unit ==";    $(MAKE) -s fe-test; \
		echo "== frontend build ==";   npm --prefix frontend run build; \
		echo; echo "check: PASS"

tidy:
	go mod tidy

# --- Lean SDD (spec-driven) ---
spec-validate: ## Hard gate: validate every FR contract block against the schema
	python3 scripts/validate_fr_contracts.py

spec-drift: ## Advisory: FR-declared endpoints/tables exist in code (exit 0)
	python3 scripts/spec_drift.py

spec-status docs-status: ## Show FRs by partition
	@for p in active done todo; do \
		echo "[$$p]"; ls docs/fr/survey/$$p/*.md 2>/dev/null | sed 's#.*/#  #' || echo "  (none)"; \
	done

docs-check: ## One-shot: schema + cross-link/index coherence (gates) + advisory drift
	@bash -c 'rc=0; \
		echo "== schema =="      ; python3 scripts/validate_fr_contracts.py || rc=1; echo; \
		echo "== coherence =="   ; python3 scripts/docs_check.py            || rc=1; echo; \
		echo "== drift (advisory) ==" ; python3 scripts/spec_drift.py; \
		echo; [ $$rc = 0 ] && echo "docs-check: PASS" || echo "docs-check: FAIL"; exit $$rc'

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
