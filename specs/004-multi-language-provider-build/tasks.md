---
description: "Tasks for feature 004-multi-language-provider-build"
---

# Tasks: Multi-Language Provider Build & Publish

**Input**: Design documents from `specs/004-multi-language-provider-build/`

**Context**: All source files were created during the spec/plan phase. These tasks validate correctness, fix the critical schema metadata issue, and complete the remaining setup steps (credentials, registry submission, logo) required to make the feature production-ready.

**Tests**: Not applicable — this feature IS the CI/build infrastructure. Validation is performed by running make targets and triggering CI workflows.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel with other tasks in the same phase
- **[Story]**: User story this task belongs to (US1–US4)

---

## Phase 1: Critical Fix — Schema Metadata Preservation

**Purpose**: Fix the one critical issue found during planning (plan.md T-001): the `make schema` jq pipeline must inject all registry and language metadata so it survives schema regeneration.

**⚠️ CRITICAL**: Must be complete before any CI run or `make schema` invocation, otherwise all registry metadata will be overwritten.

**Status**: ✅ Already applied during the planning phase. Verify correctness below.

- [x] T001 Verify `make schema` jq pipeline in `Makefile` (line `.make/schema` target) injects all 12 metadata fields: `displayName`, `description`, `publisher`, `logoUrl`, `keywords`, `language.go.*`, `language.nodejs.packageName`, `language.python.packageName`, `language.python.pyproject`, `language.csharp.packageName`, `language.java.basePackage`, `language.java.buildFiles`
- [x] T002 Run `python3 -c "import json; d=json.load(open('schema.json')); assert d.get('publisher') == 'DeltaStream Inc.', f'Missing: {d.get(\"publisher\")}'; print('schema.json metadata OK')"` at repo root to confirm current `schema.json` has the metadata (note: regeneration requires `pulumi` CLI + built provider)

---

## Phase 2: Foundational — Local Environment Validation (Blocking)

**Purpose**: Confirm the mise + devcontainer infrastructure works end-to-end on the local machine before CI validation.

**⚠️ CRITICAL**: No CI push should happen until local build succeeds.

- [x] T003 Install mise if not present: `curl https://mise.run | sh` and activate in shell
- [x] T004 Run `mise install` in repo root; verify all tools listed in `.config/mise.toml` are installed at pinned versions (go 1.25.11, node 22.23.1, python 3.11.15, dotnet 8.0.20, java corretto-11, pulumi 3.246.0, yarn 1.22.22, golangci-lint 2.12.2)
- [x] T005 Run `make provider` to build the provider binary; confirm `bin/pulumi-resource-deltastream` exists and is executable
- [x] T006 Run `make schema` to regenerate `schema.json`; confirm it contains `publisher`, `logoUrl`, `keywords`, and all five language blocks using: `jq '{publisher,logoUrl,keywords,language:.language|keys}' schema.json`
- [x] T007 Run `make test_provider` to confirm provider unit tests pass without credentials

**Checkpoint**: Local environment is healthy — CI workflows can now be validated.

---

## Phase 3: User Story 1 — Developer Gets Working Environment Instantly (P1) 🎯 MVP

**Goal**: A contributor can clone, open in devcontainer or run `mise install`, and immediately build all five language SDKs.

**Independent Test**: `make build` completes successfully with all five SDK source directories populated.

### Implementation for User Story 1

- [x] T008 [US1] Run `make build` end-to-end; confirm `sdk/nodejs/`, `sdk/python/`, `sdk/dotnet/`, `sdk/go/pulumi-deltastream/`, and `sdk/java/` all exist after completion
- [x] T009 [P] [US1] Verify `sdk/nodejs/bin/index.js` exists (TypeScript compiled): `ls sdk/nodejs/bin/index.js`
- [x] T010 [P] [US1] Verify Python SDK built: `ls sdk/python/bin/` should contain wheel/sdist artifacts
- [x] T011 [P] [US1] Verify .NET SDK built: `ls sdk/dotnet/bin/` should contain `.nupkg` file
- [x] T012 [P] [US1] Verify Go SDK generated: `ls sdk/go/pulumi-deltastream/go.mod`
- [x] T013 [P] [US1] Verify Java SDK generated: `ls sdk/java/src/`
- [x] T014 [US1] Validate devcontainer builds successfully by opening the repo in VS Code Dev Containers (or running `docker build -f .devcontainer/Dockerfile .`) and confirming all tools are on PATH inside the container
- [x] T015 [P] [US1] Add `docs/deltastream-logo.png` (or SVG) to `docs/` directory — required for `schema.json` `logoUrl` reference; without it the Pulumi Registry listing will show a broken image. Source a DeltaStream logo SVG with whitespace removed per Registry guidelines.

