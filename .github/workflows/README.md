# Required GitHub Secrets for CI/Release Workflows

The following secrets need to be set in your GitHub repository settings:

- `CI_CREDENTIALS_YAML`: YAML file containing credentials for integration tests
- `NPM_TOKEN`: Token for publishing to npm registry
- `APPLE_DEVELOPER_CERTIFICATE_P12_BASE64`: Base64-encoded Apple developer certificate
- `APPLE_DEVELOPER_CERTIFICATE_PASSWORD`: Password for the Apple developer certificate
- `APPLE_SIGNATURE_IDENTITY`: Identity used for code signing macOS binaries

## Important Notes

1. These warnings in the workflow files are expected:
   ```
   Context access might be invalid: CI_CREDENTIALS_YAML
   Context access might be invalid: NPM_TOKEN
   Context access might be invalid: APPLE_DEVELOPER_CERTIFICATE_P12_BASE64
   Context access might be invalid: APPLE_DEVELOPER_CERTIFICATE_PASSWORD
   Context access might be invalid: APPLE_SIGNATURE_IDENTITY
   ```
   These warnings occur because GitHub Actions cannot validate secret references during linting.

2. Security has been implemented to prevent leaking secrets to forked repositories:
   - CI workflow uses conditional execution based on PR source
   - Workflows with secrets only run automatically for PRs from the same repository
   - Manual approval via workflow_dispatch is required for PRs from forked repositories