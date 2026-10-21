# Research: Multi-Language Provider Build & Publish

**Feature**: 004-multi-language-provider-build
**Date**: 2026-06-29

This research phase consolidates findings from prior analysis sessions on pulumi-eks, pulumi-go-provider, ci-mgmt patterns, and the Pulumi publishing documentation. No NEEDS CLARIFICATION markers remain from the spec.

---

## Decision 1: Mise as Unified Tool Manager

**Decision**: Use `jdx/mise` as the single source of truth for all toolchain versions across local dev, CI, and devcontainer. Declare all tools in `.config/mise.toml`.

**Rationale**:
- Pulumi's own providers (pulumi-eks, pulumi-random, pulumi-command) all use this pattern via ci-mgmt.
- A single `mise install` installs Go, Node, Python, .NET, Java, Pulumi CLI, yarn, Gradle, golangci-lint, pulumictl, and schema-tools in one step.
- `jdx/mise-action` in CI replaces 5+ separate `setup-*` actions per job.
- Tool versions are trackable by Renovate as structured TOML keys.
- Works natively on developer machines without Docker; also used inside the devcontainer.

**Alternatives considered**:
- **Explicit multi-toolchain Dockerfile**: Larger, harder to maintain — every version bump requires Dockerfile edits _and_ CI workflow edits. Versions diverge between environments over time.
- **devbox/Nix**: Heavier dependency chain; pulumi-eks has already migrated away from it in favour of mise.
- **Per-job `setup-*` actions**: Already in place; creates 4-5 duplicated setup steps per job, no single source of truth.

---

## Decision 2: Thin Devcontainer (mise bootstraps all tools)

**Decision**: `.devcontainer/Dockerfile` installs only mise itself (8 lines, `FROM mcr.microsoft.com/devcontainers/base:ubuntu`). All tool installation is delegated to `mise install` which reads `.config/mise.toml`.

**Rationale**:
- Matches the exact pattern from `pulumi/pulumi-eks/.devcontainer/Dockerfile` (verified from source).
- Dockerfile never needs updating when tool versions change — only `mise.toml` does.
- VS Code Dev Containers and GitHub Codespaces both work out of the box.

**Alternatives considered**:
- **Large explicit Dockerfile** (original proposal): Pinning versions in both Dockerfile and mise.toml creates drift risk and maintenance burden.

---

## Decision 3: Reusable Workflow Decomposition

**Decision**: Decompose CI/release into six reusable `workflow_call` files: `prerequisites.yml`, `build_sdk.yml`, `build_provider.yml`, `lint.yml`, `publish.yml`, `verify-release.yml`. `ci.yml` and `release.yml` become thin orchestrators.

**Rationale**:
- Eliminates step duplication between CI and release workflows (currently, tool setup, build steps, and test steps are copy-pasted between `ci.yml` and `release.yml`).
- Matches the pulumi-eks/ci-mgmt pattern directly.
- Each job is independently cacheable and retryable.
- `build_sdk.yml` runs a language matrix — each of the 5 languages builds in parallel, independently.
- `build_provider.yml` cross-compiles all 4 platforms from a single `ubuntu-latest` runner (faster, cheaper than actual macOS runners for Go cross-compilation).

**Alternatives considered**:
- **Continue with monolithic ci.yml + release.yml**: Would require duplicating tool setup, artifact download, and credential handling in every new job added for dotnet/java.

---

## Decision 4: SDK Generation via `pulumi package gen-sdk`

**Decision**: Use `pulumi package gen-sdk --language <lang> schema.json` for all five languages. No separate codegen binary needed (unlike bridged providers).

**Rationale**:
- This repo uses `pulumi-go-provider v1.2.0` (infer/native), not Terraform bridge.
- The provider binary is the schema source. Schema is extracted via `pulumi package get-schema ./bin/pulumi-resource-deltastream`.
- `pulumi package gen-sdk` handles all five languages from the schema file.
- Confirmed correct by existing Makefile (already uses this approach for nodejs/go/python).

