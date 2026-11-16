.PHONY: all build test clean install check fmt fmt-check vet test-race

BINARY_NAME=mdpreview
BIN_DIR=bin

all: check build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/${BINARY_NAME}

install:
	go install ./cmd/${BINARY_NAME}

fmt:
	gofmt -s -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Unformatted files found:" && gofmt -l . && exit 1)

vet:
	go vet ./...

test:
	go test -v ./...

test-race:
	go test -race ./...

check: fmt-check vet test test-race

clean:
	rm -rf $(BIN_DIR)
