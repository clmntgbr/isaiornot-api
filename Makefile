COMPOSE_DEV := docker compose -f compose.dev.yaml

dev:
	$(COMPOSE_DEV) up -d

dev-down:
	$(COMPOSE_DEV) down

lint:
	$(COMPOSE_DEV) exec api golangci-lint run --fix

migrate:
	@echo "🔨 Building CLI..."
	@$(COMPOSE_DEV) exec api go build -o bin/cli ./cmd/cli
	@echo "🔄 Running migrate command..."
	@$(COMPOSE_DEV) exec api ./bin/cli migrate

retry-stale-scans:
	@$(COMPOSE_DEV) exec api go build -o bin/cli ./cmd/cli
	@$(COMPOSE_DEV) exec api ./bin/cli retry-stale-scans