**Alternatives considered**:
- **Custom codegen binary** (`pulumi-gen-deltastream`): Only required for TF-bridged providers; unnecessary complexity here.

---

## Decision 5: Cross-Compilation From Ubuntu Only

**Decision**: All provider binaries (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64) are cross-compiled from `ubuntu-latest` GitHub Actions runners via `GOOS`/`GOARCH`. No actual macOS runners needed for binary compilation.

**Rationale**:
- Go supports full cross-compilation via env vars; `CGO_ENABLED=0` (already in Makefile) ensures no C dependency that would break cross-compilation.
- Eliminates the current `matrix.os` inconsistency where macOS runners are used only to compile Go binaries.
- Reduces CI cost (macOS runners are ~10× more expensive than Linux runners on GitHub Actions).
- `crossbuild.mk` provides clean named targets (`provider-linux-amd64`, etc.) that replace the inline `cp` hack in the current `release.yml`.

**Alternatives considered**:
- **Actual OS-specific runners**: Required for languages with native code (e.g., C extensions), but not for pure Go.

---

## Decision 6: SDK Publishing via `pulumi-package-publisher`

**Decision**: Use `pulumi/pulumi-package-publisher@v0.0.23` with `sdk: all,!java` to publish npm, PyPI, and NuGet in one action. Use `pulumi/publish-go-sdk-action@v1` for Go module tagging. Java Maven Central publishing is deferred (requires OSSRH credentials).

**Rationale**:
- This is the canonical approach used by all Pulumi-maintained providers.
- Single action handles npm, PyPI, and NuGet publish with consistent version handling.
- Go SDK requires a separate action that tags the repository to create the Go module version.
- Java Maven Central requires signed artifacts and OSSRH credentials — deferred to avoid blocking the initial multi-language release.

**Alternatives considered**:
- **Per-language publish steps** (current approach for nodejs only): Does not scale to 5 languages; version handling must be duplicated per language.

---

## Decision 7: Schema Metadata for Pulumi Registry

**Decision**: Add `publisher`, `logoUrl`, `keywords`, `description` to `schema.json`. Add complete language blocks for `csharp` (packageName: `DeltaStream.Pulumi`), `java` (basePackage: `io.deltastream.pulumi.deltastream`), and `python` (packageName: `pulumi_deltastream`).

