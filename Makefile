.PHONY: build test testacc lint clean install sweep

# Load .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Build the provider
build:
	go build -o bin/terraform-provider-unifi ./main.go

# Run unit tests
test:
	go test -v ./...

# Run acceptance tests (requires .env with UniFi credentials)
testacc:
	@if [ -z "$(UNIFI_BASE_URL)" ]; then \
		echo "Error: UNIFI_BASE_URL not set. Copy .env.example to .env and configure it."; \
		exit 1; \
	fi
	TF_ACC=1 go test -v ./internal/provider -timeout 60m

# Run a specific acceptance test
# Usage: make testacc-run TEST=TestAccNetworkResource_basic
testacc-run:
	@if [ -z "$(UNIFI_BASE_URL)" ]; then \
		echo "Error: UNIFI_BASE_URL not set. Copy .env.example to .env and configure it."; \
		exit 1; \
	fi
	TF_ACC=1 go test -v ./internal/provider -timeout 60m -run $(TEST)

# Run linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f terraform-provider-unifi

# Install the provider locally for testing
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/resnickio/unifi/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp bin/terraform-provider-unifi ~/.terraform.d/plugins/registry.terraform.io/resnickio/unifi/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

# Run sweepers to clean up test resources
# Usage: make sweep
sweep:
	@if [ -z "$(UNIFI_BASE_URL)" ]; then \
		echo "Error: UNIFI_BASE_URL not set. Copy .env.example to .env and configure it."; \
		exit 1; \
	fi
	go test ./internal/provider -v -sweep=all -timeout 30m

# Generate documentation (if using tfplugindocs)
docs:
	go generate ./...
