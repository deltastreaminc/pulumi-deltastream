# Data Model: Multi-Language Provider Build & Publish

**Feature**: 004-multi-language-provider-build
**Date**: 2026-06-29

This feature introduces no new DeltaStream resources, SQL entities, or Pulumi resource state. It is a build/tooling/CI feature. The "data model" here describes the **configuration entities** and their relationships.

---

## Toolchain Configuration

### Entity: `mise.toml` (Tool Manifest)

The canonical record of all tool versions. Read by mise at install time.

| Field | Type | Example | Notes |
|---|---|---|---|
| `tools.go` | semver string | `"1.25.11"` | Must match `go.mod` Go version |
| `tools.node` | semver string | `"22.23.1"` | Node 22 LTS (Jod); @pulumi/pulumi>=3.249 requires Node >=22 | |
| `tools.python` | semver string | `"3.11.15"` | Minimum version with SLSA attestations |
| `tools."vfox:version-fox/vfox-dotnet"` | semver string | `"8.0.20"` | vfox backend required for cross-platform .NET |
| `tools.java` | corretto string | `"corretto-11"` | Pulumi Java SDK build.gradle requires `languageVersion=11` |
| `tools."npm:yarn"` | semver string | `"1.22.22"` | Node package manager |
| `tools."aqua:gradle/gradle-distributions"` | semver string | `"8.14.3"` | Minimum available in aqua registry |
| `tools."github:pulumi/pulumi"` | semver string | `"3.246.0"` | Pinned Pulumi CLI (must match pulumi/sdk/v3 cap) |
| `tools."github:pulumi/pulumictl"` | semver string | `"0.0.50"` | Provider version utilities |
| `tools."github:pulumi/schema-tools"` | semver string | `"0.7.1"` | Schema diff tool |
| `tools.golangci-lint` | semver string | `"1.64.8"` | Go linter |
| `env.PULUMI_HOME` | path string | `"{{config_root}}/.pulumi"` | Isolates Pulumi CLI install |

**Location**: `.config/mise.toml` (authoritative) and `mise.toml` (root override for local customisation).

**Test overlay**: `.config/mise.test.toml` — activated by `MISE_ENV=test` in CI test jobs; adds `gotestsum`.

---

## Provider Build Artifacts

### Entity: Provider Binary

The compiled Go binary that implements the Pulumi provider gRPC server.

| Property | Value |
|---|---|
| Name | `pulumi-resource-deltastream` |
| Build path | `bin/pulumi-resource-deltastream` |
| Cross-compile outputs | `bin/<os>-<arch>/pulumi-resource-deltastream` |
| Build flags | `CGO_ENABLED=0 GOOS=<os> GOARCH=<arch>` |
| Version injection | `-ldflags "-s -w -X provider.Version=<version>"` |
| Entry point | `cmd/pulumi-resource-deltastream/main.go` |

### Entity: Provider Tarball

Release artifact consumed by the Pulumi CLI plugin system.

| Property | Format | Example |
|---|---|---|
| Filename | `pulumi-resource-deltastream-v<version>-<os>-<arch>.tar.gz` | `pulumi-resource-deltastream-v1.0.0-linux-amd64.tar.gz` |
| Contents | binary + `README.md` + `LICENSE` | — |
| Checksums file | `pulumi-deltastream_<version>_checksums.txt` | SHA256 per tarball |
| Download URL | `github://api.github.com/deltastreaminc` | Set in `schema.json` `pluginDownloadURL` |

**Supported platforms** (this feature):

| OS | Architecture |
|---|---|
| linux | amd64 |
| linux | arm64 |
| darwin | amd64 |
| darwin | arm64 |

---

## Language SDK Artifacts

Each language SDK is generated from `schema.json` and published to the corresponding package registry.

| Language | Generator | Package Registry | Package Name | SDK Output Path |
|---|---|---|---|---|
| TypeScript/Node | `pulumi package gen-sdk --language nodejs` | npm | `@deltastream/pulumi-deltastream` | `sdk/nodejs/` |
| Python | `pulumi package gen-sdk --language python` | PyPI | `pulumi-deltastream` | `sdk/python/` |
| .NET/C# | `pulumi package gen-sdk --language dotnet` | NuGet | `DeltaStream.Pulumi` | `sdk/dotnet/` |
| Go | `pulumi package gen-sdk --language go` | Go module (git tag) | `github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream` | `sdk/go/` |
| Java | `pulumi package gen-sdk --language java` | Maven Central (deferred) | `io.deltastream:pulumi-deltastream` | `sdk/java/` |

