.PHONY: test vet check run-server run-client

vet:
	go vet ./...

test:
	go test -race ./...

# What must be green before a commit.
check: vet test

run-server:
	go run ./cmd/server

run-client:
	go run ./cmd/client
