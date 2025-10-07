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

clean:
	rm -rf $(BINDIR)
