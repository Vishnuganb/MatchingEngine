postgres:
	docker-compose up -d postgres

rabbitmq:
	docker-compose up -d rabbitmq

kafka:
	docker-compose up -d kafka zookeeper

Run-matching-engine:
	docker-compose up -d matching-engine

createdb:
	docker exec -it matchingEnginePostgres createdb --username=postgres --owner=postgres orderManagement

dropdb:
	docker exec -it matchingEnginePostgres psql -U postgres -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'orderManagement' AND pid <> pg_backend_pid();" || true
	docker exec -it matchingEnginePostgres dropdb --username=postgres --if-exists orderManagement

migrateup:
	migrate -path ../internal/db/migration -database "postgresql://postgres:secret@localhost:5433/orderManagement?sslmode=disable" -verbose up

migratedown:
	migrate -path ../internal/db/migration -database "postgresql://postgres:secret@localhost:5433/orderManagement?sslmode=disable" -verbose down

sqlc:
	sqlc generate

integration-test:
	docker compose exec matching-engine sh -c 'until nc -z kafka 9092; do echo "waiting for kafka..."; sleep 1; done'
	docker compose exec matching-engine sh -c 'until nc -z postgres 5432; do echo "waiting for postgres..."; sleep 1; done'
	docker compose exec matching-engine sh -c 'go test -v -tags integration -count=1 -timeout 300s ./integration/suite/$(TEST_FILE)'

.PHONY: postgres rabbitmq kafka Run-matching-engine createdb dropdb migrateup migratedown sqlc