# Feature Specification: Platform, CI/CD & Security

**Feature Branch**: `feat/platform-ci-security`  
**Created**: 2025-09-04  
**Status**: Accepted  
**Input / Source**: Operational hardening + release automation requirements

---
## ⚡ Quick Guidelines
Secure reproducible releases, fast PR feedback (<5m), fork-safe secrets, one-tag one-publish, signed + checksummed artifacts, enforced SQL safety.

---
## 1. Summary *(mandatory)*
Defines cross-cutting platform controls: CI workflows, release automation, security hardening, artifact integrity, and gating policies for the Pulumi provider & SDK publishing.

## 2. Goals *(mandatory)*
- Produce deterministic multi-platform plugin binaries and language SDKs.
- Enforce secret isolation for untrusted contexts (forks).
- Provide verifiable integrity (checksums, signing).
- Enable maintainers to trigger extended tests on demand.
- Prevent duplicate or partial publishes.

## 3. Non-Goals *(mandatory)*
- Higher SLSA level formal compliance claims (future upgrade path beyond minimal attestations).
- Gated (blocking) SBOM policy (advisory only in initial phase).
- Multi-provider monorepo orchestration (single provider scope only).

## 4. Background & Context *(mandatory)*
Pulumi providers require a reliable pipeline to compile native binaries for multiple OS/architectures and generate language SDKs. Security posture must prevent injection attacks (SQL) and secret leakage on forks while preserving rapid iteration speed.

## 5. Users & Personas *(mandatory)*
| Persona | Primary Need | Pain Today |
|---------|--------------|------------|
| Maintainer | Fast safe release iteration | Manual build steps, risk of missed platform |
| Security Reviewer | Secret containment & audit | Broad secret exposure risk |
| End User | Trustworthy artifacts | Hard to verify integrity |
| Contributor (Fork) | Feedback w/out secrets | Slow or blocked workflows |

## 6. User Scenarios & Testing *(mandatory)*
### Scenarios
1. Tag Release: On pushing `v1.2.3`, workflow builds all binaries, generates SDKs, publishes npm package once, attaches checksums.
2. Fork PR: External PR runs fast Linux-only checks without accessing secrets.
3. Maintainer Extended Tests: Maintainer comments `/run tests` → extended matrix (darwin/windows + arm64) executes.
4. Re-run Safety: Re-running a failed release job after artifact build does not republish npm if version exists.
5. Integrity Verification: User downloads binary, verifies SHA256 matches published checksum.

### Edge Cases
- Tag exists but npm already published → publish step is a no-op, workflow succeeds.
- Comment trigger from non-maintainer ignored safely.
- macOS signing secret missing → workflow marks artifact unsigned and fails release stage (policy choose fail fast).
- Partial matrix failure aborts publish & marks build incomplete.

## 7. Requirements *(mandatory)*
### Functional Requirements
- **FR-001**: Build plugin binaries for linux/darwin/windows across amd64 & arm64.
- **FR-002**: Publish Node SDK to npm exactly once on semantic version tag push.
- **FR-003**: Generate and attach SHA256 checksums for every released artifact.
- **FR-004**: Support maintainer `/run tests` comment to trigger extended test matrix.
- **FR-005**: Sign macOS binaries (codesign); notarization optional future.
- **FR-006**: Enforce fork gating (no secrets, minimal matrix) for untrusted PRs.
- **FR-007**: Skip publish if version already exists (idempotency guard).
- **FR-008**: Provide integrity verification instructions (checksums + signing) in release assets/readme.
- **FR-009**: Lint / build must fail if SQL quoting utilities are absent from changed SQL construction code paths.
- **FR-010**: Generate multi-language SDKs (Go, Node, Python) during release.

### Non-Functional Requirements
- **NFR-001**: Secrets MUST never be accessible in fork-origin workflows.
- **NFR-002**: Release workflow re-run MUST produce identical artifact hashes (deterministic build flags).
- **NFR-003**: PR fast path median completion time < 5 minutes.
- **NFR-004**: Signed macOS binaries MUST pass local `codesign --verify`.
- **NFR-005**: All artifact names and hashes MUST be logged in a single checksum summary step.
- **NFR-006**: Extended test matrix concurrency SHOULD not increase total time > 2x baseline.
- **NFR-007**: Failure in any build leg MUST prevent publish stage.

### Key Entities
- Release Artifact (os, arch, checksum, signature)
- Workflow Run (trigger type, matrix scope, secret exposure flag)

## 8. High-Level Design *(optional if trivial)*
Two workflows: `ci.yml` (fast path + conditional extended) and `release.yml` (tag triggered). Conditional steps gate secrets & publish. Deterministic build via fixed toolchain versions and `-trimpath`.

## 9. Detailed Notes *(optional)*
Idempotent npm publish implemented by querying registry before publish. Comment trigger filtered by actor membership (maintainer team or repo write permission).

## 10. Security & Privacy *(mandatory)*
Secret usage restricted to release & maintainer contexts. No PII stored. SQL safety enforced by centralized quoting utilities; lint ensures compliance.

## 11. Performance & Scalability *(optional)*
Matrix size small (≤6 combos). Caching Go module & npm dependencies reduces rebuild time. No horizontal scaling concerns.

## 12. Observability *(optional)*
Checksum manifest and signing verification logs included. Potential future metrics: release duration, PR latency.

## 13. Dependencies & Assumptions *(mandatory)*
- GitHub Actions availability & concurrency not heavily rate-limited.
- Maintainers maintain semantic version tagging discipline.
- Node/npm credentials stored as repository secret.

