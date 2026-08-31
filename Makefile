.PHONY: build build-reproducible test test-capture clean lint deps-proof run

BINARY := specula
VERSION := 0.1.0

build:
	go build -o $(BINARY) ./...

build-reproducible:
	CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-w -s -X main.version=$(VERSION)" -o $(BINARY) ./...

run:
	go run ./... --help

test:
	go test -race -v ./...

test-capture:
	go test -race -v ./... > test-results.txt 2>&1
	@echo "Test results written to test-results.txt"
	@cat test-results.txt

clean:
	rm -f $(BINARY)
	rm -f deps-proof.txt
	rm -f test-results.txt

lint:
	go vet ./...

deps-proof:
	go list -m all > deps-proof.txt
	@echo "deps-proof.txt generated:"
	@cat deps-proof.txt
