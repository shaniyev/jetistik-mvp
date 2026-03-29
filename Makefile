.PHONY: up down migrate migrate-down migrate-new migrate-status sqlc backend frontend test-backend test-frontend lint

DB_URL ?= postgres://jetistik:dev-password@localhost:5432/jetistik?sslmode=disable

## Infrastructure
up:
	docker compose up -d

down:
	docker compose down

## Database
migrate:
	cd backend && goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	cd backend && goose -dir migrations postgres "$(DB_URL)" down

migrate-new:
	cd backend && goose -dir migrations create $(N) sql

migrate-status:
	cd backend && goose -dir migrations postgres "$(DB_URL)" status

## Code generation
sqlc:
	cd backend && sqlc generate

## Development
backend:
	cd backend && air

frontend:
	cd frontend && npm run dev

## Testing
test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm test

## Linting
lint:
	cd backend && go vet ./...
	cd frontend && npm run check
