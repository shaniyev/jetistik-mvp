.PHONY: up down logs migrate migrate-down migrate-new migrate-status sqlc build test-backend test-frontend lint seed sync-prod restore-data

## Infrastructure
up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

## Database
migrate:
	docker compose exec backend goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	docker compose exec backend goose -dir migrations postgres "$$DATABASE_URL" down

migrate-new:
	docker compose exec backend goose -dir migrations create $(N) sql

migrate-status:
	docker compose exec backend goose -dir migrations postgres "$$DATABASE_URL" status

## Code generation
sqlc:
	docker compose exec backend sqlc generate

## Build (production)
build:
	docker compose -f docker-compose.prod.yml build

## Testing
test-backend:
	docker compose exec backend go test ./...

test-frontend:
	docker compose exec frontend npm test

## Linting
lint:
	docker compose exec backend go vet ./...
	docker compose exec frontend npm run check

## Data
sync-prod:
	./scripts/sync-prod-data.sh

restore-data:
	./scripts/restore-prod-data.sh

seed: sync-prod restore-data
