SHELL := bash
PACK := deltastream
ORG := deltastreaminc
PROJECT := github.com/$(ORG)/pulumi-$(PACK)
PROVIDER := pulumi-resource-$(PACK)
VERSION_PATH := provider.Version
WORKING_DIR := $(shell pwd)
TESTPARALLELISM ?= 10
GOTESTARGS ?=
PULUMI_PROVIDER_BUILD_PARALLELISM ?=

PROVIDER_VERSION ?= 1.0.0-alpha.0+dev
ifeq ($(shell echo $(PROVIDER_VERSION) | cut -c1),v)
$(error PROVIDER_VERSION should not start with a "v")
endif

LDFLAGS=-s -w -X $(PROJECT)/$(VERSION_PATH)=$(PROVIDER_VERSION)

_ := $(shell mkdir -p .make bin .pulumi/bin)

NODE_MODULE_NAME := @deltastream/pulumi-deltastream
SCHEMA_FILE := schema.json
# Pulumi CLI is now provided by CI (pulumi/actions) or developer environment.
PULUMI ?= pulumi

BUILD_OS ?= $(shell go env GOOS)
BUILD_ARCH ?= $(shell go env GOARCH)

define require_pulumi
    @ command -v $(PULUMI) >/dev/null 2>&1 || { echo "Pulumi CLI not found in PATH; install via pulumi/actions or local package." >&2; exit 1; }
endef

all: build schema generate build_sdks

.PHONY: build provider .FORCE
build: provider

.PHONY: provider
.FORCE:
provider: .FORCE bin/$(PROVIDER)
bin/$(PROVIDER):
	cd provider && GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) CGO_ENABLED=0 go build $(PULUMI_PROVIDER_BUILD_PARALLELISM) -o "$(WORKING_DIR)/bin/$(PROVIDER)" -ldflags "$(LDFLAGS)" $(PROJECT)/cmd/$(PROVIDER)

.PHONY: schema
schema: .make/schema

.make/schema: provider
	$(call require_pulumi)
	mkdir -p .make
	go build -o ".make/$(PROVIDER)" -ldflags "$(LDFLAGS)" $(PROJECT)/cmd/$(PROVIDER)
	$(PULUMI) package get-schema .make/$(PROVIDER) | jq 'del(.version) | .language.go.importBasePath = "$(PROJECT)/sdk/go/pulumi-$(PACK)" | .language.nodejs.packageName = "$(NODE_MODULE_NAME)"' > $(SCHEMA_FILE)
	@touch $@

.PHONY: generate
generate: generate_nodejs generate_go generate_python

# Build language SDKs (optional check step)
.PHONY: build_nodejs build_go build_python build_sdks
build_sdks: build_nodejs build_go build_python

build_nodejs:
	@if [ -d sdk/nodejs ]; then \
		cd sdk/nodejs && \
		( command -v yarn >/dev/null 2>&1 || (echo "yarn not found; please install yarn" && exit 1) ) && \
		yarn install && yarn run build; \
		cp ../../README.md ../../LICENSE package.json yarn.lock ./bin/; \
	else \
		echo "sdk/nodejs not found; run 'make generate_nodejs' first"; \
		exit 1; \
	fi

build_go:
	@if [ -d sdk/go ]; then \
		cd sdk && go list "$(PROJECT)/sdk/go/..." >/dev/null 2>&1 || true; \
	else \
		echo "sdk/go not found; run 'make generate_go' first"; \
	fi

build_python:
	@if [ -d sdk/python ]; then \
		echo "Python SDK sources present."; \
	else \
		echo "sdk/python not found; run 'make generate_python' first"; \
	fi

.PHONY: generate_nodejs
generate_nodejs: .make/generate_nodejs
.make/generate_nodejs: .make/schema
	mkdir -p .make
	$(PULUMI) package gen-sdk --language nodejs $(SCHEMA_FILE) --out sdk/ --version "$(PROVIDER_VERSION)"
	@if [ -f sdk/nodejs/package.json ]; then jq '.name = "$(NODE_MODULE_NAME)"' sdk/nodejs/package.json > sdk/nodejs/package.json.tmp && mv sdk/nodejs/package.json.tmp sdk/nodejs/package.json; fi
	@touch $@

.PHONY: generate_go
generate_go: .make/generate_go
.make/generate_go: .make/schema
	mkdir -p .make
	$(PULUMI) package gen-sdk --language go $(SCHEMA_FILE) --out sdk/ --version "$(PROVIDER_VERSION)"
	cd sdk/go/pulumi-deltastream && go mod init github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream && go mod tidy
	@touch $@

.PHONY: generate_python
generate_python: .make/generate_python
.make/generate_python: .make/schema
	mkdir -p .make
	$(PULUMI) package gen-sdk --language python $(SCHEMA_FILE) --out sdk/ --version "$(PROVIDER_VERSION)"
	@touch $@

.PHONY: clean
clean:
	rm -rf bin/ sdk/ .make/ $(SCHEMA_FILE)

# Install SDKs locally for development/testing
.PHONY: install_sdks install_nodejs_sdk install_go_sdk install_python_sdk
install_sdks: install_nodejs_sdk install_go_sdk install_python_sdk

install_nodejs_sdk: build_nodejs
	@if [ -d sdk/nodejs ]; then \
		echo "Linking local NodeJS SDK (@deltastream/pulumi-deltastream)..."; \
		yarn link --cwd $(WORKING_DIR)/sdk/nodejs/bin; \
	else \
		echo "sdk/nodejs not found; run 'make generate_nodejs' first"; \
		exit 1; \
	fi

install_go_sdk:
	@echo "Go SDK is consumed via module import; no install needed."

install_python_sdk:
	@if [ -d sdk/python ]; then \
		python3 -m venv sdk/python/.venv && \
		sdk/python/.venv/bin/python -m pip install --upgrade pip && \
		sdk/python/.venv/bin/pip install -e sdk/python; \
	else \
		echo "sdk/python not found; run 'make generate_python' first"; \
		exit 1; \
	fi

.PHONY: test
test: export PATH := $(WORKING_DIR)/bin:$(PATH)
test: install_go_sdk install_nodejs_sdk
	cd examples && go test -v -tags=nodejs -parallel $(TESTPARALLELISM) -timeout 2h $(value GOTESTARGS)

.PHONY: help
help:
	@echo "Main Targets"
	@echo "  build            Build the provider"
	@echo "  schema           Generate schema from provider"
	@echo "  generate         Generate SDKs (nodejs, go, python)"
	@echo "  build_sdks       Build all SDKs (sanity check)"
	@echo "  install_sdks     Install all SDKs for local testing (nodejs, go, python)"
	@echo "  test             Run the example tests (requires provider built)"
	@echo "  clean            Clean build artifacts"
