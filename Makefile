.PHONY: build test run clean install

BINARY_NAME=pricenexus
GO_CMD=go
MAIN_PATH=./cmd/cli

build: clean
	$(GO_CMD) build -o $(BINARY_NAME) $(MAIN_PATH)

test:
	$(GO_CMD) test -v ./...

run:
	$(GO_CMD) run $(MAIN_PATH)

clean:
	rm -f $(BINARY_NAME)

install:
	$(GO_CMD) install $(MAIN_PATH)

tidy:
	$(GO_CMD) mod tidy

deps:
	$(GO_CMD) mod download
