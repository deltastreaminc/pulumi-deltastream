# Implementation Plan: Multi-Language Provider Build & Publish

**Branch**: `004-multi-language-provider-build` | **Date**: 2026-06-29 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/004-multi-language-provider-build/spec.md`

## Summary

Transition the pulumi-deltastream provider to a production-grade multi-language build and release infrastructure. The primary changes are: (1) adopt `mise` as the unified toolchain manager replacing per-job `setup-*` actions, (2) add a devcontainer that bootstraps mise, (3) extend the Makefile to build all five language SDKs (nodejs, python, dotnet, go, java), (4) decompose CI/release into reusable `workflow_call` workflows modelled after `pulumi/pulumi-eks`, (5) update `schema.json` with complete language and registry metadata, and (6) author Pulumi Registry documentation files. All file-level implementation is complete. The plan validates correctness and defines remaining integration tasks.

## Technical Context

**Language/Version**: Go 1.25.6 (provider + Makefile); YAML (GitHub Actions); Makefile

**Primary Dependencies**:
- `jdx/mise-action@v2.2.3` — unified tool installation in CI
- `pulumi/provider-version-action@v2` — semver version calculation from git tags
- `pulumi/pulumi-package-publisher@v0.0.23` — multi-language SDK publish (npm, PyPI, NuGet)
- `pulumi/publish-go-sdk-action@v1` — Go SDK module tagging
- `softprops/action-gh-release@v2.2.1` — GitHub Release creation
- `golangci-lint v1.64.8` — Go linter
- `schema-tools v0.6.0` — Schema diff on PRs

**Storage**: `.make/` sentinel directory (build state cache); `sdk/*/` generated SDK artifacts

**Testing**: `make test_provider` (Go unit tests, no credentials); `make test` (integration tests, requires `examples/credentials.yaml`)

**Target Platform**: GitHub Actions (ubuntu-latest runners); VS Code Dev Containers; developer local machines (macOS/Linux)

**Project Type**: Build tooling / CI configuration (no new provider logic)

**Performance Goals**: CI PR run completes in under 20 minutes (SC-002); full release in under 30 minutes (SC-003)

**Constraints**: All platform binaries cross-compiled from linux only (CGO_ENABLED=0 required); no Windows binary support (out of scope); Java Maven Central publish deferred

**Scale/Scope**: 5 language SDKs × 4 platforms × 2 pipelines (CI + release)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Gate | Status | Note |
|---|------|--------|------|
| 1 | **Go only** — all implementation is in Go source under `provider/`; no generated files edited directly | ✅ | This feature makes no changes to `provider/` Go source. It modifies only build tooling, CI config, and `schema.json` language metadata. The schema metadata change is NOT a generated file edit — it is a human-authored configuration change that drives generation. |
| 2 | **SQL via MCP** — all DeltaStream SQL syntax and sys table schemas looked up via `deltastream-docs` MCP before implementation | ✅ N/A | This feature involves no SQL or DeltaStream API interaction. |
| 3 | **SQL documented** — every SQL statement used is from the SQL Reference; no custom/undocumented SQL | ✅ N/A | No SQL statements used. |
| 4 | **Sys tables used** — object status/readiness polled via `deltastream.sys.[object]` tables | ✅ N/A | No DeltaStream resources created or modified. |
| 5 | **Schema/SDK not edited** — `schema.json` and SDKs are generated artifacts; only Go source modified | ✅ | `schema.json` language metadata (`publisher`, `keywords`, `csharp`, `java`, `python.packageName`) are human-authored configuration fields, not generated output fields. The Makefile's `schema` target regenerates all other schema content from Go source and overwrites it — these metadata fields are preserved via the `jq` pipeline in the `schema` target. **Note**: The `schema` target must be verified to preserve these additions. |
| 6 | **Integration tests planned** — test coverage included in task list | ✅ | This feature is CI infrastructure, not provider logic. Integration tests are the CI jobs themselves (validated by the quickstart scenarios). Existing provider integration tests (`examples/`) continue to run unchanged. |
| 7 | **Single project** — no new project added | ✅ | No new Go module or project. `scripts/crossbuild.mk` is a Makefile include, not a separate project. |
| 8 | **Complexity justified** — abstractions documented | ✅ | The reusable workflow decomposition (6 workflow files) adds structural complexity justified by eliminating step duplication. Documented in research.md Decision 3. |
| 9 | **Pulumi config** — provider configuration via Pulumi config blocks, not hardcoded | ✅ N/A | No provider configuration changes. |

### Constitution Check Gate 5 — Critical Verification Required

The Makefile `schema` target uses `jq` to post-process the schema output from the provider binary:

```makefile
$(PULUMI) package get-schema .make/$(PROVIDER) \
  | jq 'del(.version) \
        | .language.go.importBasePath = "..." \
        | .language.nodejs.packageName = "..."' \
  > $(SCHEMA_FILE)
```

**Risk**: This `jq` pipeline only preserves `language.go.importBasePath` and `language.nodejs.packageName`. The new fields added to `schema.json` (`publisher`, `logoUrl`, `keywords`, `language.csharp.*`, `language.java.*`, `language.python.packageName`) will be **overwritten** when `make schema` is next run — because `pulumi package get-schema` regenerates the schema from the Go binary, which does not know about these registry metadata fields.

**Required Fix**: The `jq` pipeline in the Makefile `schema` target must be extended to also preserve/inject:
- `publisher`
- `logoUrl`
- `keywords`
- `displayName`
- `language.csharp.packageName`, `language.csharp.rootNamespace`
- `language.java.*`
- `language.python.packageName`

This is an **outstanding implementation task** (see Task T-001 below).

## Project Structure

### Documentation (this feature)

```text
specs/004-multi-language-provider-build/
├── plan.md                          # This file
├── research.md                      # Phase 0: decisions and rationale
├── data-model.md                    # Phase 1: configuration entities
├── quickstart.md                    # Phase 1: validation scenarios
├── contracts/
│   ├── makefile-targets.md          # Makefile target contract
│   └── github-actions-workflows.md  # Workflow I/O contract
└── tasks.md                         # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
# Build tooling
.config/
├── mise.toml           # All tool versions (authoritative)
└── mise.test.toml      # Test overlay (gotestsum)
mise.toml               # Root override for local customisation
.devcontainer/
├── Dockerfile          # Thin: installs mise, runs mise install
└── devcontainer.json   # VS Code devcontainer config
scripts/
└── crossbuild.mk       # Cross-compile targets (included by Makefile)
.golangci.yml           # golangci-lint rules
Makefile                # Build automation (rewritten)

# CI workflows
.github/workflows/
├── ci.yml              # PR orchestrator (thin, calls reusable workflows)
├── release.yml         # Release orchestrator (thin, calls reusable workflows)
├── prerequisites.yml   # Build provider + unit tests + schema + PR diff comment
├── build_sdk.yml       # Language matrix SDK build
├── build_provider.yml  # Platform matrix cross-compile
├── lint.yml            # golangci-lint
├── publish.yml         # GitHub Release + all SDK publish
└── verify-release.yml  # Post-release smoke tests

# Schema and registry
schema.json             # Provider schema (language metadata updated)
docs/
├── _index.md           # Pulumi Registry overview page
└── installation-configuration.md  # Pulumi Registry install/config page
```

## Complexity Tracking

| Pattern | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| 6 reusable `workflow_call` files | Eliminates step duplication between CI and Release pipelines; aligns with pulumi-eks standard | Monolithic `ci.yml`/`release.yml` would require duplicating all 5-language setup, artifact download, and credential steps in both files |
| `scripts/crossbuild.mk` include | Separates cross-compile concerns from main Makefile; keeps platform-specific targets clean | Inlining into main Makefile mixes SDK and binary build logic; existing `release.yml` had inline `cp` hacks that proved error-prone |
| `.make/` sentinel pattern | Prevents unnecessary rebuilds in incremental local dev; required for CI artifact restoration protocol | Without sentinels, `make build` would re-run `mise install`, re-compile provider, and re-run codegen even when artifacts already exist |

## Outstanding Implementation Tasks

These tasks were identified during the plan phase and must be completed for the feature to be production-ready.

### T-001: Fix `make schema` to Preserve Registry Metadata (CRITICAL)

**File**: `Makefile`, target `.make/schema`

**Problem**: The `jq` pipeline in the `schema` target only preserves `language.go.importBasePath` and `language.nodejs.packageName`. All other new metadata (`publisher`, `logoUrl`, `keywords`, `language.csharp.*`, `language.java.*`, `language.python.packageName`) will be lost on next `make schema` run.

**Fix required**: Extend the `jq` pipeline to inject all registry and language metadata:

```makefile
$(PULUMI) package get-schema .make/$(PROVIDER) \
  | jq 'del(.version)
        | .displayName = "DeltaStream"
        | .description = "A Pulumi native provider for DeltaStream..."
        | .publisher = "DeltaStream Inc."
        | .logoUrl = "https://raw.githubusercontent.com/deltastreaminc/pulumi-deltastream/main/docs/deltastream-logo.png"
        | .keywords = ["pulumi","deltastream","category/database","kind/native"]
        | .language.go.importBasePath = "$(PROJECT)/sdk/go/pulumi-$(PACK)"
        | .language.go.generateResourceContainerTypes = true
        | .language.go.respectSchemaVersion = true
        | .language.nodejs.packageName = "$(NODE_MODULE_NAME)"
        | .language.nodejs.respectSchemaVersion = true
        | .language.python.packageName = "pulumi_deltastream"
        | .language.python.respectSchemaVersion = true
        | .language.python.pyproject.enabled = true
        | .language.csharp.packageName = "Pulumi.DeltaStream"
        | .language.csharp.rootNamespace = "Pulumi"
        | .language.csharp.respectSchemaVersion = true
        | .language.java.basePackage = "io.deltastream.pulumi.deltastream"
        | .language.java.buildFiles = "gradle"
        | .language.java.gradleNexusPublishPluginVersion = "1.3.0"
        | .language.java.dependencies["com.pulumi:pulumi"] = "0.10.0"' \
  > $(SCHEMA_FILE)
```

### T-002: Add DeltaStream Logo

**File**: `docs/deltastream-logo.png` (or SVG)

**Problem**: `schema.json` references `https://raw.githubusercontent.com/.../docs/deltastream-logo.png` but this file does not yet exist. The Pulumi Registry will display a broken image.

**Fix required**: Add a DeltaStream logo file at `docs/deltastream-logo.png` (ideally SVG, with whitespace removed per Registry guidelines).

### T-003: Configure Package Registry Publishing Credentials

**Location**: GitHub repository Settings → Secrets and variables → Actions; nuget.org account settings

#### PyPI: Add `PYPI_API_TOKEN` repository secret

`publish.yml` uses `pypa/gh-action-pypi-publish` (via `pulumi-package-publisher`) which requires a PyPI API token.

1. Create an API token at https://pypi.org/manage/account/token/ scoped to `pulumi-deltastream`
2. Add as GitHub repository secret: `PYPI_API_TOKEN`

> **Future**: PyPI also supports OIDC Trusted Publishing (no token needed). This can be adopted later by adding `PYPI_OIDC_ENVIRONMENT` configuration to `pulumi-package-publisher`.

#### NuGet: Trusted Publishing (OIDC — no long-lived secret needed)

`publish.yml` uses `NuGet/login@v1` (OIDC token exchange) instead of a stored API key. The temporary key is valid for 1 hour and is obtained at workflow runtime.

**One-time setup on nuget.org** (requires nuget.org account ownership of `Pulumi.DeltaStream`):

1. Log into **nuget.org**
2. Navigate to your username → **Trusted Publishing**
3. Add a new trusted publishing policy:
   - **Repository Owner**: `deltastreaminc`
   - **Repository**: `pulumi-deltastream`
   - **Workflow File**: `publish.yml` *(filename only — do not include `.github/workflows/` prefix)*
   - **Environment**: *(leave empty)*
4. Add a single repository secret: `NUGET_USERNAME` — set to the nuget.org username (profile name, not email) that owns the `Pulumi.DeltaStream` package

> The `NUGET_API_TOKEN` secret is **not needed**. The `NuGet/login@v1` action exchanges the GitHub OIDC token for a short-lived key at publish time. The `publish_sdk` job has `id-token: write` permission for this purpose.

### T-004: Submit PR to Pulumi Registry

**Location**: https://github.com/pulumi/registry

**Problem**: `docs/_index.md` and `docs/installation-configuration.md` are authored but not yet submitted to the Pulumi Registry.

**Fix required**:
1. Fork `pulumi/registry`
2. Add `"deltastreaminc/pulumi-deltastream"` to `community-packages/package-list.json`
3. Copy `docs/_index.md` and `docs/installation-configuration.md` to `themes/default/content/registry/packages/deltastream/`
4. Open PR and request review

### T-005: Validate Java SDK Generation

**Problem**: Java SDK generation (`make generate_java`) requires `pulumi package gen-sdk --language java`. This requires the `pulumi-java` plugin to be available. The mise config installs Gradle (for building) but the Pulumi Java SDK generator may need to be installed separately.

**Fix required**: Verify that `make generate_java` succeeds in the devcontainer environment. The Java SDK generator may need to be installed via `pulumi plugin install` or may be bundled with the Pulumi CLI.

### T-006: Verify `provider-version-action@v2` Compatibility

**Problem**: `pulumi/provider-version-action@v2` is referenced in `prerequisites.yml` but its output format and `set-env` behaviour must be verified against the actual action version.

**Fix required**: Check that the action produces a version string without a leading `v` (the Makefile enforces this constraint with an error). Verify in a test run.
