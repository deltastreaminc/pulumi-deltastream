# Data Model: GitHub Actions CI and Release Workflows

## Workflow Entities

### CI Workflow
**Purpose**: Run tests and validations on pull requests
**Properties**:
- Name: `ci`
- Triggers: 
  - Pull request to main branch
  - Manual workflow dispatch (for testing forked PRs)
- Jobs:
  - Setup (prepare environment)
  - Lint (code quality checks)
  - Build (ensure project builds successfully)
  - Test (run test suite)
  - Security (run security checks)
- Conditions:
  - Automatic test execution for PRs from the same repository
  - Manual approval required for PRs from forked repositories
- Outputs:
  - Test results
  - Build logs
  - Status checks on PR

### Release Workflow
**Purpose**: Generate and publish release artifacts
**Properties**:
- Name: `release`
- Triggers:
  - Git tag pushed to main branch
  - Manual workflow dispatch (for testing)
- Jobs:
  - Setup (derive version, prerelease flags)
  - Build (compile provider + generate schema + SDKs per platform matrix)
  - Package (create tarballs, checksums, attach schema artifacts)
  - Test (integration tests using built artifacts)
  - Publish SDKs (Node, Go; optional future languages)
  - Release (GitHub Release creation with schema diff notes)
  - Verify (post-publish smoke tests)
- Platforms:
  - linux x86_64
  - linux arm64
  - darwin arm64
- Outputs:
  - Tarballs: `pulumi-resource-deltastream-v<version>-<os>-<arch>.tar.gz`
  - Checksum file: `pulumi-deltastream_<version>_checksums.txt`
  - Schema artifacts: `schema.json`, optionally `schema-embed.json`
  - Published SDKs (npm, Go module via tag & curated commit)
  - Release notes with schema diff summary
  - Verification job result status

## Secret Entities

### GitHub Secrets
**Purpose**: Store sensitive information securely
**Properties**:
- NPM_TOKEN: Authentication token for publishing to npm
- GITHUB_TOKEN: Authentication token for GitHub operations
- CI_CREDENTIALS_YAML: Credentials for running integration tests
- APPLE_SIGNATURE_IDENTITY: Identity used for code signing macOS binaries
 - (Future optional) PYPI_API_TOKEN, NUGET_PUBLISH_KEY, OSSRH_USERNAME, OSSRH_PASSWORD, JAVA_SIGNING_KEY* for later language enablement

## GitHub Events

### Pull Request Event
**Properties**:
- Event type: `pull_request`
- Actions: `opened`, `synchronize`, `reopened`
- Context: Contains information about the PR author, source repo, etc.

### Push Event
**Properties**:
- Event type: `push`
- Filter: Tags matching `v*` pattern
- Context: Contains information about the tag, commit, etc.

### Workflow Dispatch Event
### Release Verification Event (Indirect)
**Properties**:
- Triggered via `workflow_call` or dispatch from main release workflow
- Inputs: provider version, sdk language inclusion flags, python version (resolved)
- Produces: Pass/fail smoke test status for inclusion in release confidence metrics
**Properties**:
- Event type: `workflow_dispatch`
- Inputs: Optional parameters for manual triggering
- Context: Contains information about the user who triggered the workflow