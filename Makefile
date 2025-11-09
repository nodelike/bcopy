.DEFAULT_GOAL := help

.PHONY: help build install run test clean release-test

help:
	@echo "bcopy - Available commands:"
	@echo "  make build        - Build binary to bin/bcopy"
	@echo "  make install      - Install to GOPATH"
	@echo "  make run          - Run without building"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make release-test - Test release build locally"

build:
	go build -o bin/bcopy ./cmd/bcopy

install:
	go install ./cmd/bcopy

run:
	go run ./cmd/bcopy

clean:
	rm -rf bin/ dist/

release-test:
	goreleaser release --snapshot --clean

