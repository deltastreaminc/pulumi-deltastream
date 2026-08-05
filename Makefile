SHELL := bash
PACK        := deltastream
ORG         := deltastreaminc
PROJECT     := github.com/$(ORG)/pulumi-$(PACK)
PROVIDER    := pulumi-resource-$(PACK)
VERSION_PATH := provider.Version
WORKING_DIR  := $(shell pwd)
TESTPARALLELISM ?= 10
GOTESTARGS  ?=
PULUMI_PROVIDER_BUILD_PARALLELISM ?=

PROVIDER_VERSION ?= 1.0.0-alpha.0+dev
ifeq ($(shell echo $(PROVIDER_VERSION) | cut -c1),v)
$(error PROVIDER_VERSION should not start with a "v")
endif

LDFLAGS = -s -w -X $(PROJECT)/$(VERSION_PATH)=$(PROVIDER_VERSION)

NODE_MODULE_NAME := @deltastream/pulumi-deltastream
SCHEMA_FILE      := provider/cmd/$(PROVIDER)/schema.json
PULUMI           ?= pulumi

BUILD_OS   ?= $(shell go env GOOS)
BUILD_ARCH ?= $(shell go env GOARCH)

# Ensure sentinel and output directories exist before any target evaluates.
_ := $(shell mkdir -p .make bin .pulumi/bin provider/cmd/$(PROVIDER))

# ─── Mise integration ────────────────────────────────────────────────────────
# mise_install: installs all tools declared in .config/mise.toml.
# The sentinel file means 'mise install' only re-runs when tools actually change.
# mise_env: validates the mise environment on every invocation (order-only).

mise_install: .make/mise_install | mise_env
mise_env:
	@mise env -q > /dev/null
.make/mise_install:
	@mise install -q
	@touch $@
.PHONY: mise_install mise_env

# ─── Top-level aggregates ────────────────────────────────────────────────────

all: build schema generate build_sdks

build: .make/mise_install provider build_sdks
build: | mise_env

generate: generate_nodejs generate_go generate_python generate_dotnet generate_java

build_sdks: build_nodejs build_go build_python build_dotnet build_java

install_sdks: install_nodejs_sdk install_go_sdk install_python_sdk install_dotnet_sdk \
              install_java_sdk

.PHONY: all build generate build_sdks install_sdks

# ─── Provider binary ─────────────────────────────────────────────────────────

.PHONY: provider
provider: .make/mise_install bin/$(PROVIDER)
bin/$(PROVIDER):
	cd provider && GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) CGO_ENABLED=0 \
		go build $(PULUMI_PROVIDER_BUILD_PARALLELISM) \
		-o "$(WORKING_DIR)/bin/$(PROVIDER)" \
		-ldflags "$(LDFLAGS)" \
		$(PROJECT)/provider/cmd/$(PROVIDER)

# ─── Schema ──────────────────────────────────────────────────────────────────

.PHONY: schema
schema: .make/schema
.make/schema: provider
	@ command -v $(PULUMI) >/dev/null 2>&1 || \
		{ echo "Pulumi CLI not found; install via mise or pulumi/actions." >&2; exit 1; }
	mkdir -p .make
	go build -o ".make/$(PROVIDER)" -ldflags "$(LDFLAGS)" $(PROJECT)/provider/cmd/$(PROVIDER)
	@printf '%s\n' \
		'del(.version)' \
		'| .displayName = "DeltaStream"' \
		'| .description = "A Pulumi native provider for DeltaStream — manage databases, namespaces, stores, streams, queries, and applications as infrastructure."' \
		'| .publisher = "DeltaStream Inc."' \
		'| .logoUrl = "https://raw.githubusercontent.com/deltastreaminc/pulumi-deltastream/main/docs/deltastream-logo.png"' \
		'| .keywords = ["pulumi","deltastream","category/database","kind/native"]' \
		'| .language.go.importBasePath = "$(PROJECT)/sdk/go/pulumi-$(PACK)"' \
		'| .language.go.generateResourceContainerTypes = true' \
		'| .language.go.respectSchemaVersion = true' \
		'| .language.nodejs.packageName = "$(NODE_MODULE_NAME)"' \
		'| .language.nodejs.respectSchemaVersion = true' \
		'| .language.python.packageName = "pulumi_deltastream"' \
		'| .language.python.respectSchemaVersion = true' \
		'| .language.python.pyproject.enabled = true' \
		'| .language.csharp.packageName = "DeltaStream.Pulumi"' \
		'| .language.csharp.respectSchemaVersion = true' \
		'| .language.java.basePackage = "io.deltastream.pulumi.deltastream"' \
		'| .language.java.buildFiles = "gradle"' \
		'| .language.java.gradleNexusPublishPluginVersion = "1.3.0"' \
		'| .language.java.dependencies["com.pulumi:pulumi"] = "1.0.0"' \
		> .make/schema.jq
	$(PULUMI) package get-schema .make/$(PROVIDER) | jq -f .make/schema.jq > $(SCHEMA_FILE)
	@touch $@

# ─── NodeJS SDK ──────────────────────────────────────────────────────────────

