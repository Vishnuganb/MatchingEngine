postgres:
	docker-compose up -d postgres

rabbitmq:
	docker-compose up -d rabbitmq

kafka:
	docker-compose up -d kafka zookeeper

matching-engine:
	docker-compose up -d matching-engine

createdb:
	docker exec -it matchingEnginePostgres createdb --username=postgres --owner=postgres orderManagement

dropdb:
	docker exec -it matchingEnginePostgres dropdb --username=postgres --if-exists orderManagement

migrateup:
	migrate -path ../internal/db/migration -database "postgresql://postgres:secret@localhost:5433/orderManagement?sslmode=disable" -verbose up

migratedown:
	migrate -path ../internal/db/migration -database "postgresql://postgres:secret@localhost:5433/orderManagement?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres rabbitmq kafka matching-engine createdb dropdb migrateup migratedown sqlc test