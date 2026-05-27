BINARY := bin/mig

.PHONY: run build test cover lint tidy docker clean

run:
	go run ./cmd/server

build:
	@mkdir -p bin
	go build -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/server

test:
	go test -race -count=1 ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run

tidy:
	go mod tidy

docker:
	docker build -t mig:dev .

clean:
	rm -rf bin coverage.out
