.PHONY: setup run migrate docker-up docker-down

include .env
export

PSQL=PGPASSWORD=${POSTGRES_PASSWORD} psql -h ${POSTGRES_HOST} -p ${POSTGRES_PORT} -U ${POSTGRES_USER} -d ${POSTGRES_DB}

setup: docker-up db-create
	go mod tidy
	mkdir -p uploads

docker-up:
	docker-compose up -d
	# Wait for postgres to be ready
	sleep 1

docker-down:
	docker-compose down

# Database commands
db-create: docker-up
	$(PSQL) -f sql/create_tables.sql

db-drop: docker-up
	$(PSQL) -f sql/drop_tables.sql

db-truncate: docker-up
	$(PSQL) -f sql/truncate_tables.sql

# Connect to psql
psql:
	$(PSQL)

run:
	go run main.go

# Run all steps in sequence
start: setup run