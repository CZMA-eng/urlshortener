# enter_bash
enter_bash:
	docker exec -it postgres_urls bash

# migrate
install_migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# sqlc
install_sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# postgres
launch_postgres:
	docker run --name postgres_urls \
	-e POSTGRES_USER=mcz \
	-e POSTGRES_PASSWORD=password \
	-e POSTGRES_DB=urldb \
	-p 5432:5432 \
	-d postgres

# redis
launch_redis:
	docker run --name=redis_urls \
	-p 6379:6379 \
	-d redis

# create_migration
create_migration:
	migrate create -ext=sql -dir=./database/migrate -seq init_schema

databaseURL="postgres://mcz:password@127.0.0.1:5432/urldb?sslmode=disable"

# migrate_up
migrate_up:
	migrate -path="./database/migrate" -database=${databaseURL} up 

# migrate_drop
migrate_down:
	migrate -path="./database/migrate" -database=${databaseURL} down - f

