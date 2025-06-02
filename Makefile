postgres:
	docker run --name matchingEngineContainer -p 5433:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=secret -d postgres:latest || true

createdb:
	docker exec -it matchingEngine createdb --username=root --owner=root userManagement

dropdb:
	docker exec -it matchingEngine dropdb --if-exists userManagement

migrateup:
	migrate -path internal/db/migration -database "postgresql://root:secret@localhost:5433/userManagement?sslmode=disable" -verbose up

migratedown:
	migrate -path internal/db/migration -database "postgresql://root:secret@localhost:5433/userManagement?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test