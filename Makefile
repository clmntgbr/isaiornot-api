COMPOSE_DEV := docker compose -f compose.dev.yaml

# ============================================
# Development (compose.dev.yaml)
# ============================================

dev:
	$(COMPOSE_DEV) up -d

dev-down:
	$(COMPOSE_DEV) down

lint:
	$(COMPOSE_DEV) exec api golangci-lint run --fix

# ============================================
# CLI Commands (via Docker)
# ============================================

migrate:
	@echo "🔨 Building CLI..."
	@$(COMPOSE_DEV) exec api go build -o bin/cli ./cmd/cli
	@echo "🔄 Running migrate command..."
	@$(COMPOSE_DEV) exec api ./bin/cli migrate