**Checkpoint**: `make build` passes end-to-end; devcontainer is functional; logo file exists.

---

## Phase 4: User Story 2 — Pull Request Build Validates All Languages (P2)

**Goal**: Opening a PR triggers the full CI pipeline: prerequisites, 5-language SDK matrix build, lint, and fork-safe test gating.

**Independent Test**: Push a branch with a trivial change and open a PR — all 7 CI jobs must pass (prerequisites, build_sdk×5, lint), and a schema diff comment must appear on the PR.

### Implementation for User Story 2

- [x] T016 [US2] Commit all new and modified files to a feature branch and push to GitHub: `git add .config/ .devcontainer/ scripts/ .golangci.yml mise.toml Makefile schema.json .gitignore .github/workflows/ docs/ specs/004-multi-language-provider-build/ && git commit -m "feat: add mise devcontainer multi-language build CI"`
- [x] T017 [US2] Open a pull request against `main` and observe the CI workflow run; verify jobs `prerequisites`, `build_sdk (nodejs)`, `build_sdk (python)`, `build_sdk (dotnet)`, `build_sdk (go)`, `build_sdk (java)`, and `lint` all appear
- [x] T018 [US2] Confirm the `prerequisites` job completes: provider binary built, schema generated, `provider-bin` and `schema.json` artifacts uploaded
- [x] T019 [US2] Confirm all 5 `build_sdk` matrix jobs pass independently (check each language tab in Actions)
- [x] T020 [US2] Confirm the `lint` job passes; if it fails, run `make lint` locally, fix violations with `make lint.fix`, and push fixes
- [ ] T021 [US2] Confirm the `test` job is skipped if the PR is from a fork (validate FR-015); if testing from the same repo, confirm test job runs after build_sdk completes
- [x] T022 [US2] Confirm the schema diff comment appears on the PR (may show "No previous release found" on first run — that is correct behaviour)
- [x] T023 [P] [US2] Fix any `build_sdk (dotnet)` failures: check that `sdk/dotnet/` is generated correctly and `dotnet build` succeeds; common issue is missing `version.txt` (the Makefile writes it at generate time)
- [x] T024 [P] [US2] Fix any `build_sdk (java)` failures: verify Java SDK generation succeeds with `make generate_java` locally; the Java generator may require `pulumi plugin install` — document any additional setup in `README.md` if needed (plan.md T-005)
- [x] T025 [US2] Verify `pulumi/provider-version-action@v2` produces a version string without a leading `v` (the Makefile enforces this with an error guard); check the `prerequisites` job output for the computed version (plan.md T-006)

**Checkpoint**: PR CI passes all 7 jobs; fork gating works; schema diff comment posted.

---

## Phase 5: User Story 3 — Release Publishes All Language Packages (P2)

**Goal**: Pushing a version tag triggers the full release pipeline: cross-compile 4 platforms, build 5 language SDKs, create GitHub Release, publish npm + PyPI + NuGet + Go SDK, run smoke tests.

**Independent Test**: Push `v1.0.0-rc.1` tag — a draft GitHub Release appears with 4 platform tarballs, and npm/PyPI/NuGet receive the package.

### Setup: Configure Publishing Credentials

- [ ] T026 [US3] Create a PyPI API token at https://pypi.org/manage/account/token/ scoped to project `pulumi-deltastream` and add it as GitHub repository secret `PYPI_API_TOKEN` (Settings → Secrets and variables → Actions)
- [ ] T027 [US3] Configure NuGet Trusted Publishing on nuget.org (Account → Trusted Publishing → Add policy): Repository Owner: `deltastreaminc`, Repository: `pulumi-deltastream`, Workflow File: `publish.yml`, Environment: *(empty)*; then add GitHub secret `NUGET_USERNAME` set to the nuget.org profile name (not email)
- [ ] T028 [US3] Verify `NPM_TOKEN` repository secret is present and valid (already existed; confirm it has publish scope for `@deltastream/pulumi-deltastream`)

