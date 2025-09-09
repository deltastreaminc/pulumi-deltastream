# Feature Specification: Build CI and Release GitHub Actions

**Feature Branch**: `003-build-a-ci`  
**Created**: September 7, 2025  
**Status**: Draft  
**Input**: User description: "Build a CI and a Release github action which can be used to test PR and release the provider from main. It should be built in a manner that does not leak github secrets to forked repos."

## User Scenarios & Testing *(mandatory)*

### Primary User Story
As a contributor to the Pulumi Deltastream provider, I want automated CI/CD workflows that ensure code quality on pull requests and facilitate reliable releases, so that I can confidently develop and deploy changes without worrying about manual testing or release processes.

### Acceptance Scenarios
1. **Given** a contributor creates a pull request, **When** the CI workflow runs, **Then** all tests are executed and results are reported on the PR without exposing secrets.
2. **Given** a maintainer merges code to the main branch, **When** the Release workflow is triggered, **Then** all tests are executed and results are recorded against the commit.
3. **Given** a maintainer tags a revision on the main branch, **When** the Release workflow runs, **Then** versioned provider tarballs, checksum file, schema, and supported SDKs (Node.js, Go) are produced and published.
4. **Given** a fork submits a pull request, **When** the CI workflow runs, **Then** tests do not run automatically but can be triggered by repo members.
5. **Given** a release completes, **When** post-publish verification runs, **Then** minimal programs for Node.js and Go succeed using the just-published version.

### Edge Cases
- What happens when tests fail during a CI run?
	- The pull request status shows failure. The contributor can fix the issues and submit a new commit or PR update to rerun the tests.
- How does the system handle version conflicts during release?
	- The release version is based on the git tag of the commit being released, ensuring no version conflicts.
- What happens if a release workflow is interrupted or fails midway?
	- It is the responsibility of the repository owner or members to retrigger the workflow manually.
- How are secrets managed when a PR comes from a fork?
	- Tests that require secrets only run if explicitly triggered by a repository member; otherwise, they are skipped for forked PRs.
- How is version management handled for automated releases?
	- The release version is determined by the git tag associated with the commit being released.

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: System MUST implement a CI workflow that runs on all pull requests to verify code quality and functionality.
- **FR-002**: System MUST implement a Release workflow that creates and publishes new releases of the provider.
- **FR-003**: CI workflow MUST prevent exposure of repository secrets to forked repositories.
- **FR-004**: CI workflow MUST run unit tests, integration tests, and code quality checks.
- **FR-005**: CI workflow MUST report test results back to the pull request.
**FR-006**: Release workflow MUST generate artifacts for linux amd64, linux arm64, and darwin arm64 (matching pulumi/pulumi-aws platform set sans signed binaries).
**FR-007**: Release workflow MUST package provider binaries as tarballs named `pulumi-resource-deltastream-v<version>-<os>-<arch>.tar.gz` and produce a consolidated SHA256 checksum file (align with pulumi/pulumi-aws format).
**FR-008**: Release workflow MUST attach `schema.json` and include a schema diff summary in release notes (skip diff if first release or prerelease).
**FR-009**: Release workflow MUST publish the Node.js SDK via `pulumi/pulumi-package-publisher@v0.0.22` (with `sdk: nodejs`).
**FR-010**: Release workflow MUST publish the Go SDK via `pulumi/publish-go-sdk-action@v1` after Node publish succeeds.
**FR-011**: Release workflow MUST support prerelease (draft) handling via `isPrerelease` flag and treat semantic versions containing a hyphen as prereleases automatically.
**FR-012**: Release workflow MUST run a post-publish verification job executing smoke tests for Node.js and Go using the just-published versions.
**FR-013**: Version normalization MUST strip leading `v` for package metadata while preserving `v` for git tags (Node.js & Go fetch logic).
**FR-014**: Both workflows MUST follow GitHub Actions security best practices to protect sensitive information (fork PR secret isolation, least-privilege tokens).
**FR-015**: Release workflow MUST remove macOS code signing steps present in earlier iteration and produce unsigned darwin arm64 binaries (mirrors current pulumi/pulumi-aws approach for unsigned community providers).
**FR-016**: Release workflow SHOULD minimize bespoke logic by delegating Node packaging/publish to the package publisher action (single step replacing manual version bump + yarn publish) while retaining local tarball & checksum generation for provider distribution.
**FR-017**: Release workflow MUST fail fast if the package publisher action reports an error (subsequent Go publish and release creation MUST not proceed).
- **FR-018**: Documentation MUST indicate that adding future languages (.NET, Python, Java) would extend the `sdk:` parameter of the publisher action, not reintroduce bespoke publish steps.
	- Node: Published via package publisher composite action
	- Go: Git tag already exists (trigger) and curated Go SDK commit via publish-go-sdk action
	- Future languages: Extend publisher action `sdk:` input
	- Platforms: linux amd64, linux arm64, darwin arm64 (unsigned)
- **GitHub Workflow**: Definition of automated processes that run on GitHub events (PR creation, merge to main, etc.)
- **GitHub Secret**: Sensitive information stored securely in the repository settings, used by workflows.
- **Release Artifact**: Built packages, binaries, or other files that are published as part of a release.

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [ ] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous  
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified
- [x] Deltastream system tables usage documented (if applicable)

---

## Execution Status
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [ ] Review checklist passed

---
*Based on Constitution v2.2.0 - See `/memory/constitution.md`*