.PHONY: generate_nodejs build_nodejs install_nodejs_sdk
generate_nodejs: .make/generate_nodejs
.make/generate_nodejs: .make/schema
	mkdir -p .make
	$(PULUMI) package gen-sdk --language nodejs $(SCHEMA_FILE) --out sdk/ \
		--version "$(PROVIDER_VERSION)"
	@if [ -f sdk/nodejs/package.json ]; then \
		jq '.name = "$(NODE_MODULE_NAME)"' sdk/nodejs/package.json \
			> sdk/nodejs/package.json.tmp \
			&& mv sdk/nodejs/package.json.tmp sdk/nodejs/package.json; \
	fi
	printf "module fake_nodejs_module // Exclude from Go tools\n\ngo 1.21\n" \
		> sdk/nodejs/go.mod
	@touch $@

build_nodejs: .make/build_nodejs
.make/build_nodejs: .make/generate_nodejs
	cd sdk/nodejs && yarn install && yarn run tsc
	cp README.md LICENSE sdk/nodejs/bin/
	# package.json is required in bin/ so that:
	#   1) require('./package.json') in utilities.js resolves for version detection
	#   2) pulumi-package-publisher can validate and publish the npm package
	# yarn.lock is copied so the bin/ dir is a self-contained npm package root.
	cp sdk/nodejs/package.json sdk/nodejs/yarn.lock sdk/nodejs/bin/
	@touch $@

install_nodejs_sdk: .make/install_nodejs_sdk
.make/install_nodejs_sdk: .make/build_nodejs
	yarn link --cwd $(WORKING_DIR)/sdk/nodejs/bin
	@touch $@

# ─── Go SDK ──────────────────────────────────────────────────────────────────

.PHONY: generate_go build_go install_go_sdk
generate_go: .make/generate_go
.make/generate_go: .make/schema
	mkdir -p .make
	$(PULUMI) package gen-sdk --language go $(SCHEMA_FILE) --out sdk/ \
		--version "$(PROVIDER_VERSION)"
	cd sdk/go/pulumi-$(PACK) \
		&& go mod init $(PROJECT)/sdk/go/pulumi-$(PACK) \
		&& go mod tidy
	@touch $@

build_go: .make/build_go
.make/build_go: .make/generate_go
	cd sdk/go/pulumi-$(PACK) && go build ./...
	@touch $@

install_go_sdk:
	@echo "Go SDK consumed via module import; no install step needed."

# ─── Python SDK ──────────────────────────────────────────────────────────────

.PHONY: generate_python build_python install_python_sdk
generate_python: .make/generate_python
.make/generate_python: .make/schema
	mkdir -p .make
	$(PULUMI) package gen-sdk --language python $(SCHEMA_FILE) --out sdk/ \
		--version "$(PROVIDER_VERSION)"
	cp README.md sdk/python/
	printf "module fake_python_module // Exclude from Go tools\n\ngo 1.21\n" \
		> sdk/python/go.mod
	@touch $@

build_python: .make/build_python
.make/build_python: .make/generate_python
	cd sdk/python && \
		rm -rf ./bin/ ../python.bin/ && cp -R . ../python.bin && mv ../python.bin ./bin && \
		rm -f ./bin/go.mod && \
		python3 -m venv venv && \
		./venv/bin/python -m pip install --quiet build==1.2.1 && \
		cd ./bin && ../venv/bin/python -m build .
	@touch $@

install_python_sdk:
	@if [ -d sdk/python ]; then \
		python3 -m venv sdk/python/.venv && \
		sdk/python/.venv/bin/python -m pip install --upgrade pip -q && \
		sdk/python/.venv/bin/pip install -e sdk/python -q; \
	else \
		echo "sdk/python not found; run 'make generate_python' first"; \
		exit 1; \
	fi

# ─── .NET SDK ────────────────────────────────────────────────────────────────

.PHONY: generate_dotnet build_dotnet install_dotnet_sdk
generate_dotnet: .make/generate_dotnet
.make/generate_dotnet: .make/schema
	mkdir -p .make
	$(PULUMI) package gen-sdk --language dotnet $(SCHEMA_FILE) --out sdk/ \
		--version "$(PROVIDER_VERSION)"
	echo "$(PROVIDER_VERSION)" > sdk/dotnet/version.txt
	printf "module fake_dotnet_module // Exclude from Go tools\n\ngo 1.21\n" \
		> sdk/dotnet/go.mod
	# The codegen'd .csproj derives its NuGet PackageId from the C# root
	# namespace/filename, not from language.csharp.packageName above, so pin
	# it explicitly here to avoid publishing under the reserved "Pulumi.*"
	# NuGet ID prefix (owned by pulumi-bot).
	@count=$$(find sdk/dotnet -maxdepth 1 -name '*.csproj' | wc -l | tr -d ' '); \
		if [ "$$count" -ne 1 ]; then \
			echo "generate_dotnet: expected exactly one .csproj in sdk/dotnet, found $$count" >&2; \
			exit 1; \
		fi; \
		csproj=$$(find sdk/dotnet -maxdepth 1 -name '*.csproj'); \
		if ! grep -q '<PackageId>' "$$csproj"; then \
			awk '{ print } /<GeneratePackageOnBuild>true<\/GeneratePackageOnBuild>/ && !done { print "    <PackageId>DeltaStream.Pulumi</PackageId>"; done=1 }' \
				"$$csproj" > "$$csproj.tmp" && mv "$$csproj.tmp" "$$csproj"; \
		fi
	@touch $@

