# AGENTS.md - Coding Assistant Instructions

## Project Overview
This Pulumi provider implements resources to orchestrate the DeltaStream cloud platform. It will expose resources for managing DeltaStream databases, pipelines, and other platform components. We have two important reference codebases:

1. `github.com:deltastreaminc/terraform-provider-deltastream` - An existing Terraform provider that can be used as reference for resource design and API interactions
2. `github.com:deltastreaminc/go-deltastream` - The Go SDK that will be used for making calls to the DeltaStream backend API

## Reference Documentation
- DeltaStream SQL Syntax: https://docs.deltastream.io/reference/sql-syntax

## Build Commands
- Build provider: `make build`
- Test provider: `make test_provider` (full suite)
- Test specific: `cd provider && go test -v -count=1 ./... -run TestName`
- Lint code: `make lint`
- Generate schema: `make schema`
- Generate SDKs: `make build_sdks` (creates all language SDKs)
- Clean build artifacts: `make clean`
- Install locally: `make install` (copies binary to GOPATH/bin)

## Pulumi Provider Best Practices
- Define clear resource interfaces with Create/Read/Update/Delete methods
- Use `infer.Annotator` to document all resources and properties
- Make operations idempotent where possible
- Handle transient errors gracefully with appropriate retries
- Ensure detailed error messages for troubleshooting
- Implement proper state diffing logic to detect changes
- Make resource deletions safe with proper validation
- Use `pulumi:",optional"` tag for optional parameters
- Include thorough documentation in schema
- Test with `pulumi-go-provider` testing framework

## Integration Testing
	- `DELTASTREAM_API_KEY`, `DELTASTREAM_SERVER` (required)
	- Optional: `DELTASTREAM_ORGANIZATION`, `DELTASTREAM_ROLE`, `DELTASTREAM_INSECURE_SKIP_VERIFY`, `DELTASTREAM_SESSION_ID`.
# Agents

This document captures requirements and decisions for the DeltaStream Pulumi provider repo.

Requirements
- Provider config: `server` and `apiKey` required; optional `organization`, `role`, `insecureSkipVerify`, `sessionId`.
- Database resource with real SQL semantics; `getDatabase` and `getDatabases` invokes.
- Examples in TypeScript, and Go.
- Integration tests follow Pulumi official style and run examples end-to-end.
- Test credentials are read from `examples/credentials.yaml` with fields: `apiKey`, `server`, optional `organization`, `role`, `insecureSkipVerify`, `sessionId`.
- Node SDK package name: `@deltastream/pulumi-deltastream`.
- CI: PRs build provider, generate SDKs, run unit and example checks, and (optionally) integration tests if credentials are provided. Releases publish the Node package to npm.

Decisions
- Use YAML for test credentials to allow richer optional fields.
- Integration tests copy local SDKs into temp project directories and rewrite dependencies to local paths, avoiding external registry reliance.
- The Go example uses a replace directive injected by tests to point at the local Go SDK.

How to run integration tests locally
1. Create `examples/credentials.yaml` using `examples/credentials.yaml.example`.
2. Build provider and SDKs:
   - `make build`
   - `make codegen`
3. Run: `make test_integration`

Release process
- Tag the repository with `vX.Y.Z` to trigger the npm publish workflow for the Node SDK. Ensure `NPM_TOKEN` is configured in repo secrets.

## Package Naming
- Node SDK package name: `@deltastream/pulumi-deltastream`.
- Go SDK import path: `github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream`.

## CI and Releases
- GitHub Actions workflow `.github/workflows/ci.yml` runs on PRs: builds provider and SDKs, compiles examples, and runs integration tests.
- Release job publishes the Node SDK to npm under `@deltastream/pulumi-deltastream` on version tags (`vX.Y.Z`). Requires `NPM_TOKEN` secret.
- Provider binaries and other SDK releases can be added later.

## Code Style Guidelines
- Use Go 1.24+
- Follow idiomatic Go conventions.
- Write clear and concise comments for exported functions and types.
- Ensure code is formatted using `gofmt`.
- Import ordering: standard library, third-party, internal with new line between each section
- Error handling: use `fmt.Errorf("error message: %w", err)` for wrapping
- Type definitions: place at top of file, followed by constructors
- Comments: all exported functions/types need comments
- Resource naming: PascalCase singular nouns (e.g., Database not Databases)
- Use pointers for optional parameters with `*string` type
- Avoid custom logging, use the provider logger via `p.GetLogger(ctx)`
- Keep resource properties consistent with DeltaStream API naming
- Include copyright notice at the top of each file: "Copyright <year>, DeltaStream Inc."
- Function definitions should be on a single line (e.g., `func (r *Resource) Create(ctx context.Context, name string, input Resource, preview bool) (id string, output Resource, err error) {`)
- Add a single empty line after control statement blocks (if, for, switch, etc.)

## Specifications & Memory (Constitution System)

This repository uses a spec-kit aligned template (`specs/_template.md`) and a project constitution (`memory/constitution.md`) to provide durable context.

### Hierarchy of Truth (Agents & Contributors)
1. `memory/constitution.md` (principles, guardrails, decision clauses)
2. Accepted specs in `specs/` (Status: Accepted)
3. Draft specs in `specs/` (Status: Draft/Review)
4. Open issues / discussions
5. Source code heuristics

Agents MUST consult sources in the above order when resolving ambiguities.

### Spec Lifecycle
1. Author draft from `_template.md` (remove irrelevant optional sections)
2. Fill Goals, Summary, User Scenarios, Functional & Non-Functional requirements
3. Mark unknowns with `[NEEDS CLARIFICATION: ...]`
4. PR review removes / resolves all markers
5. On acceptance: update Status to `Accepted` & link from constitution index (future index section)

### Requirements Formatting
- Functional: `FR-###` sequential, testable, no implementation details
- Non-Functional: `NFR-###` sequential, measurable where possible
- Do not reuse numbers after removal; mark deprecated if withdrawn

### Ambiguity Handling
If an agent encounters behavior not covered by constitution or accepted spec it MUST:
1. Add a `[NEEDS CLARIFICATION: ...]` marker in the relevant draft spec OR
2. Propose a new clause in a PR referencing source context

### Decision Recording
- Each material change to principles or process should add or modify a clause in `memory/constitution.md`.
- Clauses should be atomic and reference an issue/PR for traceability.
- Superseded clauses retain id with note: `Superseded by CLAUSE-XYZ (PR #NNN)`.

### Agent Behavior Requirements
- MUST never invent requirements absent from hierarchy.
- SHOULD cite the FR/NFR or Clause ID when performing non-trivial changes.
- MUST avoid embedding secrets/reference tokens in specs or constitution.

### Lint / Future Automation (Planned)
- CI validation for unresolved `[NEEDS CLARIFICATION]` in Accepted specs.
- Optional script to ensure sequential FR/NFR numbering.

### Getting Started (Spec Author Quick Steps)
```
cp specs/_template.md specs/<feature>.md
git add specs/<feature>.md
# Edit mandatory sections, add FR/NFR lists, push PR.
```

### Rationale
Centralizing durable design intent decreases drift, accelerates onboarding, and enables safer automation by providing an explicit contract for behavior and scope.
