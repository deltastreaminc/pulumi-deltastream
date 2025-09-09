# Research: GitHub Actions CI and Release Workflows

## Research Topics

### GitHub Actions Security Best Practices for Forked Repositories

**Decision**: Use conditional job execution with `if: github.event.pull_request.head.repo.full_name == github.repository` to restrict secret access on forks, and implement manual approval workflows with `workflow_dispatch` events.

**Rationale**: GitHub's recommended approach for handling secrets in forked repositories is to prevent automatic execution of workflows that require secrets. By conditionally running jobs only when the PR is from the same repository or using manual approval triggers, we can prevent secrets from being exposed to potentially malicious forked repositories.

**Alternatives considered**: 
- Environment protection rules: More complex to set up, but provides additional safeguards
- GitHub Apps for authentication: Provides more granular permissions but increases complexity
- Self-hosted runners: Offers more control but increases maintenance burden

### Artifact Generation for Multiple Platforms

**Decision**: Use GitHub Actions matrix strategy with cross-platform runners (ubuntu-latest for Linux, macos-latest for Darwin) and architecture-specific build commands.

**Rationale**: Matrix builds allow parallel execution across different OS environments, and architecture-specific build flags can be passed to the build tools to generate artifacts for different architectures.

**Alternatives considered**:
- Docker cross-compilation: More complex setup but allows building all architectures on a single runner
- Self-hosted runners for specific architectures: More control but increases maintenance
- Third-party actions for cross-compilation: Simpler but introduces external dependencies

### Node.js Package Publishing

**Decision**: Use yarn with NPM registry tokens for publishing to @deltastream/pulumi-deltastream.

**Rationale**: Yarn provides a more consistent dependency management experience while still using NPM registry tokens for authentication. The tokens provide the necessary permissions to publish packages without exposing full NPM account credentials.

**Alternatives considered**:
- GitHub Packages: Tighter integration with GitHub but more complex for consumers
- Manual publishing: More control but defeats automation purpose
- Custom publishing script: More flexible but increases complexity

### Multi-Language Pulumi Provider Publishing (Archived vs Current vs Upstream Reference)

**Context**: We compared three sources:
- Archived reusable workflow (`main-archive/.github/workflows/publish.yml`): Two-phase reusable `workflow_call`, artifacts already packaged as versioned tarballs, checksum + schema diff generation, multi-language SDK publication (Node, Go; hooks for Java, .NET, Python) via `pulumi-package-publisher` and `publish-go-sdk-action`.
- Current `release.yml`: Single linear workflow; builds raw binaries, publishes only Node.js SDK, uploads raw `bin/**` to GitHub Release, no tarballs, no checksums, no schema diff, no Python/Go commit action.
- Upstream `pulumi/pulumi-aws` publish workflow: Canonical reference with provider tarballs, checksum file, schema diff summary embedded in release notes, multi-language SDK publish (nodejs, python, dotnet, go, java), post-publish verification and docs build dispatch.

**Decisions**:
- Reintroduce provider tarball packaging per OS/ARCH: `pulumi-resource-deltastream-v<version>-<os>-<arch>.tar.gz` containing binary + `schema.json` + licensing artifacts.
- Generate and attach a consolidated SHA256 checksum file: `pulumi-deltastream_<version>_checksums.txt`.
- Attach `schema.json` and (optionally) `schema-embed.json` explicitly to the release and produce a schema diff summary with `schema-tools compare` for release notes.
- Add multi-language SDK publication steps: immediate support for Node.js, Go (with curated commit via `publish-go-sdk-action`), Python (wheel + sdist via `pulumi-package-publisher` or manual build). Leave .NET/Java optional but design tasks to allow easy enablement.
- Introduce `isPrerelease` / draft handling to align with upstream pattern and allow safe preview releases.
- Add a verification (smoke test) workflow invocation post-publish to validate installation across languages (mirrors `verify_release` job in aws provider).

**Rationale**: Aligns with established Pulumi provider ecosystem patterns, improves consumer trust (integrity via checksums, visibility via schema diff), and simplifies future expansion to additional SDK languages without redesign.

**Alternatives considered**:
- Continue raw binary upload: Simpler but inconsistent with Pulumi tooling expectations; harder to integrate with automated installers.
- Single giant artifact: Reduces artifact count but complicates distribution to end users and checksum verification granularity.
- Skip schema diff: Faster, but loses automated change visibility and risk classification.

### Python SDK Publishing (Deferred)

**Decision**: Defer Python packaging/publishing for initial release scope; keep directory but exclude from automated pipeline.

**Rationale**: Focus effort on stabilizing provider, Node.js, and Go flows first; reduces complexity and secret surface until Python demand materializes.

**Alternatives considered**:
- Immediate enablement (adds maintenance overhead now)
- Removing directory entirely (loses future scaffold; harder to reintroduce)

### Checksums & Integrity

**Decision**: Use SHA256 sums over all tarballs; store in single text file, commit nothing back to repo (only release assets).

**Rationale**: Standard practice; matches upstream; easy for users to verify downloads.

**Alternatives considered**: GPG signatures (stronger assurance but requires key management) — deferred until external distribution channel demands it.

### Schema Diff Automation

**Decision**: Use `schema-tools compare` (as in archived & aws workflows) to append diff summary to release notes for stable releases; optional for prereleases.

**Rationale**: Early detection of unintended breaking changes; communicates surface changes clearly.

**Alternatives considered**: Manual changelog curation only (higher effort, risk of omissions); custom diff script (reinventing tooling already maintained upstream).

### Verification Job Post-Publish

**Decision**: Add separate `verify_release` (or similarly named) reusable workflow invocation after SDK publish to run minimal Pulumi programs in each language (Node, Go, Python) referencing the just-published version.

**Rationale**: Catches packaging/version mismatches quickly; mirrors reliability practices of major Pulumi providers.

**Alternatives considered**: Inline smoke tests inside publish job (less isolation, harder to rerun independently) or no verification (risk of latent release defects).

### Go Release Process

**Decision**: Create git tags for Go releases without additional publishing steps, as Go modules are consumed directly from GitHub repositories.

**Rationale**: Go's module system relies on version control tags to identify module versions. Creating a git tag is sufficient for Go consumers to reference a specific version.

**Alternatives considered**:
- Proxy services: Unnecessary complexity for standard Go modules
- Binary distribution: Not needed as Go modules are typically imported as source
- Custom Go package registry: Adds complexity without significant benefits

### Workflow Trigger Configuration

**Decision**: Use event-based triggers with conditional logic:
- `pull_request` event for CI workflow
- `push` event with tag filter for Release workflow
- `workflow_dispatch` for manual triggering of test workflows for forked PRs

**Rationale**: This configuration provides automatic testing for internal PRs, automatic releases when tags are pushed, and manual triggering capability for repository maintainers to test forked PRs.

**Alternatives considered**:
- Scheduled workflows: Not appropriate for PR testing or releases
- API-triggered workflows: Adds unnecessary complexity
- Complex event filters: May increase maintenance burden without clear benefits