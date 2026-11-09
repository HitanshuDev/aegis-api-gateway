.PHONY: fmt test vet build run

fmt:
	gofmt -w .

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./...

run:
	go run ./cmd/aegis