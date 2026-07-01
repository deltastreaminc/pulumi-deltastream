# Feature Specification: Multi-Language Provider Build & Publish

**Feature Branch**: `004-multi-language-provider-build`

**Created**: 2026-06-29

**Status**: In Progress (US1 ✅, US2 ✅, US3 pending release secrets, US4 pending registry PR)

**Input**: User description: "based on above research, transition to using devcontainers, mise, multi-language builds, and updated CI jobs. also read https://www.pulumi.com/docs/iac/guides/building-extending/packages/publishing-packages/ and prepare to publish the pulumi provider"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Gets Working Environment Instantly (Priority: P1)

A new contributor clones the repository and opens it in VS Code. The devcontainer automatically sets up a complete, working development environment with all language toolchains (Go, Node.js, Python, .NET, Java) and tools (Pulumi CLI, yarn, Gradle, golangci-lint) without any manual installation steps. They can immediately run `make build` to compile the provider and generate all five language SDKs.

**Why this priority**: Contributor onboarding friction is the most direct blocker to community adoption. A reproducible environment eliminates the "works on my machine" problem and is the foundation for all other stories.

**Independent Test**: Clone the repo, open in VS Code devcontainer, run `make build` — all five SDKs must compile without errors.

**Acceptance Scenarios**:

1. **Given** a fresh clone of the repository, **When** a developer opens it in VS Code with the Dev Containers extension, **Then** the container builds and all required tools (`go`, `node`, `python3`, `dotnet`, `java`, `pulumi`, `golangci-lint`, `yarn`) are available on PATH without any additional setup.
2. **Given** a developer with mise installed locally (no Docker), **When** they run `mise install` in the repo root, **Then** all tools are installed at the pinned versions defined in `.config/mise.toml`.
3. **Given** a developer who runs `make build`, **When** the command completes, **Then** the provider binary exists in `bin/` and SDK source is generated under `sdk/nodejs/`, `sdk/python/`, `sdk/dotnet/`, `sdk/go/`, and `sdk/java/`.

---

### User Story 2 - Pull Request Build Validates All Languages (Priority: P2)

A contributor opens a pull request. The CI pipeline automatically builds the provider binary, generates and compiles all five language SDKs in parallel, runs the Go linter, runs provider unit tests, and reports schema changes — all without requiring any secrets or manual approval for non-fork PRs.

**Why this priority**: CI that covers all target languages catches SDK generation regressions before they reach main. It also provides the schema diff feedback loop that prevents breaking changes from shipping silently.

**Independent Test**: Open a PR with a trivial Go change — all five language build jobs must pass and a schema diff comment must appear on the PR.

**Acceptance Scenarios**:

1. **Given** a pull request is opened against main, **When** the CI workflow runs, **Then** the provider binary is built, all five language SDKs are generated and compiled in parallel (nodejs, python, dotnet, go, java), and all jobs report success.
2. **Given** a pull request that modifies the provider schema, **When** the CI workflow runs, **Then** a comment is posted on the PR summarising the schema diff against the current main branch.
3. **Given** a PR from a forked repository, **When** the CI workflow runs, **Then** the build and SDK generation jobs run (no secrets needed) but integration tests requiring credentials are skipped automatically.
4. **Given** a PR with a Go linting violation, **When** the lint job runs, **Then** the job fails and reports the specific violation to the contributor.
5. **Given** a PR with a failing provider unit test, **When** the prerequisites job runs, **Then** the pipeline fails before SDK generation begins.

---

### User Story 3 - Release Publishes All Language Packages (Priority: P2)

A maintainer pushes a version tag (e.g., `v1.2.0`). The release pipeline cross-compiles provider binaries for all supported platforms (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64), creates a GitHub Release with tarballs and checksums, and publishes all five language SDKs to their respective package registries (npm, PyPI, NuGet, Maven Central, Go module tag).

**Why this priority**: Publishing all language SDKs is the core deliverable. Users of each language must be able to install the provider through their standard package manager.

**Independent Test**: Tag a prerelease version — GitHub Release is created with binaries, and all five package registries receive the new version.

**Acceptance Scenarios**:

