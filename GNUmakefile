default: build

VERSION := 0.1.0
BINARY := terraform-provider-unifi
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)

build:
	go build -o $(BINARY)

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/resnickio/unifi/$(VERSION)/$(OS_ARCH)
	mv $(BINARY) ~/.terraform.d/plugins/registry.terraform.io/resnickio/unifi/$(VERSION)/$(OS_ARCH)/

test:
	go test -v ./...

testacc:
	TF_ACC=1 go test -v ./... -timeout 120m

generate:
	go generate ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)

.PHONY: build install test testacc generate fmt lint clean
