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