1. **Given** a maintainer pushes a `v*.*.*` tag, **When** the release workflow completes, **Then** a GitHub Release exists containing provider tarballs for linux-amd64, linux-arm64, darwin-amd64, and darwin-arm64, plus a SHA256 checksums file and `schema.json`.
2. **Given** a successful release build, **When** the publish step runs, **Then** `@deltastream/pulumi-deltastream` is available on npm, `pulumi-deltastream` is available on PyPI, `Pulumi.DeltaStream` is available on NuGet, and the Go SDK is tagged in the repository.
3. **Given** a version tag containing a hyphen (e.g., `v1.2.0-alpha.1`), **When** the release workflow runs, **Then** it is treated as a prerelease and the GitHub Release is created as a draft rather than a public release.
4. **Given** a completed release publication, **When** the verify job runs, **Then** a smoke test successfully installs and imports the Node.js SDK from npm using the published version.

---

### User Story 4 - Provider Discoverable in Pulumi Registry (Priority: P3)

The DeltaStream provider is listed in the public Pulumi Registry with a complete overview page, installation and configuration guide, and automatically generated API documentation for all resources and functions.

**Why this priority**: Registry listing is important for community discovery but does not block users who know the provider exists — they can install it directly. It requires a one-time PR to the pulumi/registry repository.

**Independent Test**: The provider appears at `pulumi.com/registry/packages/deltastream` with overview, installation, and API docs tabs populated.

**Acceptance Scenarios**:

1. **Given** the provider repository contains `docs/_index.md` and `docs/installation-configuration.md`, **When** a PR is submitted to pulumi/registry adding the provider to the community package list, **Then** the Pulumi team can merge it and the registry listing becomes live.
2. **Given** the schema.json contains `publisher`, `logoUrl`, `keywords`, and `description` metadata, **When** the registry page renders, **Then** the provider appears with the correct branding and category.

---

### Edge Cases

- What happens when a language SDK build fails in the CI matrix? The failing language job is reported independently; other language jobs continue (fail-fast disabled for Renovate PRs, enabled otherwise).
- What happens if mise cannot install a tool due to a network issue in CI? The mise-action retries up to 3 times (`http_retries = 3`) before failing; the job fails with a clear error.
- What happens when a release tag is pushed but tests have not passed? The publish job depends on the test job — if tests fail, publish does not run.
- What happens when a fork PR contains a schema change? Schema diff is computed and posted as a PR comment using locally built schema; no secrets are required for this step.
- What happens if dotnet or Java build fails but other languages succeed? Each language is an independent matrix job; failures are reported per-language without blocking other languages.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The repository MUST include a devcontainer configuration that provides a complete, reproducible development environment containing all language toolchains and tools required to build the provider and all language SDKs.
- **FR-002**: All tool versions MUST be declared in a single mise configuration file (`.config/mise.toml`) that serves as the authoritative source of truth for local development, CI, and devcontainer environments.
- **FR-003**: The Makefile MUST support generating and building SDKs for all five languages: nodejs, python, dotnet, go, and java — each as an independent target with make sentinel caching.
- **FR-004**: The Makefile MUST include named cross-compilation targets (`provider-<os>-<arch>`) and a `provider_dist` aggregate target that produces versioned tarballs ready for release.
- **FR-005**: The Makefile MUST include `lint` and `lint.fix` targets that run golangci-lint against the provider Go code.
- **FR-006**: The Makefile MUST include a `test_provider` target that runs Go unit tests against the provider package without requiring external credentials.
- **FR-007**: The CI workflow MUST replace all per-language tool setup steps with a single `jdx/mise-action` invocation, with tool cache saved only in the prerequisites job.
- **FR-008**: The CI pipeline MUST be decomposed into reusable `workflow_call` files: `prerequisites.yml`, `build_sdk.yml`, `build_provider.yml`, `lint.yml`, `publish.yml`, and `verify-release.yml`.
- **FR-009**: The `build_sdk.yml` workflow MUST use a language matrix (`nodejs`, `python`, `dotnet`, `go`, `java`) so each language builds independently and in parallel.
- **FR-010**: The `build_provider.yml` workflow MUST cross-compile all platform binaries from `ubuntu-latest` runners only, covering linux-amd64, linux-arm64, darwin-amd64, and darwin-arm64.
- **FR-011**: The `prerequisites.yml` workflow MUST run `make test_provider` (unit tests) before SDK generation and post a schema diff comment on pull requests.
- **FR-012**: The `publish.yml` workflow MUST publish all five language SDKs using `pulumi/pulumi-package-publisher` with `sdk: all`, publish the Go SDK via `pulumi/publish-go-sdk-action`, and create a GitHub Release with provider tarballs, checksums, and `schema.json`.
- **FR-013**: The `verify-release.yml` workflow MUST run smoke tests for the Node.js and Python SDKs after publication to confirm the published packages are importable.
- **FR-014**: The `schema.json` MUST include complete language metadata for all five languages: nodejs (`packageName`), python (`pyproject`), dotnet (`packageName`, `rootNamespace`), go (`importBasePath`), and java (`basePackage`).
- **FR-015**: Fork PRs MUST trigger the build and lint jobs but NOT the integration test job (which requires secrets).
- **FR-016**: The `.gitignore` MUST NOT exclude `.config/` (required for mise configuration) and MUST exclude generated SDK directories `sdk/dotnet/`, `sdk/java/`.
- **FR-017**: The repository MUST contain `docs/_index.md` and `docs/installation-configuration.md` suitable for submission to the Pulumi Registry.
- **FR-018**: The `schema.json` MUST contain `publisher`, `logoUrl`, `keywords` (category and kind), and `description` metadata required for Pulumi Registry listing.

