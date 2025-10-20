# Makefile for goose v3
.PHONY: db up down redo version force create

export DATABASE_URL=$(shell grep ^DATABASE_URL= .env | cut -d '=' -f2- | tr -d '"')
export MIGRATIONS_DIR=./db/postgresql/migrations

# Create migration: make db name=create_users
db:
	@echo "Creating migration: $(name)"
	goose create $(name) sql -dir $(MIGRATIONS_DIR)

# Run all up migrations
up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

# Rollback one step
down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down 

# Redo last migration
redo:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down 1
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

# Show current DB version
version:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" version

# Force migration version (interactive)
force:
	@read -p "Enter version to force to: " v; \
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" force $$v

# Interactive migration creation
create:
	@read -p "Migration name: " name; \
	goose create $$name sql -dir $(MIGRATIONS_DIR)