## 14. Risks & Mitigations *(mandatory)*
| Risk | Impact | Mitigation |
|------|--------|------------|
| Fork secret exposure | Credential leak | Strict `if` conditions, no secret steps on forks (FR-006) |
| Duplicate publish | Version conflict | Registry existence check (FR-007) |
| Unsigned macOS binary | User trust loss | Signing step failure blocks publish (FR-005) |
| SQL injection | Data compromise | Quoting utilities + lint (FR-009) |
| Partial matrix success | Inconsistent release | Require full matrix success (NFR-007) |

## 15. Alternatives Considered *(optional)*
- Single unified workflow (less separation; harder to reason about triggers).
- Manual release scripts (higher toil, error prone).

## 16. Decisions & Remaining Open Questions
### D-PLAT-001: Provenance Strategy (Minimal Attestations, No Immediate SLSA Level Claim)
Adopt minimal provenance attestations for release artifacts without formally claiming a SLSA level in the initial phase. Each release attaches a generic in-toto/DSSE signed provenance statement (tool: GitHub Actions attestation or future SLSA generator) covering: source repo, commit SHA, build workflow path, artifact digests, build timestamp. No hermeticity or dependency pin completeness is asserted yet. Public communication: "Provenance attestations available" (avoid overstating maturity). Upgrade path: evaluate SLSA Level 2 prerequisites (build service assurances, dependency capture) by 2025-11-30; target decision D-PLAT-Prov-Upgrade then. Rationale: faster time-to-signal, avoids premature compliance claims, incremental hardening.

### D-PLAT-002: Canary & Scheduled Verification Strategy
Adopt dual-schedule approach:
1. Weekly Full Matrix Build ("weekly canary"): Sundays 06:00 UTC via cron. Matrix aligns with release OS/ARCH (linux amd64+arm64, darwin arm64, windows amd64 when enabled). Actions: build provider, run unit tests, compile example projects. No publishing.
2. Daily Integration Tests: Daily 05:00 UTC single ubuntu-latest runner executes provider unit + integration tests and selected examples.

Alerting & Issue Policy:
- Weekly canary failure: create or update issue `Weekly Canary Failure <YYYY-MM-DD>` with labels `platform,canary-failure`.
- Daily integration consecutive failures (2 in a row): create issue `Daily Integration Regressions (2 consecutive)` with same labels.
- Issues reused (no duplicates) if already open; add comment with new failure context.

Rationale:
- Balances cost & coverage: broad weekly cross-arch + lightweight daily drift detection.
- Noise reduction via consecutive-failure threshold on daily job.
- Enables board automation through consistent labeling.

Future Evolution: Expand daily matrix to include darwin amd64/arm64 after 30 stable days; incorporate provenance/SBOM verification steps post related decisions.

### D-PLAT-003: Changelog Generation Strategy (GitHub Auto Release Notes)
Adopt GitHub auto-generated release notes for each semantic version tag. The release workflow creates (or updates) a GitHub Release with `generate_release_notes: true`, attaching built artifacts. No persistent `CHANGELOG.md` is committed initially. Rationale: zero additional contributor ceremony, immediate value, aligns with lightweight governance. Limitations: grouping and detection of breaking changes depend on PR titles; future upgrade path to Conventional Commits + structured sections if signal quality proves insufficient. Success Metric: Release body present with sections (Features, Bug Fixes, etc.) for >=90% of tagged releases starting with first post-decision tag.

### D-PLAT-004: SBOM Strategy (CycloneDX Advisory Phase)
Adopt CycloneDX JSON SBOM generation in advisory (non-gating) mode for Go and Node components starting immediately. Two separate SBOM files are produced:
1. `sbom-go.cdx.json` via `cyclonedx-gomod mod -json -licenses -output sbom-go.cdx.json` executed at repository root after Go module tidy.
2. `sbom-node.cdx.json` via `npx @cyclonedx/cyclonedx-npm --output sbom-node.cdx.json` within `sdk/nodejs`.

Both artifacts are attached to the GitHub Release. Generation failures log a warning but do not block publish (advisory mode). Rationale: fast incremental transparency with minimal friction; avoids premature release blocking while tooling reliability is validated. Upgrade Path: evaluate promotion to soft gate (fail only on tool error) after three successful releases with SBOM present; later consider merged aggregate SBOM or SPDX dual format if downstream consumers require.

Success Metrics: SBOM artifacts present for ≥80% of releases in first 60 days; <10% generation error rate.

Remaining Open Questions: None.

## 17. Rollout / Adoption Plan *(mandatory)*
1. Implement FR-001..FR-010 incrementally; start with build + publish + gating.
2. Add checksum/sign & publish idempotency guards.
3. Introduce comment-trigger extended tests.
4. Add lint for SQL safety utilities.
5. Document verification steps in README.
6. Revisit open questions; promote to Accepted once resolved.

## 18. Review & Acceptance Checklist *(auto or manual)*
### Content Quality
- [x] Goals clear & measurable
- [x] Mandatory sections populated
- [x] No unresolved markers

### Requirements Completeness
- [x] FR-/NFR- testable
- [x] Idempotency defined
- [x] All open questions resolved

### Risk & Dependency Readiness
- [x] Critical risks mitigated
- [x] Dependencies identified

## 19. Execution Status *(update as progresses)*
- [x] Drafted
- [x] Scenarios authored
- [x] Requirements finalized
- [x] Review completed
- [x] Accepted

## 20. Appendix *(optional)*
Key Files: `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `provider/conn.go`  
Verification Commands:  
`shasum -a 256 <artifact>`  
`codesign --verify --deep --strict <binary>`  
Utilities: `quoteIdent`, `quoteString`, `ptr.Deref`
