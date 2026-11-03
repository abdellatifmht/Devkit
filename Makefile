.PHONY: build test run tidy

build:
	go build -ldflags "-X github.com/abdellatifmht/Brik/internal/cli.Version=$$(git describe --tags --always 2>/dev/null || echo dev)" -o bin/devkit ./cmd/devkit

test:
	go test ./...

run:
	go run ./cmd/devkit new

tidy:
	go mod tidy

clean:
	rm -rf bin/devkit