**Rationale**:
- Pulumi Registry listing requires these fields per the [publishing guide](https://www.pulumi.com/docs/iac/guides/building-extending/packages/publishing-packages/).
- The NuGet package name `DeltaStream.Pulumi` avoids the `Pulumi.*` NuGet ID prefix, which is reserved on nuget.org for the `pulumi-bot` account (used by Pulumi Corp's own providers, e.g. `Pulumi.Aws`, `Pulumi.Azure`). Any new package submitted under a reserved prefix by a non-owning account is rejected by nuget.org with a 409 response; this was discovered when the initial `Pulumi.DeltaStream` package repeatedly failed to publish in CI (see Decision 7a below).
- The Python package name `pulumi_deltastream` follows PyPI snake_case convention.
- `keywords` array must include `category/database` and `kind/native` for correct registry classification.

**Alternatives considered**:
- `Pulumi.DeltaStream` (Pulumi ecosystem naming convention, e.g. `Pulumi.Aws`, `Pulumi.Azure`): Initially chosen to match the established convention, but rejected after discovering that `Pulumi.*` is a reserved NuGet ID prefix owned by `pulumi-bot`; nuget.org rejects any new package submission under this prefix from a different account.

---

## Decision 7a: NuGet Package ID vs. `language.csharp.packageName` schema field

**Decision**: Set `language.csharp.packageName` to `DeltaStream.Pulumi` for documentation/metadata purposes, but explicitly pin the actual NuGet `PackageId` in the generated `.csproj` via a `Makefile` post-processing step after `pulumi package gen-sdk --language dotnet` runs.

**Rationale**:
- Testing with `pulumi package gen-sdk` locally showed that the codegen's `language.csharp.packageName` schema field does **not** control the published NuGet package ID. The actual ID is derived from the generated `.csproj` filename, which in turn is derived from `language.csharp.rootNamespace` + the capitalized schema/module name (`Deltastream`) — always appending the module suffix, with no schema-only way to produce a clean two-segment name like `DeltaStream.Pulumi`.
- The robust fix is to explicitly set `<PackageId>DeltaStream.Pulumi</PackageId>` in the `.csproj` after generation, which is the standard MSBuild/NuGet mechanism for pinning a package ID independent of assembly name/namespace.
- `language.csharp.rootNamespace` is left unset (default), so the C# namespace remains whatever codegen derives (`Deltastream.Deltastream`) — this is only a source-code namespace and has no effect on the published package ID once `<PackageId>` is pinned.

**Alternatives considered**:
- Setting only `language.csharp.rootNamespace`/`packageName` in the schema and hoping the codegen'd filename matches: Rejected — verified empirically that no combination of `rootNamespace`/`packageName` produces a clean `DeltaStream.Pulumi.csproj` without the module-name suffix.

---

## Decision 8: Version Calculation

**Decision**: Use `pulumi/provider-version-action@v2` to derive the provider version from git tags, replacing the current inline shell version logic in `release.yml`.

**Rationale**:
- Eliminates 15 lines of version shell logic (`setver`, `prerelease`, `norm` steps).
- Handles semver normalization, prerelease detection, and `v` prefix stripping correctly.
- Used by all ci-mgmt-generated providers.
- Works for both push-tag triggers and `workflow_dispatch` manual triggers.

---

## Decision 9: Go Module Cache Strategy

**Decision**: Use `actions/setup-go` with `cache-dependency-path` pointing to all `go.sum` files for Go module caching. Use `actions/cache` directly for cross-compile jobs where only provider code is compiled.

**Rationale**:
- `actions/setup-go` with `cache-dependency-path` handles the multi-module case (root `go.sum` + `sdk/go/pulumi-deltastream/go.sum`).
- Cross-compile jobs can use a targeted cache keyed on `go.sum` + platform to avoid cache collisions.

---

## Resolved Clarifications

All questions from spec were pre-resolved via prior research sessions:

| Question | Resolution |
|---|---|
| .NET package name | `DeltaStream.Pulumi` (avoids reserved `Pulumi.*` NuGet ID prefix) |
| Java in default build | Yes, included in full build pipeline |
| Devcontainer approach | Explicit Dockerfile (thin, mise-bootstrapped) |
| Windows binaries | No, Linux + macOS only (consistent with existing matrix) |
| PyPI publishing | Yes, via `pulumi-package-publisher` |
| Workflow architecture | Reusable `workflow_call` files (pulumi-eks pattern) |

---

## Implementation Status

This feature's file-level implementation is **complete** as of the `/speckit.specify` execution. The following were created/modified:

| File | Status |
|---|---|
| `.config/mise.toml` | Created |
| `.config/mise.test.toml` | Created |
| `mise.toml` | Created |
| `.devcontainer/Dockerfile` | Created |
| `.devcontainer/devcontainer.json` | Created |
| `scripts/crossbuild.mk` | Created |
| `.golangci.yml` | Created |
| `Makefile` | Rewritten |
| `schema.json` | Updated (language metadata) |
| `.gitignore` | Updated (removed `.config/`, added `sdk/dotnet/`, `sdk/java/`, `.pulumi/`) |
| `.github/workflows/prerequisites.yml` | Created |
| `.github/workflows/build_sdk.yml` | Created |
| `.github/workflows/build_provider.yml` | Created |
| `.github/workflows/lint.yml` | Created |
| `.github/workflows/publish.yml` | Created |
| `.github/workflows/verify-release.yml` | Created |
| `.github/workflows/ci.yml` | Rewritten |
| `.github/workflows/release.yml` | Rewritten |
| `docs/_index.md` | Created |
| `docs/installation-configuration.md` | Created |

Remaining tasks are **validation** (verifying files work correctly together) and **documentation** of the plan artifacts.

---

## Implementation Corrections (discovered during container build validation)

The following issues were found and resolved while running `make build` inside the devcontainer. Each entry records the symptom, root cause, and fix applied.

---

### C-001: Makefile jq multiline single-quote breaks bash in recipe

**Symptom**: `make schema` fails with `unexpected EOF while looking for matching '\''` inside the container.

**Root cause**: A Makefile recipe using a multi-line single-quoted `jq` filter:
```makefile
$(PULUMI) package get-schema ... \
    | jq 'del(.version)
          | .publisher = "..."
          ...' \
    > $(SCHEMA_FILE)
```
Each continuation line in a Makefile recipe is a separate shell invocation joined with `;`. The single-quote opened on one line is not closed before the shell join boundary, breaking bash's quote parser.

**Fix**: Write the jq filter line-by-line with `printf` to a sentinel file (`.make/schema.jq`) and invoke `jq -f .make/schema.jq`:
```makefile
@printf '%s\n' \
    'del(.version)' \
    '| .publisher = "DeltaStream Inc."' \
    ... \
    > .make/schema.jq
$(PULUMI) package get-schema .make/$(PROVIDER) | jq -f .make/schema.jq > $(SCHEMA_FILE)
```

**Files changed**: `Makefile` (`.make/schema` target)

---

### C-002: Devcontainer Dockerfile `mise install` at build time ineffective

**Symptom**: Running `docker run` with the repo mounted fails — tools are not on PATH, `mise install` reports trust errors.

**Root cause**: The original Dockerfile ran `mise install` at image build time in `/code`, but VS Code mounts the workspace at `/workspaces/pulumi-deltastream`. Mise stores tools keyed to the project directory path, and the config files in the mounted workspace are untrusted at build time (the workspace isn't mounted during `docker build`).

**Fix**: Remove `WORKDIR /code` and `RUN mise install` from the Dockerfile. Move tool installation to `devcontainer.json`'s `postCreateCommand` which runs after the workspace is mounted:
```json
"postCreateCommand": "mise trust --yes && mise install"
```
The Dockerfile now only installs the mise binary itself and sets up PATH.

**Files changed**: `.devcontainer/Dockerfile`, `.devcontainer/devcontainer.json`

---

### C-003: GitHub API rate limit (403) during `mise install` in devcontainer

**Symptom**: `postCreateCommand` fails with `HTTP status client error (403 Forbidden)` on `github:*` and `aqua:*` tools. Unauthenticated requests from container IP ranges exhaust the 60 req/hr anonymous GitHub API limit before `mise install` completes.

**Root cause**: `github:` backends (`pulumi`, `pulumictl`, `schema-tools`) and `aqua:` backend (`gradle`) call `api.github.com/repos/<owner>/<repo>/releases?per_page=100` to resolve download URLs. No `GITHUB_TOKEN` is available inside a freshly started devcontainer.

**Investigation**: The mise docs (`/dev-tools/github-tokens.md`) explain that `mise-versions.jdx.dev` already caches version lists and proxies attestation verification — so most API calls are avoided. The remaining calls are for **download URL resolution** (one per tool install). The correct long-term fix is a lockfile, not forwarding a token.

**Fix**: Generate `mise.lock` (`.config/mise.lock` + `mise.lock`) once locally with `GITHUB_TOKEN` available, commit both files, and set `lockfile = true` in `.config/mise.toml`. With the lockfile in place, mise reads download URLs directly from disk — no GitHub API calls are needed for URL resolution. Attestation verification goes through `mise-versions.jdx.dev` (not rate-limited).

```bash
# Run once locally (requires GITHUB_TOKEN for initial generation):
GITHUB_TOKEN=$(gh auth token) mise lock
# Then commit .config/mise.lock and mise.lock
```

**Files changed**: `.config/mise.toml` (`lockfile = false` → `lockfile = true`), `.config/mise.lock` (new), `mise.lock` (new), `.devcontainer/devcontainer.json` (removed `remoteEnv`)

**What `GITHUB_TOKEN` is needed for in mise** (summary from source analysis):

| Use | Handled by lockfile? | Notes |
|---|---|---|
| Version list lookup (`releases?per_page=100`) | Yes — lockfile skips version enumeration | Also cached by `mise-versions.jdx.dev` |
| Download URL resolution (`releases/tags/v<version>`) | **Yes** — lockfile stores exact URL per platform | This was the rate-limit trigger |
| SLSA attestation verification | N/A — proxied via `mise-versions.jdx.dev` | No GitHub rate limit applies |

---

### C-004: Node.js GPG signature verification fails in container

**Symptom**: `core:node@20.19.5: gpg exited with non-zero status: exit code 2`

**Root cause**: Mise v2026+ verifies Node.js release tarballs using GPG signatures. It imports bundled Node.js release keys (`node.asc`, 28 keys) via `gpg --import` before verifying. In `mcr.microsoft.com/devcontainers/base:ubuntu`, the GPG keyring directory is not initialised, so `gpg --import` exits with code 2.

**Fix**: Add `node.gpg_verify = false` to `[settings]` in `.config/mise.toml`. The SHA256 checksum is still verified (mise always does this regardless of GPG), so download integrity is preserved. Only the additional GPG signature step is skipped.

**Files changed**: `.config/mise.toml`

---

### C-005: Python 3.11.8 fails SLSA attestation verification

**Symptom**: `No GitHub artifact attestations found for python@3.11.8` — even though the API returns HTTP 200, the attestations array is empty.

**Root cause**: `astral-sh/python-build-standalone` added SLSA attestation signing in mid-2024. Python 3.11.8 was released on 2024-02-24, before attestations were enabled. The SHA256 for the 3.11.8 tarball returns `{"attestations": []}` from the API. Mise v2026+ treats an empty attestations list as a failure (fail-safe: cannot distinguish "not published" from "deleted/tampered").

**Fix**: Upgrade to Python `3.11.15` — the latest 3.11 patch release, published in June 2026, which has 2 valid attestations.

**Files changed**: `.config/mise.toml` (`python = "3.11.8"` → `python = "3.11.15"`)

---

### C-006: Gradle `aqua:gradle/gradle-distributions@8.8` not in aqua registry

**Symptom**: `HTTP status client error (404 Not Found)` when mise tries to resolve Gradle 8.8 via the aqua backend.

**Root cause**: The aqua registry for `gradle/gradle-distributions` only carries versions from `8.10.2` onwards. Version `8.8` predates its inclusion.

**Fix**: Upgrade to `8.14.3` — latest stable Gradle release available in the aqua registry.

**Files changed**: `.config/mise.toml` (`"aqua:gradle/gradle-distributions" = "8.8"` → `"8.14.3"`)

---

### C-007: Java corretto-21 fails Gradle toolchain resolution for generated Java SDK

**Symptom**: `Cannot find a Java installation matching: {languageVersion=11}` when Gradle tries to compile the generated `sdk/java/`.

**Root cause**: The Pulumi Java SDK generator emits a `build.gradle` with `toolchain { languageVersion = JavaLanguageVersion.of(11) }`. Gradle's toolchain resolver cannot find a JDK 11 installation when only Java 21 (Corretto) is installed. Gradle toolchain auto-provisioning requires a download repository to be configured, which is not set up.

**Fix**: Change `java = "corretto-21"` to `java = "corretto-11"` in `.config/mise.toml`, matching the version that pulumi-eks uses and that the generated `build.gradle` requires.

**Files changed**: `.config/mise.toml` (`java = "corretto-21"` → `java = "corretto-11"`)

---

### C-008: Java SDK `com.pulumi:pulumi:0.10.0` missing `InvokeOutputOptions`

**Symptom**: 39 Java compile errors: `error: cannot find symbol — class InvokeOutputOptions`.

**Root cause**: The `schema.json` Java language block declared `"com.pulumi:pulumi": "0.10.0"` as the dependency. `InvokeOutputOptions` was introduced in `v0.11.0`. The generated `DeltastreamFunctions.java` and `Utilities.java` reference this class extensively.

**Fix**: Update the dependency to `"1.0.0"` — the version used by pulumi-eks, which includes `InvokeOutputOptions`.

**Files changed**: `schema.json` (`"com.pulumi:pulumi": "0.10.0"` → `"1.0.0"`), `Makefile` (jq filter updated to match)

---

### C-009: `pulumi-go-provider v1.3.x` incompatible with `pulumi/pkg v3.247.0+`

**Symptom**: Build error: `cannot use prov (variable of struct type schema.ResourceSpec) as *schema.ResourceSpec value in assignment` in `pulumi-go-provider@v1.3.2/middleware/schema/schema.go:351`.

**Root cause**: `pulumi/pkg v3.247.0` changed `PackageSpec.Provider` from `ResourceSpec` (value type) to `*ResourceSpec` (pointer). `pulumi-go-provider v1.3.2` was built against `v3.232.0` where the field was still a value type. All v1.3.x releases have this issue.

**Fix**: Cap `pulumi/pkg/v3` and `pulumi/sdk/v3` at `v3.246.0` — the last version before the breaking pointer change. This is the highest version compatible with `pulumi-go-provider v1.3.2`.

**Constraint**: When `pulumi-go-provider` releases a version built against `v3.247.0+`, this cap can be lifted.

**Files changed**: `go.mod` (`v3.210.0` → `v3.246.0`), also `go.sum` regenerated

---

### C-010: Pulumi CLI version in mise should match SDK version

**Observation**: The Pulumi CLI version pinned in `.config/mise.toml` was `3.216.0` (original), then `3.248.0` (latest), then corrected to `3.246.0` (to match the Go SDK). Keeping them in sync avoids schema format mismatches between the CLI used for `pulumi package gen-sdk` and the SDK the provider is compiled against.

**Fix**: Pin `"github:pulumi/pulumi" = "3.246.0"` to match `pulumi/sdk/v3 v3.246.0`.

**Files changed**: `.config/mise.toml`

---

### C-011: NuGet publishing via OIDC Trusted Publishing (no long-lived secret)

**Decision**: Use `NuGet/login@v1` (GitHub OIDC → short-lived API key) instead of storing a `NUGET_API_TOKEN` secret.

**How it works**: `NuGet/login@v1` exchanges the GitHub Actions OIDC JWT for a temporary NuGet API key (valid 1 hour). The key is passed to `pulumi-package-publisher` via `NUGET_PUBLISH_KEY=${{ steps.nuget-login.outputs.NUGET_API_KEY }}`. The job requires `id-token: write` permission.

**One-time setup on nuget.org**: Add a Trusted Publishing policy (Account → Trusted Publishing) with Repository Owner: `deltastreaminc`, Repository: `pulumi-deltastream`, Workflow File: `publish.yml`.

**Secrets required**: Only `NUGET_USERNAME` (the nuget.org profile name, not email). No `NUGET_API_TOKEN`.

**Files changed**: `.github/workflows/publish.yml` (added `NuGet/login@v1` step, `id-token: write` permission, replaced `secrets.NUGET_API_TOKEN` with step output)
