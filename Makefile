BINARY_NAME=mcp-wire
GO=/home/mitesh/Documents/cobalt/.tools/go/bin/go
LDFLAGS=-ldflags="-s -w"

.PHONY: all build clean test lint install

all: build

build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) main.go

clean:
	rm -f $(BINARY_NAME)

test:
	$(GO) test -v ./...

lint:
	$(GO) vet ./...

install: build
	cp $(BINARY_NAME) $(HOME)/go/bin/$(BINARY_NAME) || cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
