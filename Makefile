include .env
export

MIGRATE=migrate
MIGRATIONS_PATH=./migrations

DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up
migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down
migrate-create:
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

migrate-force:
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $(version)

run:
	go run ./cmd/server