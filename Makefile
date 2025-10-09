GOCMD = go
BINDIR = bin
BINARY = $(BINDIR)/cmiyc

.PHONY: all build run clean

all: build

build:
	@mkdir -p $(BINDIR)
	$(GOCMD) build -o $(BINARY) cmd/cmiyc/main.go
	@chmod +x $(BINARY)

run: build
	$(BINARY)

validate:
	$(GOCMD) test ./...
	$(GOCMD) vet ./...
	$(GOCMD) fmt ./...
	$(GOCMD) mod tidy
	$(GOCMD) mod verify

clean:
	rm -rf $(BINDIR)
