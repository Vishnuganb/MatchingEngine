unit-test:
	go test -v -cover -tags='!integration' ./...

import-reviser:
	goimports-reviser -rm-unused -set-alias -format ./...

lint:
	golangci-lint run --timeout 15m

sqlc:
	sqlc generate

.PHONY: unit-test import-reviser lint sqlc
