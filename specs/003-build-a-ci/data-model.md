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
  - Setup (prepare environment)
  - Build (generate artifacts for all platforms)
  - Test (verify artifacts work correctly)
  - Publish (publish artifacts to appropriate channels)
- Platforms:
  - linux x86_64
  - linux arm64
  - darwin arm64
- Outputs:
  - Published package to npm registry via yarn (@deltastream/pulumi-deltastream)
  - Git tag for Go releases
  - Release notes

## Secret Entities

### GitHub Secrets
**Purpose**: Store sensitive information securely
**Properties**:
- NPM_TOKEN: Authentication token for publishing to npm
- GITHUB_TOKEN: Authentication token for GitHub operations
- CI_CREDENTIALS_YAML: Credentials for running integration tests
- APPLE_DEVELOPER_CERTIFICATE_P12_BASE64: Base64-encoded Apple developer certificate for code signing
- APPLE_DEVELOPER_CERTIFICATE_PASSWORD: Password for the Apple developer certificate
- APPLE_SIGNATURE_IDENTITY: Identity used for code signing macOS binaries

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
**Properties**:
- Event type: `workflow_dispatch`
- Inputs: Optional parameters for manual triggering
- Context: Contains information about the user who triggered the workflow