### Implementation for User Story 3

- [ ] T029 [US3] Push a prerelease tag to trigger the release workflow: `git tag v1.0.0-rc.1 && git push origin v1.0.0-rc.1`
- [ ] T030 [US3] Observe `release.yml` workflow in GitHub Actions; verify jobs `prerequisites`, `build_provider` (4 platform matrix), `build_sdk` (5 language matrix), `lint`, `test`, `publish`, and `verify` all appear
- [ ] T031 [US3] Confirm all 4 `build_provider` matrix jobs produce tarballs: `pulumi-resource-deltastream-v1.0.0-rc.1-linux-amd64.tar.gz`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`
- [ ] T032 [US3] Confirm `publish` job creates a draft GitHub Release with all 4 tarballs, checksums file, and `schema.json` attached
- [ ] T033 [US3] Confirm `publish_sdk` job publishes `@deltastream/pulumi-deltastream@1.0.0-rc.1` to npm: `npm view @deltastream/pulumi-deltastream@1.0.0-rc.1 version`
- [ ] T034 [US3] Confirm `publish_sdk` job publishes `pulumi-deltastream==1.0.0-rc.1` to PyPI: `pip index versions pulumi-deltastream`
- [ ] T035 [US3] Confirm `publish_sdk` job publishes `DeltaStream.Pulumi 1.0.0-rc.1` to NuGet via OIDC Trusted Publishing (no API key stored)
- [ ] T036 [US3] Confirm `verify` job smoke tests pass: Node.js `require('@deltastream/pulumi-deltastream')` and Python `import pulumi_deltastream` both succeed
- [ ] T037 [US3] Confirm Go SDK is tagged by `pulumi/publish-go-sdk-action`: check that a `sdk/v1.0.0-rc.1` tag exists on the repository after the publish job

**Checkpoint**: Full release pipeline completes; all 4 platform binaries + 3 SDK registries (npm, PyPI, NuGet) receive the prerelease version; Go SDK tagged.

---

## Phase 6: User Story 4 — Provider Discoverable in Pulumi Registry (P3)

**Goal**: Submit a PR to `pulumi/registry` so the DeltaStream provider appears at `pulumi.com/registry/packages/deltastream`.

**Independent Test**: PR to `pulumi/registry` is mergeable and passes the registry's CI checks.

### Implementation for User Story 4

- [x] T038 [US4] Verify `schema.json` contains all required Pulumi Registry metadata: `publisher`, `logoUrl` (pointing to an existing file), `keywords` with `category/database` and `kind/native`, `description`
- [x] T039 [US4] Verify `docs/deltastream-logo.png` (or `.svg`) exists at the path referenced in `schema.json` `logoUrl` (created in T015); if not, complete T015 first
- [x] T040 [US4] Review `docs/_index.md` for completeness: title matches `displayName` from `schema.json`, examples are accurate, resources list is current
- [x] T041 [US4] Review `docs/installation-configuration.md` for completeness: all config options match `schema.json` config variables, install commands are accurate for all 4 languages
- [ ] T042 [US4] Fork `github.com/pulumi/registry` and clone locally
- [ ] T043 [US4] Add entry to `community-packages/package-list.json` in the cloned `pulumi/registry` fork: `{ "repoSlug": "deltastreaminc/pulumi-deltastream", "schemaFile": "provider/cmd/pulumi-resource-deltastream/schema.json" }`
- [ ] T044 [US4] Copy `docs/_index.md` and `docs/installation-configuration.md` to `themes/default/content/registry/packages/deltastream/` in the `pulumi/registry` fork
- [ ] T045 [US4] Open a PR against `pulumi/registry` main branch with the above changes; reference the community package submission guide and a working example (e.g., `pulumi/registry#10358`)

**Checkpoint**: PR opened against `pulumi/registry`; registry CI checks pass.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup and documentation after all user stories are validated.