build_dotnet: .make/build_dotnet
.make/build_dotnet: .make/generate_dotnet
	cd sdk/dotnet && dotnet build
	@touch $@

install_dotnet_sdk: .make/install_dotnet_sdk
.make/install_dotnet_sdk: .make/build_dotnet
	mkdir -p nuget
	find sdk/dotnet/bin -name '*.nupkg' -print -exec cp -p "{}" $(WORKING_DIR)/nuget \;
	dotnet nuget add source "$(WORKING_DIR)/nuget" --name "$(WORKING_DIR)/nuget" \
		2>/dev/null || true
	@touch $@

# ─── Java SDK ────────────────────────────────────────────────────────────────

.PHONY: generate_java build_java install_java_sdk
generate_java: .make/generate_java
.make/generate_java: .make/schema
	mkdir -p .make
	PULUMI_HOME=$(WORKING_DIR)/.pulumi \
		$(PULUMI) package gen-sdk --language java $(SCHEMA_FILE) --out sdk/ \
		--version "$(PROVIDER_VERSION)"
	printf "module fake_java_module // Exclude from Go tools\n\ngo 1.21\n" \
		> sdk/java/go.mod
	@touch $@

build_java: .make/build_java
.make/build_java: .make/generate_java
	cd sdk/java && gradle --console=plain build && gradle --console=plain javadoc
	@touch $@

install_java_sdk:
	@echo "Java SDK consumed via Maven/Gradle; no local install step needed."

# ─── Install SDKs ────────────────────────────────────────────────────────────

.PHONY: install_sdks install_nodejs_sdk install_go_sdk install_python_sdk \
        install_dotnet_sdk install_java_sdk

# ─── Testing ─────────────────────────────────────────────────────────────────

.PHONY: test test_provider
test: export PATH := $(WORKING_DIR)/bin:$(PATH)
test: install_nodejs_sdk install_go_sdk
	cd examples && go test -v -tags=all -parallel $(TESTPARALLELISM) \
		-timeout 2h $(value GOTESTARGS)

test_provider:
	cd provider && go test -v -short -parallel $(TESTPARALLELISM) \
		-coverprofile=coverage.txt ./...

# ─── Linting ─────────────────────────────────────────────────────────────────

.PHONY: lint lint.fix
lint: .make/mise_install | mise_env
	cd provider && golangci-lint run -c ../.golangci.yml ./...

lint.fix: .make/mise_install | mise_env
	cd provider && golangci-lint run -c ../.golangci.yml --fix ./...

# ─── Clean ───────────────────────────────────────────────────────────────────

.PHONY: clean
clean:
	rm -rf bin/ sdk/ .make/ $(SCHEMA_FILE)
	dotnet nuget remove source "$(WORKING_DIR)/nuget" 2>/dev/null || true

# ─── Cross-platform provider builds ──────────────────────────────────────────
# provider-<os>-<arch>, provider_dist-<os>-<arch>, and provider_dist aggregate
# are defined in scripts/crossbuild.mk.

include scripts/crossbuild.mk

# ─── Help ────────────────────────────────────────────────────────────────────

.PHONY: help
help:
	@echo "Main Targets"
	@echo "  build            Build provider + all SDKs (nodejs, go, python, dotnet, java)"
	@echo "  schema           Generate schema from provider binary"
	@echo "  generate         Generate all language SDK sources"
	@echo "  build_sdks       Compile all SDKs (sanity check)"
	@echo "  install_sdks     Install SDKs for local testing"
	@echo "  test             Run integration tests (requires built provider + installed SDKs)"
	@echo "  test_provider    Run provider unit tests (no credentials needed)"
	@echo "  lint             Run golangci-lint on provider Go code"
	@echo "  lint.fix         Run golangci-lint and auto-fix where possible"
	@echo "  clean            Remove all build artifacts"
	@echo ""
	@echo "Language Targets"
	@echo "  generate_nodejs / build_nodejs / install_nodejs_sdk"
	@echo "  generate_go     / build_go     / install_go_sdk"
	@echo "  generate_python / build_python / install_python_sdk"
	@echo "  generate_dotnet / build_dotnet / install_dotnet_sdk"
	@echo "  generate_java   / build_java   / install_java_sdk"
	@echo ""
	@echo "Cross-compile Targets (from scripts/crossbuild.mk)"
	@echo "  provider-linux-amd64 / provider-linux-arm64"
	@echo "  provider-darwin-amd64 / provider-darwin-arm64"
	@echo "  provider_dist         (all four tarballs)"
	@echo ""
	@echo "Mise Targets"
	@echo "  mise_install     Install all tools declared in .config/mise.toml"
