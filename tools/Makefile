GOCMD = go
BINDIR = bin
BINARY = $(BINDIR)/server

.PHONY: all build run clean

all: build

build:
	@mkdir -p $(BINDIR)
	$(GOCMD) build -o $(BINARY) server.go
	@chmod +x $(BINARY)

run: build
	$(BINARY)

clean:
	rm -rf $(BINDIR)
