# Quickstart: Validate Multi-Language Provider Build

**Feature**: 004-multi-language-provider-build
**Date**: 2026-06-29

This guide documents how to validate that the build infrastructure works correctly end-to-end. These are the scenarios that prove the feature is complete.

---

## Prerequisites

- Docker and VS Code with the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) installed, **OR**
- `mise` installed locally (`curl https://mise.run | sh`)
- Go 1.25+ (if running outside devcontainer and without mise)

---

## Scenario 1: Devcontainer Environment (P1)

Validates FR-001, FR-002 (User Story 1).

```bash
# 1. Clone the repository
git clone https://github.com/deltastreaminc/pulumi-deltastream.git
cd pulumi-deltastream

# 2. Open in VS Code — it will prompt to "Reopen in Container"
code .
# OR: cmd/ctrl+shift+P → "Dev Containers: Reopen in Container"
```

**Expected**: Container builds successfully and VS Code opens in the container workspace.

```bash
# 3. In the container terminal, verify all tools are present:
go version           # Should show go1.25.11
node --version       # Should show v22.x (Node 22 LTS)
python3 --version    # Should show Python 3.11.15
dotnet --version     # Should show 8.0.x
java --version       # Should show OpenJDK 11 (Corretto)
pulumi version       # Should show 3.246.0 or later
golangci-lint version
yarn --version       # Should show 1.22.22
```

**Expected**: All tools available at pinned versions.

---

## Scenario 2: Local Build Without Docker (P1)

Validates FR-002 (mise as single source of truth).

```bash
# 1. Install mise if not already installed
curl https://mise.run | sh
# Add to shell: echo 'eval "$(mise activate bash)"' >> ~/.bashrc && source ~/.bashrc

# 2. In the repo root, install all tools
cd pulumi-deltastream
mise install

# 3. Verify all tools
go version && node --version && python3 --version && dotnet --version
java --version && pulumi version && golangci-lint version
```

**Expected**: All tools installed at versions matching `.config/mise.toml`.

---

## Scenario 3: Full Local Build (P1)

Validates FR-003, FR-004 (all five language SDKs generated and built).

```bash
# Requires: provider built, Pulumi CLI available
make build
```

**Expected**:
- `bin/pulumi-resource-deltastream` exists and is executable
- `schema.json` is generated at repo root
- `sdk/nodejs/` exists with TypeScript source + compiled `bin/`
- `sdk/python/` exists with Python source
- `sdk/dotnet/` exists with C# source
- `sdk/go/pulumi-deltastream/` exists with Go source
- `sdk/java/` exists with Java source

```bash
# Verify per-language:
ls sdk/nodejs/bin/index.js          # Node compiled
ls sdk/python/bin/                  # Python built
ls sdk/dotnet/bin/                  # dotnet built
ls sdk/go/pulumi-deltastream/       # Go generated
ls sdk/java/src/                    # Java generated
```

---

## Scenario 4: Cross-Compile Provider Binaries (FR-004)

```bash
make provider-linux-amd64
make provider-linux-arm64
make provider-darwin-amd64
make provider-darwin-arm64
```

**Expected**: `bin/<os>-<arch>/pulumi-resource-deltastream` exists for each platform.

```bash
# Package all four tarballs
make provider_dist PROVIDER_VERSION=1.0.0-test
ls bin/pulumi-resource-deltastream-v1.0.0-test-*.tar.gz
```

**Expected**: Four `.tar.gz` files, one per platform.

---

## Scenario 5: Linting (FR-005)

```bash
make lint
```

**Expected**: Exits 0 with no violations. If violations exist, output shows file:line:message.

```bash
# Auto-fix lint issues
make lint.fix
```

---

## Scenario 6: Provider Unit Tests (FR-006)

```bash
make test_provider
```

**Expected**: All unit tests pass, `provider/coverage.txt` generated. No DeltaStream credentials needed.

---

## Scenario 7: CI Pull Request Workflow (FR-007 through FR-011)

Open a pull request against `main`. Observe the GitHub Actions run.

**Expected jobs**: `prerequisites`, `build_sdk` (5 parallel language jobs), `lint`

**Verify**:
1. Each of the 5 language SDK build jobs runs independently
2. `prerequisites` job posts a schema diff comment on the PR (if `schema.json` changed)
3. If PR is from a fork, the `test` job is skipped
4. If PR is from the same repo, the `test` job runs after build + SDK jobs

---

## Scenario 8: Release Pipeline (FR-010 through FR-013)

Push a version tag to trigger a release. For testing use a prerelease tag:

```bash
git tag v1.0.0-rc.1
git push origin v1.0.0-rc.1
```

**Expected jobs**: `prerequisites` → `build_provider` (4 parallel platforms) + `build_sdk` (5 parallel languages) → `test` → `publish` → `verify`

**Verify after completion**:
```bash
# GitHub Release exists
gh release view v1.0.0-rc.1

# Provider tarballs present (4 platforms)
gh release view v1.0.0-rc.1 --json assets -q '.assets[].name'

# Node smoke test
npm install @deltastream/pulumi-deltastream@1.0.0-rc.1
node -e "require('@deltastream/pulumi-deltastream')"

# Python smoke test
pip install pulumi-deltastream==1.0.0-rc.1
python -c "import pulumi_deltastream"
```

---

## Scenario 9: Schema Metadata for Pulumi Registry (FR-017, FR-018)

Verify the schema contains all required registry metadata:

```bash
jq '{
  name,
  displayName,
  description,
  publisher,
  logoUrl,
  keywords,
  pluginDownloadURL
}' schema.json
```

**Expected output**:
```json
{
  "name": "deltastream",
  "displayName": "DeltaStream",
  "description": "A Pulumi native provider for DeltaStream...",
  "publisher": "DeltaStream Inc.",
  "logoUrl": "https://raw.githubusercontent.com/.../deltastream-logo.png",
  "keywords": ["pulumi", "deltastream", "category/database", "kind/native"],
  "pluginDownloadURL": "github://api.github.com/deltastreaminc"
}
```

Verify docs files exist:

```bash
ls docs/_index.md docs/installation-configuration.md
```

---

## Validation Checklist

| Scenario | FR(s) Validated | Pass Criteria |
|---|---|---|
| 1. Devcontainer environment | FR-001, FR-002 | All tools on PATH in container |
| 2. Local build with mise | FR-002 | `mise install` installs all tools |
| 3. Full local build | FR-003 | All 5 SDK directories exist and are compiled |
| 4. Cross-compile + tarballs | FR-004 | 4 platform tarballs produced |
| 5. Linting | FR-005 | `make lint` exits 0 |
| 6. Provider unit tests | FR-006 | Tests pass without credentials |
| 7. CI PR workflow | FR-007–FR-011, FR-015 | 7 jobs run, schema diff posted, fork test skipped |
| 8. Release pipeline | FR-010–FR-013 | 4 tarballs + 3 SDK registries + smoke tests |
| 9. Schema registry metadata | FR-014, FR-017, FR-018 | Schema has all required fields + docs exist |
