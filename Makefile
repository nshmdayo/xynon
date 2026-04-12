.PHONY: build plugins test clean

# Directories
GO123 := $(shell ls -d $(HOME)/.local/share/mise/installs/go/1.23.* 2>/dev/null | sort -V | tail -1)
TINYGO_ENV := GOROOT="$(GO123)" GOWORK=off PATH="$(GO123)/bin:$$PATH"
PLUGIN_SRC_DIR := examples/plugins
PLUGIN_OUT_DIR := examples/plugins

## build: compile the proxy binary (CGO_ENABLED=0)
build:
	CGO_ENABLED=0 go build -o bin/xynon ./cmd/xynon

## plugins: compile all TinyGo WASM plugins under examples/plugins/
plugins:
	@for dir in $(shell ls -d $(PLUGIN_SRC_DIR)/*/); do \
		name=$$(basename $$dir); \
		echo "Building plugin: $$name"; \
		cd /tmp && $(TINYGO_ENV) tinygo build \
			-o $(CURDIR)/$(PLUGIN_OUT_DIR)/$$name/$$name.wasm \
			-target wasip1 \
			$(CURDIR)/$(PLUGIN_SRC_DIR)/$$name/main.go && \
		cd $(CURDIR); \
	done

## test: run all Go tests
test:
	go test ./...

## clean: remove compiled artifacts
clean:
	rm -f bin/xynon
	find $(PLUGIN_OUT_DIR) -name '*.wasm' -delete
