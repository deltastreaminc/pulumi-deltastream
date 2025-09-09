# Quickstart Guide: GitHub Actions CI and Release Workflows

This guide provides a quick overview of how to use the CI and Release workflows for the Pulumi Deltastream provider.

## CI Workflow

### Required Secrets

The CI workflow requires the following secrets to be set up in your GitHub repository:

- `CI_CREDENTIALS_YAML`: YAML file containing credentials for integration tests

### For Repository Members

1. **When creating a pull request**:
   - The CI workflow will automatically run when you open a PR against the main branch
   - Build steps will run for all PRs to verify your changes build successfully
   - Test steps will only run automatically for PRs from the same repository
   - Results will be posted as status checks on your PR

2. **When reviewing a PR from a fork**:
   - CI will not automatically run tests that require secrets
   - To run tests with secrets:
     - Go to the "Actions" tab in GitHub
     - Select the "CI" workflow
     - Click "Run workflow"
     - Enter the PR number in the input field
     - Click "Run workflow"

3. **When tests fail**:
   - Review the workflow logs to identify the issue
   - Make necessary changes to your branch
   - Push the changes to update the PR
   - CI will automatically run again

## Release Workflow

### Required Secrets

The Release workflow requires the following secrets to be set up in your GitHub repository:

- `NPM_TOKEN`: Token for publishing to npm registry
- `CI_CREDENTIALS_YAML`: YAML file containing credentials for integration tests
- `APPLE_SIGNATURE_IDENTITY`: Identity used for code signing macOS binaries

### Creating a New Release

1. **Tag the commit**:
   ```bash
   # Checkout the main branch
   git checkout main
   git pull

   # Create and push a tag
   git tag v1.2.3
   git push origin v1.2.3
   ```

2. **Monitor the release process**:
   - Go to the "Actions" tab in GitHub
   - Select the "Release" workflow
   - You'll see the workflow running for the tag you pushed
   - Wait for all jobs to complete

3. **Verify the release**:
   - Check that the package has been published to the npm registry:
     ```bash
     yarn info @deltastream/pulumi-deltastream@1.2.3
     ```
   - Verify the GitHub release has been created with artifacts
   - Confirm the git tag is accessible for Go modules

### Manually Triggering a Release

1. **Trigger the workflow**:
   - Go to the "Actions" tab in GitHub
   - Select the "Release" workflow
   - Click "Run workflow"
   - Enter the version in the input field (e.g., v1.2.3)
   - Click "Run workflow"

2. **Monitor and verify the release** as described above

## Troubleshooting

### CI Issues

- **Workflow fails to start**: Ensure the workflow file exists at `.github/workflows/ci.yml`
- **Tests fail due to missing secrets**: Verify the PR is from the same repository or use manual triggering
- **Pulumi command not found**: Ensure the Pulumi CLI action is correctly configured
- **Go build errors**: Verify you're using Go 1.24.x
- **Node modules not found**: Make sure Yarn 1.22.22+ is installed and node_modules are properly cached

### Release Issues

- **yarn publishing fails**: Verify the NPM_TOKEN secret is correctly set in the repository settings
- **Artifact generation fails**: Check build logs for specific errors related to the failing platform
- **Manual release doesn't start**: Ensure you have proper permissions to trigger workflows
- **macOS code signing fails**: Verify the Apple signing secrets are correctly set in repository settings

## Security Notes

- Never commit secrets to the repository
- For forked repositories, sensitive tests are skipped by default
- Repository secrets are managed in the GitHub repository settings