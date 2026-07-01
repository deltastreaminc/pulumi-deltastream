# Contract: GitHub Actions Workflows

**Feature**: 004-multi-language-provider-build
**Type**: CI/CD interface

This contract defines the workflow interfaces (inputs, outputs, required secrets, artifact names) that callers (`ci.yml`, `release.yml`) must satisfy when calling reusable workflows.

---

## `prerequisites.yml`

**Trigger**: `workflow_call`

### Inputs

| Name | Type | Required | Description |
|---|---|---|---|
| `is_pr` | boolean | yes | Whether this is a pull request run (enables schema diff comment) |
| `default_branch` | string | yes | Default branch name for schema diff comparison |

### Outputs

| Name | Type | Description |
|---|---|---|
| `version` | string | Computed semver version (e.g., `1.2.3`) — no leading `v` |

### Artifacts Produced

| Artifact Name | Contents |
|---|---|
| `provider-bin` | `pulumi-resource-deltastream` binary (linux-amd64) |
| `schema.json` | Generated provider schema |

### Required Secrets

`GITHUB_TOKEN` (for PR comments and schema diff via `gh` CLI)

---

## `build_sdk.yml`

**Trigger**: `workflow_call`

### Inputs

| Name | Type | Required | Description |
|---|---|---|---|
| `version` | string | yes | Provider version (no leading `v`) |

### Artifacts Consumed

| Artifact Name | Required By |
|---|---|
| `provider-bin` | All languages (binary placed in `bin/`) |
| `schema.json` | All languages (used by `pulumi package gen-sdk`) |

### Artifacts Produced

| Artifact Name | Contents |
|---|---|
| `sdk-nodejs` | Generated + compiled TypeScript SDK (`sdk/nodejs/`) |
| `sdk-python` | Generated + built Python SDK (`sdk/python/`) |
| `sdk-dotnet` | Generated + built .NET SDK (`sdk/dotnet/`) |
| `sdk-go` | Generated Go SDK (`sdk/go/`) |
| `sdk-java` | Generated + built Java SDK (`sdk/java/`) |

### Matrix

`language` ∈ `[nodejs, python, dotnet, go, java]` — runs in parallel.

---

## `build_provider.yml`

**Trigger**: `workflow_call`

### Inputs

| Name | Type | Required | Description |
|---|---|---|---|
| `version` | string | yes | Provider version (no leading `v`) |

### Artifacts Consumed

| Artifact Name | Required By |
|---|---|
| `schema.json` | Sentinel restoration only |

### Artifacts Produced

| Artifact Name | Contents |
|---|---|
| `pulumi-resource-deltastream-v<version>-linux-amd64.tar.gz` | Binary + README + LICENSE |
| `pulumi-resource-deltastream-v<version>-linux-arm64.tar.gz` | Binary + README + LICENSE |
| `pulumi-resource-deltastream-v<version>-darwin-amd64.tar.gz` | Binary + README + LICENSE |
| `pulumi-resource-deltastream-v<version>-darwin-arm64.tar.gz` | Binary + README + LICENSE |

### Matrix

`platform` ∈ `[{linux,amd64}, {linux,arm64}, {darwin,amd64}, {darwin,arm64}]` — all cross-compiled on `ubuntu-latest`.

---

## `lint.yml`

**Trigger**: `workflow_call` (no inputs)

Runs `make lint` against `provider/` Go code. Fails if any golangci-lint violations are found.

---

## `publish.yml`

**Trigger**: `workflow_call`

### Inputs

| Name | Type | Required | Description |
|---|---|---|---|
| `version` | string | yes | Provider version (no leading `v`) |
| `isPrerelease` | boolean | yes | Whether to create a draft/prerelease GitHub Release |
| `setLatestRelease` | boolean | yes | Whether to mark the GitHub Release as "latest" |
| `skipGoSdk` | boolean | no | Skip Go SDK publish (default: false) |

### Artifacts Consumed

| Pattern | Contents |
|---|---|
| `pulumi-resource-deltastream-v<version>-*.tar.gz` | Provider platform tarballs |
| `schema.json` | For checksums + GitHub Release attachment |
| `sdk-nodejs`, `sdk-python`, `sdk-dotnet`, `sdk-go`, `sdk-java` | Language SDKs for publish |

### Required Secrets

| Secret | Purpose |
|---|---|
| `GITHUB_TOKEN` | Create GitHub Release |
| `NPM_TOKEN` | Publish to npm |
| `PYPI_API_TOKEN` | Publish to PyPI |
| `NUGET_USERNAME` | nuget.org username for OIDC token exchange — no long-lived NuGet API key needed |

> **NuGet Trusted Publishing**: The `publish_sdk` job has `id-token: write` permission and calls `NuGet/login@v1` to exchange a GitHub OIDC token for a short-lived NuGet API key. No `NUGET_API_TOKEN` secret is needed. A one-time Trusted Publishing policy must be configured on nuget.org pointing to `publish.yml`.

### Side Effects

1. Creates GitHub Release at `v<version>` with tarballs, checksums, schema
2. Publishes npm package `@deltastream/pulumi-deltastream@<version>`
3. Publishes PyPI package `pulumi-deltastream==<version>`
4. Publishes NuGet package `Pulumi.DeltaStream <version>`
5. Tags Go SDK commit via `pulumi/publish-go-sdk-action`

---

## `verify-release.yml`

**Trigger**: `workflow_call` or `workflow_dispatch`

### Inputs

| Name | Type | Required | Description |
|---|---|---|---|
| `version` | string | yes | Provider version to verify (no leading `v`) |

### Verification Steps

| Runtime | Check |
|---|---|
| nodejs | `npm install @deltastream/pulumi-deltastream@<version>` + `node -e "require(...)"` |
| python | `pip install pulumi-deltastream==<version>` + `python -c "import pulumi_deltastream"` |

---

## Artifact Retention

All CI artifacts are retained for **7 days**. This is sufficient for the publish job (same pipeline) and local debugging within a sprint.

---

## Tarball Naming Convention

Provider tarballs must follow this exact format for the Pulumi CLI plugin system to locate them:

```
pulumi-resource-<provider>-v<semver>-<os>-<arch>.tar.gz
```

Example: `pulumi-resource-deltastream-v1.2.0-linux-amd64.tar.gz`

The `pluginDownloadURL` in `schema.json` (`github://api.github.com/deltastreaminc`) instructs the Pulumi CLI to look for these tarballs in the GitHub Releases of the `deltastreaminc` organization's repository matching the provider name.
