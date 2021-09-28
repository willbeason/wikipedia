MAIN = cmd/wikoptica/wikoptica.go
OUT = .out
BINARY := wikoptica
VERSION := $(shell git describe --always --long --dirty --tags)
LDFLAGS :=-ldflags "-X main.version=${VERSION}"

all: test build

build:
	@go build ${LDFLAGS} -o "${OUT}/${BINARY}" ${MAIN}

clean:
	@rm -rf ${OUT}

fmt:
	@golangci-lint run --fix

install:
	@go install ${LDFLAGS} ${MAIN}

lint:
	@golangci-lint run

test:
	@go test ./...

version:
	@echo "${VERSION}"