- [ ] T046 [P] Update `README.md` to document the new developer workflow: `mise install`, `make build`, `make test_provider`, devcontainer usage — add a "Development" section if not present
- [ ] T047 [P] Update `README.md` to list all five language package names and install commands so users can find the correct package for their language
- [ ] T048 Run the full quickstart validation checklist from `specs/004-multi-language-provider-build/quickstart.md` scenarios 1–9 and confirm all pass
- [x] T049 [P] Remove any `sdk/go/pulumi-deltastream/` files that should be regenerated (if the Go SDK was committed at an old version); run `make generate_go` and commit the refreshed SDK
- [ ] T050 Update `specs/004-multi-language-provider-build/spec.md` status from `Draft` to `Complete`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Critical Fix)**: No dependencies — verify immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — must pass before any CI push
- **Phase 3 (US1 — Local Build)**: Depends on Phase 2 — validates local environment
- **Phase 4 (US2 — CI)**: Depends on Phase 3 — requires a working local build to commit
- **Phase 5 (US3 — Release)**: Depends on Phase 4 (CI must pass on main before tagging) + T026–T028 (credentials)
- **Phase 6 (US4 — Registry)**: Depends on Phase 5 (first release should exist before registry submission)
- **Phase 7 (Polish)**: Depends on Phase 3 completed; can run alongside Phases 4–6

### User Story Dependencies

- **US1 (P1)**: No story dependencies — start immediately after Phase 2
- **US2 (P2)**: Depends on US1 completing successfully (must have a working build to commit)
- **US3 (P2)**: Depends on US2 (CI must pass before tagging a release); also depends on T026–T028
- **US4 (P3)**: Depends on US3 (first release gives a concrete version for the registry listing)

### Within Each Phase: Parallel Opportunities

**Phase 2**: T003 and T004 are sequential (install then verify); T005–T007 can run in parallel once mise is installed.

**Phase 3**: T009–T013 are all parallel (different SDK directories); T008 must precede them.

**Phase 4**: T023 and T024 (SDK build fixes) can run in parallel; T016 must precede all.

**Phase 5**: T026–T028 (credential setup) can run in parallel; T029 must wait for all three.

**Phase 6**: T038–T041 (review/verify) can run in parallel; T042–T045 are sequential.

---

## Parallel Execution Example: Phase 3 (US1)

```bash
# After T008 (make build) completes:
# Launch all SDK verification tasks in parallel:
Task: T009 — verify sdk/nodejs/bin/index.js
Task: T010 — verify sdk/python/bin/
Task: T011 — verify sdk/dotnet/bin/*.nupkg
Task: T012 — verify sdk/go/pulumi-deltastream/go.mod
Task: T013 — verify sdk/java/src/

# T015 (logo) runs independently of SDK verification
Task: T015 — add docs/deltastream-logo.png
```

---

## Implementation Strategy

### MVP Scope (US1 — Working Local Build)

1. Complete Phase 1 (verify schema fix)
2. Complete Phase 2 (local mise install + provider build)
3. Complete Phase 3 (all five SDKs build locally + devcontainer + logo)
4. **STOP and VALIDATE**: `make build` passes end-to-end; devcontainer confirmed working
5. Commit — this is a shippable, useful improvement even before CI is validated

### Full Feature Delivery

1. MVP (above) → Commit to feature branch
2. Phase 4 (US2) → Push branch, open PR, fix any CI failures → Merge to main
3. Phase 5 (US3) → Configure secrets, push a prerelease tag, verify release pipeline
4. Phase 6 (US4) → Submit Pulumi Registry PR after first real release
5. Phase 7 (Polish) → Update README, run full quickstart validation

---

## Notes

- `[P]` tasks operate on different files or are independently executable — safe to run in parallel
- This feature's implementation files are already created; these tasks are validation + remaining setup steps
- plan.md T-001 (schema jq fix) was applied during planning — T001/T002 verify it is correct
- plan.md T-005 (Java SDK generation) and T-006 (provider-version-action compatibility) are validated by T024 and T025 respectively
- All `make` targets assume `mise install` has been run (or mise is active in shell)
- Commit after each phase checkpoint to preserve working state
