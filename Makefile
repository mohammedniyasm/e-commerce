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

MINIO_BIN ?= $(shell go env GOPATH)/bin/minio
MINIO_DATA_DIR ?= $(HOME)/minio-data
MINIO_LOG_FILE ?= /tmp/ecommerce-minio.log

.PHONY: minio-up

minio-up:
	@if curl -fsS http://127.0.0.1:9000/minio/health/ready >/dev/null 2>&1; then \
		echo "MinIO is already running."; \
	else \
		if [ ! -x "$(MINIO_BIN)" ]; then \
			echo "MinIO binary not found: $(MINIO_BIN)"; \
			exit 1; \
		fi; \
		mkdir -p "$(MINIO_DATA_DIR)"; \
		echo "Starting MinIO..."; \
		nohup "$(MINIO_BIN)" server "$(MINIO_DATA_DIR)" \
			--address ":9000" \
			--console-address ":9001" \
			> "$(MINIO_LOG_FILE)" 2>&1 < /dev/null & \
		ready=0; \
		for i in $$(seq 1 30); do \
			if curl -fsS http://127.0.0.1:9000/minio/health/ready >/dev/null 2>&1; then \
				ready=1; \
				break; \
			fi; \
			sleep 1; \
		done; \
		if [ "$$ready" -ne 1 ]; then \
			echo "MinIO failed to start."; \
			tail -n 40 "$(MINIO_LOG_FILE)"; \
			exit 1; \
		fi; \
		echo "MinIO is ready."; \
	fi
run:minio-up
	go run ./cmd/server