MIGRATIONS_DIR := migrations

.PHONY: migrate-up migrate-down migrate-force migrate-version

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" down 1

migrate-force:
	migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" force $(version)

migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" version

