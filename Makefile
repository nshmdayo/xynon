.PHONY: build wasm test clean

# Directories
PLUGIN_SRC_DIR := examples/plugins
PLUGIN_OUT_DIR := examples/plugins

## build: compile the proxy binary (CGO_ENABLED=0)
build:
	CGO_ENABLED=0 go build -o bin/xynon ./cmd/xynon

## wasm: compile all TinyGo WASM plugins under examples/plugins/
wasm:
	@for dir in $(shell ls -d $(PLUGIN_SRC_DIR)/*/); do \
		name=$$(basename $$dir); \
		echo "Building plugin: $$name"; \
		cd /tmp && tinygo build \
			-o $(CURDIR)/$(PLUGIN_OUT_DIR)/$$name.wasm \
			-target wasip1 \
			$(CURDIR)/$(PLUGIN_SRC_DIR)/$$name/main.go && \
		cd $(CURDIR); \
	done

## test: run all Go tests
test:
	go test ./...

## test-e2e: run E2E tests
test-e2e:
	./scripts/test-e2e.sh

## clean: remove compiled artifacts
clean:
	rm -rf bin/
	find $(PLUGIN_OUT_DIR) -name '*.wasm' -delete
