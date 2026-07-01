# Provider cross-platform build & packaging
# Included by the root Makefile via: include scripts/crossbuild.mk
#
# Provides:
#   provider-<os>-<arch>          — cross-compile for a specific platform
#   provider_dist-<os>-<arch>     — compile + package as versioned tarball
#   provider_dist                 — all four platform tarballs

SHELL := /bin/bash -o pipefail

# Cross-compile targets: set GOOS/GOARCH per output path, then invoke the
# standard provider build. CGO_ENABLED=0 ensures static binaries.

bin/linux-amd64/$(PROVIDER):   GOOS := linux
bin/linux-amd64/$(PROVIDER):   GOARCH := amd64
bin/linux-arm64/$(PROVIDER):   GOOS := linux
bin/linux-arm64/$(PROVIDER):   GOARCH := arm64
bin/darwin-amd64/$(PROVIDER):  GOOS := darwin
bin/darwin-amd64/$(PROVIDER):  GOARCH := amd64
bin/darwin-arm64/$(PROVIDER):  GOOS := darwin
bin/darwin-arm64/$(PROVIDER):  GOARCH := arm64

bin/%/$(PROVIDER): .make/mise_install
	@mkdir -p $(dir $@)
	cd provider && GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 \
		go build $(PULUMI_PROVIDER_BUILD_PARALLELISM) \
		-o "$(WORKING_DIR)/$@" \
		-ldflags "$(LDFLAGS)" \
		$(PROJECT)/provider/cmd/$(PROVIDER)

provider-linux-amd64:  bin/linux-amd64/$(PROVIDER)
provider-linux-arm64:  bin/linux-arm64/$(PROVIDER)
provider-darwin-amd64: bin/darwin-amd64/$(PROVIDER)
provider-darwin-arm64: bin/darwin-arm64/$(PROVIDER)
.PHONY: provider-linux-amd64 provider-linux-arm64 provider-darwin-amd64 provider-darwin-arm64

# Versioned tarballs: pulumi-resource-deltastream-v<version>-<os>-<arch>.tar.gz
# Each tarball contains the binary plus README.md and LICENSE.

bin/$(PROVIDER)-v$(PROVIDER_VERSION)-linux-amd64.tar.gz:  bin/linux-amd64/$(PROVIDER)
bin/$(PROVIDER)-v$(PROVIDER_VERSION)-linux-arm64.tar.gz:  bin/linux-arm64/$(PROVIDER)
bin/$(PROVIDER)-v$(PROVIDER_VERSION)-darwin-amd64.tar.gz: bin/darwin-amd64/$(PROVIDER)
bin/$(PROVIDER)-v$(PROVIDER_VERSION)-darwin-arm64.tar.gz: bin/darwin-arm64/$(PROVIDER)

bin/$(PROVIDER)-v$(PROVIDER_VERSION)-%.tar.gz:
	tar --gzip -cf $@ README.md LICENSE -C $(dir $<) $(notdir $<)

provider_dist-linux-amd64:  bin/$(PROVIDER)-v$(PROVIDER_VERSION)-linux-amd64.tar.gz
provider_dist-linux-arm64:  bin/$(PROVIDER)-v$(PROVIDER_VERSION)-linux-arm64.tar.gz
provider_dist-darwin-amd64: bin/$(PROVIDER)-v$(PROVIDER_VERSION)-darwin-amd64.tar.gz
provider_dist-darwin-arm64: bin/$(PROVIDER)-v$(PROVIDER_VERSION)-darwin-arm64.tar.gz

provider_dist: provider_dist-linux-amd64 provider_dist-linux-arm64 \
               provider_dist-darwin-amd64 provider_dist-darwin-arm64
.PHONY: provider_dist-linux-amd64 provider_dist-linux-arm64 \
        provider_dist-darwin-amd64 provider_dist-darwin-arm64 provider_dist