**Generated, not committed** (except Go): `sdk/nodejs/`, `sdk/python/`, `sdk/dotnet/`, `sdk/java/` are gitignored and regenerated each CI run. `sdk/go/` is committed.

---

## Schema Metadata

`schema.json` is the machine-readable provider contract. Relevant fields for publishing:

| Field | Value |
|---|---|
| `name` | `"deltastream"` |
| `displayName` | `"DeltaStream"` |
| `description` | `"A Pulumi native provider for DeltaStream..."` |
| `publisher` | `"DeltaStream Inc."` |
| `logoUrl` | `https://raw.githubusercontent.com/.../deltastream-logo.png` |
| `keywords` | `["pulumi", "deltastream", "category/database", "kind/native"]` |
| `pluginDownloadURL` | `"github://api.github.com/deltastreaminc"` |
| `language.csharp.packageName` | `"DeltaStream.Pulumi"` (see note below) |
| `language.java.basePackage` | `"io.deltastream.pulumi.deltastream"` |
| `language.python.packageName` | `"pulumi_deltastream"` |
| `language.nodejs.packageName` | `"@deltastream/pulumi-deltastream"` |
| `language.go.importBasePath` | `"github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"` |

**Note on `language.csharp.packageName`**: this schema field is documentation/metadata only — testing showed `pulumi package gen-sdk --language dotnet` does not use it to set the NuGet `PackageId`. The actual published package ID is pinned via a `Makefile` post-processing step that injects `<PackageId>DeltaStream.Pulumi</PackageId>` directly into the generated `.csproj` after codegen runs (avoids the reserved `Pulumi.*` NuGet ID prefix owned by `pulumi-bot`).

---

## CI Workflow DAG

Job dependency graph for the CI pipeline (`ci.yml`):

```
prerequisites  ──┬──► build_sdk (matrix: nodejs, python, dotnet, go, java)
                 │         │
                 │         └──► test (non-fork PRs only)
                 │
lint  ───────────┘
```

Job dependency graph for the release pipeline (`release.yml`):

```
prerequisites ──┬──► build_sdk ──┐
                │                ├──► test ──► publish ──► verify
                └──► build_provider ──┘
lint ───────────────────────────────────┘
```

---

## Makefile Sentinel State Machine

The `.make/` directory holds empty sentinel files. Each sentinel is `touch`-ed when its target completes, so `make` skips already-done work:

```
.make/mise_install   → mise install completed
.make/schema         → schema.json generated from provider binary
.make/generate_*     → per-language SDK source generated
.make/build_*        → per-language SDK compiled
.make/install_*      → per-language SDK installed locally
```

**Invalidation**: `make clean` removes all sentinels and built artifacts. CI jobs restore specific sentinels via `touch` after downloading pre-built artifacts.

---

## Secrets Inventory

| Secret Name | Used In | Purpose | Status |
|---|---|---|---|
| `GITHUB_TOKEN` | All workflows | GitHub Release, PR comments | Automatic |
| `CI_CREDENTIALS_YAML` | test jobs | Integration test credentials (DeltaStream server + API key) | Already present |
| `NPM_TOKEN` | `publish.yml` | Publish to npm registry | Already present |
| `PYPI_API_TOKEN` | `publish.yml` | Publish to PyPI | **Must be added** |
| `NUGET_USERNAME` | `publish.yml` | nuget.org username for OIDC token exchange (`NuGet/login@v1`) | **Must be added** |

> **NuGet Trusted Publishing**: The `publish_sdk` job uses `NuGet/login@v1` with `id-token: write` permission to exchange a GitHub OIDC token for a short-lived NuGet API key at runtime. No `NUGET_API_TOKEN` long-lived secret is needed. A one-time Trusted Publishing policy must be configured on nuget.org (see `plan.md` T-003).
