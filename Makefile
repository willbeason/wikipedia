BINARY := wikoptica
VERSION := $(shell git describe --always --long --dirty --tags)
LDFLAGS :=-ldflags "-X main.version=${VERSION}"

build:
	@go build ${LDFLAGS} -o ${BINARY} cmd/wikoptica/wikoptica.go

fmt:
	@golangci-lint run --fix

install:
	@go install ${LDFLAGS} cmd/wikoptica/wikoptica.go

lint:
	@golangci-lint run

test:
	@go test ./...

version:
	@echo "${VERSION}"
