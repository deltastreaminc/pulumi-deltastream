# Contract: Makefile Targets

**Feature**: 004-multi-language-provider-build
**Type**: Build system interface

This contract defines the stable Makefile targets that developers and CI workflows depend on. Any change to target names or semantics is a breaking change requiring a plan update.

---

## Primary Targets

| Target | Inputs | Outputs | Side Effects |
|---|---|---|---|
| `make build` | Source files | `bin/pulumi-resource-deltastream`, `sdk/*/` | Runs `mise install` if tools not present |
| `make provider` | `provider/**/*.go` | `bin/pulumi-resource-deltastream` | — |
| `make schema` | `bin/pulumi-resource-deltastream` | `schema.json`, `.make/schema` | Requires `pulumi` CLI on PATH |
| `make generate` | `schema.json` | `sdk/nodejs/`, `sdk/python/`, `sdk/dotnet/`, `sdk/go/`, `sdk/java/` | All five language SDKs generated |
| `make build_sdks` | `sdk/*/` | Compiled SDK artifacts | Runs tsc, dotnet build, go build, gradle build |
| `make install_sdks` | Compiled SDKs | Linked/installed SDKs for local testing | yarn link, dotnet nuget add source |
| `make test` | `bin/`, installed SDKs, `examples/credentials.yaml` | Test results | Requires live DeltaStream credentials |
| `make test_provider` | `provider/**/*.go` | `provider/coverage.txt` | No credentials needed |
| `make lint` | `provider/**/*.go`, `.golangci.yml` | Lint report | Exits non-zero on violations |
| `make lint.fix` | `provider/**/*.go`, `.golangci.yml` | Fixed source files | Modifies source in place |
| `make clean` | — | — | Removes `bin/`, `sdk/`, `.make/`, `schema.json` |
| `make schema` | — | `schema.json` | Requires `pulumi` CLI |

## Per-Language Targets

Each language follows this pattern:

| Target Pattern | Description |
|---|---|
| `make generate_<lang>` | Generate SDK source from `schema.json` |
| `make build_<lang>` | Compile generated SDK source |
| `make install_<lang>_sdk` | Install SDK for local test use |

Where `<lang>` ∈ `{nodejs, go, python, dotnet, java}`.

## Cross-Compile Targets (from `scripts/crossbuild.mk`)

| Target | Output |
|---|---|
| `make provider-linux-amd64` | `bin/linux-amd64/pulumi-resource-deltastream` |
| `make provider-linux-arm64` | `bin/linux-arm64/pulumi-resource-deltastream` |
| `make provider-darwin-amd64` | `bin/darwin-amd64/pulumi-resource-deltastream` |
| `make provider-darwin-arm64` | `bin/darwin-arm64/pulumi-resource-deltastream` |
| `make provider_dist-linux-amd64` | `bin/pulumi-resource-deltastream-v<version>-linux-amd64.tar.gz` |
| `make provider_dist-linux-arm64` | `bin/pulumi-resource-deltastream-v<version>-linux-arm64.tar.gz` |
| `make provider_dist-darwin-amd64` | `bin/pulumi-resource-deltastream-v<version>-darwin-amd64.tar.gz` |
| `make provider_dist-darwin-arm64` | `bin/pulumi-resource-deltastream-v<version>-darwin-arm64.tar.gz` |
| `make provider_dist` | All four tarballs |

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PROVIDER_VERSION` | `1.0.0-alpha.0+dev` | Injected into binary and SDK packages. Must NOT start with `v`. |
| `BUILD_OS` | `$(go env GOOS)` | Override for cross-compilation |
| `BUILD_ARCH` | `$(go env GOARCH)` | Override for cross-compilation |
| `PULUMI` | `pulumi` | Path to Pulumi CLI binary |
| `TESTPARALLELISM` | `10` | Parallel test workers |
| `GOTESTARGS` | `` | Extra args passed to `go test` |

---

## Sentinel File Protocol

CI jobs that download pre-built artifacts (instead of building from scratch) must `touch` the corresponding sentinel files to prevent make from re-running completed steps:

```bash
mkdir -p .make
# After downloading provider binary:
touch .make/schema           # schema was generated in prerequisites job
touch .make/mise_install     # mise was run in prerequisites job
# After downloading SDK artifacts for language X:
touch .make/generate_<lang>
touch .make/build_<lang>
```