### Key Entities

- **mise.toml**: Declarative tool version manifest read by mise to install all development and CI toolchains. Lives in `.config/mise.toml` with an optional root-level `mise.toml` for local overrides.
- **devcontainer**: VS Code/GitHub Codespaces containerized development environment that bootstraps mise and delegates all tool installation to it.
- **Reusable Workflow**: GitHub Actions workflow file callable via `workflow_call`, enabling DRY CI pipelines where `ci.yml` and `release.yml` are thin orchestrators.
- **SDK artifact**: Language-specific generated package (npm, wheel/sdist, nupkg, go module, jar) uploaded as a CI artifact and then published to the corresponding package registry.
- **Provider tarball**: Compressed archive (`pulumi-resource-deltastream-v<version>-<os>-<arch>.tar.gz`) containing the provider binary, uploaded to a GitHub Release as the plugin download target.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer with only Docker and VS Code installed can go from fresh clone to a passing `make build` (all five language SDKs) in under 15 minutes on a standard laptop.
- **SC-002**: A pull request CI run completes all five language SDK builds in parallel and reports results in under 20 minutes.
- **SC-003**: A full release pipeline (build all platforms + publish all SDKs) completes end-to-end in under 30 minutes from tag push to GitHub Release creation.
- **SC-004**: All five language SDKs (`@deltastream/pulumi-deltastream` on npm, `pulumi-deltastream` on PyPI, `Pulumi.DeltaStream` on NuGet, Go module, Java on Maven) are installable by users using standard package manager commands within 10 minutes of a release completing.
- **SC-005**: The provider appears in the Pulumi Registry at `pulumi.com/registry/packages/deltastream` following a one-time registry PR submission.
- **SC-006**: CI tool setup time is reduced by eliminating per-language setup actions — all tools installed via a single mise step with cache restoration.

## Assumptions

- The DeltaStream provider is hosted at `github.com/deltastreaminc/pulumi-deltastream` and GitHub Actions is the CI platform.
- Pulumi ESC is not available; secrets are stored as GitHub repository secrets (`NPM_TOKEN`, `PYPI_API_TOKEN`, `NUGET_USERNAME`, `CI_CREDENTIALS_YAML`). NuGet publishing uses OIDC Trusted Publishing (`NuGet/login@v1`) — no long-lived `NUGET_API_TOKEN` is stored.
- Windows binary cross-compilation is out of scope for this feature (linux and macOS only, consistent with the existing release matrix).
- Java SDK publishing to Maven Central requires additional OSSRH/Sonatype credentials (`OSSRH_USERNAME`, `OSSRH_PASSWORD`, `JAVA_SIGNING_KEY*`) — these are noted as future requirements; Java SDK build is included but Maven publish is deferred.
- The Go SDK is published by tagging the repository (not a separate module repository) using `pulumi/publish-go-sdk-action`.
- `pulumi/provider-version-action@v2` is used to calculate the provider version from git tags, replacing the existing inline shell version logic.
- The devcontainer uses `mcr.microsoft.com/devcontainers/base:ubuntu` as the base image (consistent with pulumi-eks).
- Pulumi Registry listing requires a separate PR to the `pulumi/registry` repository; this spec covers authoring the required `docs/` files and schema metadata.

## Clarifications

### Session 2026-06-29

- Q: How is NuGet publishing authenticated? → A: OIDC Trusted Publishing via `NuGet/login@v1`; `NUGET_USERNAME` secret only; no long-lived `NUGET_API_TOKEN` required.
- Q: Which language SDKs are smoke-tested in `verify-release.yml`? → A: Node.js and Python (both verified via package install + import).
