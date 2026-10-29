BINARY_NAME := bay
DIST_DIR    := dist
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
PACKAGE_NAME := $(BINARY_NAME)-$(VERSION)-linux-amd64

.PHONY: all build package clean local

all: package

# linux amd64 only (release artifact)
build:
	@mkdir -p $(DIST_DIR)
	go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $(DIST_DIR)/$(BINARY_NAME) ./cmd/bay
	@echo "built: $(DIST_DIR)/$(BINARY_NAME)"

# binary + README.md
package: build
	@test -f README.md || (echo "README.md required" && exit 1)
	cp -f README.md $(DIST_DIR)/
	cd $(DIST_DIR) && tar -czf $(PACKAGE_NAME).tar.gz $(BINARY_NAME) README.md
	@echo "packed: $(DIST_DIR)/$(PACKAGE_NAME).tar.gz"

# host-native binary for local use / bay install
local:
	go build -o $(BINARY_NAME) ./cmd/bay
	@echo "built: ./$(BINARY_NAME)  (next: ./bay install)"

clean:
	rm -rf $(DIST_DIR) $(BINARY_NAME)
